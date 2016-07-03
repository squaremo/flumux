package main

import (
	"fmt"

	"github.com/spf13/cobra"
	git "gopkg.in/libgit2/git2go.v24"
)

type latestOpts struct {
	registryOpts
	gitOpts
}

func addLatestCommand(top *cobra.Command) {
	opts := &latestOpts{}
	cmd := &cobra.Command{
		Use:   "latest",
		Short: "output the name of the latest image, relative to the git revision given",
		RunE:  opts.run,
	}
	opts.addGitFlags(cmd)
	opts.addRegistryFlags(cmd)
	top.AddCommand(cmd)
}

func (opts *latestOpts) run(_ *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("expected argument <image>")
	}
	image := args[0]

	repo, err := opts.openRepository()
	if err != nil {
		return err
	}

	walk, err := repo.Walk()
	if err != nil {
		return err
	}
	err = walk.PushHead()
	if err != nil {
		return err
	}

	walk.Sorting(git.SortTopological)

	regClient, err := opts.newRegistryClient(image)
	if err != nil {
		return err
	}

	tags, err := regClient.Repository.ListTags(image, regClient.auth)
	if err != nil {
		return err
	}

	done := false
	repo.iterateImages(walk, tags, func(tag string, _ *git.Commit) bool {
		done = true
		fmt.Println(imageName(image, tag))
		return false
	})

	if !done {
		return fmt.Errorf("no result (no images, or no images that correspond to a commit)")
	}
	return nil
}
