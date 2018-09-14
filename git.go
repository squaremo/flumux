package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	git "gopkg.in/libgit2/git2go.v26"
)

type gitOpts struct {
	gitrepo string
}

func (opts *gitOpts) addGitFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&opts.gitrepo, "git-repo-path", "C", ".",
		"path to git repository")
}

type repository struct{ *git.Repository }

func (opts *gitOpts) openRepository() (repository, error) {
	repo := repository{}
	var err error
	repo.Repository, err = git.OpenRepository(opts.gitrepo)
	return repo, err
}

func (repo repository) commitFromTag(tag string) (*git.Commit, string) {
	bits := strings.Split(tag, "-")
	if len(bits) != 2 {
		return nil, "tag does not correspond to a commit"
	}
	commitObj, err := repo.RevparseSingle(bits[1])
	if err != nil {
		return nil, err.Error()
	}
	commit, err := commitObj.AsCommit()
	if err != nil {
		return nil, err.Error()
	}
	return commit, ""
}

// take a map of tag -> whatever and return a map of commit ID -> tag,
// discarding any tags that don't correspond to a commit
func (repo repository) tagsToCommits(tags []string) map[string]string {
	commits := make(map[string]string)
	for _, tag := range tags {
		commit, _ := repo.commitFromTag(tag)
		if commit != nil {
			commits[commit.Id().String()] = tag
		} else {
			fmt.Fprintf(os.Stderr, "No commit found for tag %s\n", tag)
		}
	}
	return commits
}

type imageIterator func(string, *git.Commit) bool

func (repo repository) iterateImages(walk *git.RevWalk, tags []string, visit imageIterator) {
	commits := repo.tagsToCommits(tags)
	walk.Iterate(func(commit *git.Commit) bool {
		if tag, found := commits[commit.Id().String()]; found {
			delete(commits, commit.Id().String())
			return visit(tag, commit) && len(commits) > 0
		}
		return true
	})
}
