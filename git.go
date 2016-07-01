package main

import (
	"sort"
	"strings"

	"github.com/spf13/cobra"
	git "gopkg.in/libgit2/git2go.v24"
)

type gitOpts struct {
	gitrepo string
}

func (opts *gitOpts) addGitFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&opts.gitrepo, "git-repo-path", "C", ".",
		"path to git repository")
}

func (opts *gitOpts) openRepository() (*git.Repository, error) {
	return git.OpenRepository(opts.gitrepo)
}

func commitFromTag(repo *git.Repository, tag string) (*git.Commit, string) {
	bits := strings.Split(tag, "-")
	if len(bits) != 2 {
		return nil, "tag does not correspond to a commit"
	}
	commitRev, err := repo.RevparseSingle(bits[1])
	if err != nil {
		return nil, err.Error()
	}
	commit, err := commitRev.AsCommit()
	if err != nil {
		return nil, err.Error()
	}
	return commit, ""
}

func makeImageHistory(tags map[string]string, repo *git.Repository) []*image {
	result := &imageList{repo, make([]*image, len(tags))}
	i := 0
	for tag, _ := range tags {
		entry := &image{tag: tag}
		result.entries[i] = entry
		i++

		additional := ""
		// hard-code tag format for now
		if strings.HasSuffix(tag, "-WIP") {
			additional = " (uncommitted changes)"
			tag = tag[:len(tag)-4]
		}
		commit, otherwise := commitFromTag(repo, tag)
		if otherwise != "" {
			entry.msg = otherwise
		} else {
			entry.commitID = commit.Id()
			entry.msg = strings.Split(commit.Message(), "\n")[0]
		}
		entry.msg = entry.msg + additional
	}
	sort.Sort(result)
	return result.entries
}

type image struct {
	tag      string
	msg      string
	commitID *git.Oid
}

type imageList struct {
	repo    *git.Repository
	entries []*image
}

func (result *imageList) Less(i, j int) bool {
	if result.entries[i].commitID == nil {
		return false
	} else if result.entries[j].commitID == nil {
		return true
	}
	// Define: A < B iff A is a descendant of B, i.e., comes after it in git history
	res, err := result.repo.DescendantOf(result.entries[i].commitID, result.entries[j].commitID)
	// assume an error indicates no relative ordering
	return (err == nil) && res
}

func (result *imageList) Swap(i, j int) {
	t := result.entries[i]
	result.entries[i] = result.entries[j]
	result.entries[j] = t
}

func (result *imageList) Len() int {
	return len(result.entries)
}
