package main

import (
	"fmt"

	"github.com/spf13/cobra"
	git "gopkg.in/libgit2/git2go.v24"
)

type closestOpts struct {
	gitOpts
	registryOpts
	head     string
	ancestor string
}

func addClosestCommand(top *cobra.Command) {
	opts := closestOpts{}
	cmd := &cobra.Command{
		Use:   "closest <image> <ref>",
		Short: "give the oldest version of <image> that is after <ref>",
		RunE:  opts.run,
	}
	opts.addGitFlags(cmd)
	opts.addRegistryFlags(cmd)
	cmd.Flags().StringVar(&opts.head, "head", "master", "treat this ref as the git repo head")
	cmd.Flags().StringVar(&opts.ancestor, "ancestor", "", "stop at this ref")
	top.AddCommand(cmd)
}

func (opts *closestOpts) run(_ *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("expected argument <image>")
	}
	image := args[0]

	repo, err := opts.openRepository()
	if err != nil {
		return err
	}

	headObj, err := repo.RevparseSingle(opts.head)
	if err != nil {
		return err
	}
	var refObj *git.Object
	if opts.ancestor != "" {
		refObj, err = repo.RevparseSingle(opts.ancestor)
		if err != nil {
			return err
		}
	}

	regClient, err := opts.newRegistryClient(image)
	if err != nil {
		return err
	}

	tags, err := regClient.Repository.ListTags(image, regClient.auth)
	if err != nil {
		return err
	}

	walk, err := repo.Walk()
	if err != nil {
		return err
	}
	walk.Sorting(git.SortTopological)

	if err := walk.Push(headObj.Id()); err != nil {
		return err
	}
	if refObj != nil {
		if err := walk.Hide(refObj.Id()); err != nil {
			return err
		}
	}

	repo.iterateImages(walk, tags, func(tag string, commit *git.Commit) bool {
		fmt.Println(imageName(image, tag))
		return true
	})
	return nil
}
