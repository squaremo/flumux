package main

import (
	"fmt"
	"context"

	"github.com/spf13/cobra"
	git "gopkg.in/libgit2/git2go.v24"
)

type latestOpts struct {
	registryOpts
	gitOpts
	head string
}

func addLatestCommand(top *cobra.Command) {
	opts := &latestOpts{}
	cmd := &cobra.Command{
		Use:   "latest",
		Short: "output the name of the latest image relative to a git ref",
		RunE:  opts.run,
	}
	opts.addGitFlags(cmd)
	opts.addRegistryFlags(cmd)
	cmd.Flags().StringVar(&opts.head, "ref", "master", "look for the latest image relative to this ref")
	top.AddCommand(cmd)
}

func (opts *latestOpts) run(_ *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("expected argument <image>")
	}
	imageStr := args[0]
	image, err := parseImage(imageStr)
	if err != nil {
		return err
	}

	repo, err := opts.openRepository()
	if err != nil {
		return err
	}

	head, err := repo.References.Dwim(opts.head)
	if err != nil {
		return err
	}
	headObj, err := head.Resolve()
	if err != nil {
		return err
	}

	walk, err := repo.Walk()
	if err != nil {
		return err
	}
	err = walk.Push(headObj.Target())
	if err != nil {
		return err
	}

	walk.Sorting(git.SortTopological)

	regClient, err := opts.newRegistryClient(image)
	if err != nil {
		return err
	}

	ctx := context.Background()
	tags, err := regClient.Tags(ctx)
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
