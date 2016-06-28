package main

import (
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
