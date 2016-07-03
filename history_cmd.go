package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	git "gopkg.in/libgit2/git2go.v24"
)

type listOpts struct {
	gitOpts
	registryOpts
}

func (opts *listOpts) run(_ *cobra.Command, args []string) error {
	var image, revRange string
	switch len(args) {
	case 1:
		image = args[0]
	case 2:
		image = args[0]
		revRange = args[1]
	default:
		return fmt.Errorf("expected argument <image> and optionally, <revision range>")
	}

	repo, err := opts.openRepository()
	if err != nil {
		return err
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

	if revRange != "" {
		if err := walk.PushRange(revRange); err != nil {
			return err
		}
	} else {
		if err := walk.PushHead(); err != nil {
			return err
		}
	}

	out := tabwriter.NewWriter(os.Stdout, 0, 4, 1, ' ', 0)

	repo.iterateImages(walk, tags, func(tag string, commit *git.Commit) bool {
		fmt.Fprint(out, imageName(image, tag))
		fmt.Fprint(out, "\t")
		fmt.Fprint(out, commit.Summary())
		fmt.Fprint(out, "\n")
		return true
	})
	out.Flush()
	return nil
}

func addHistoryCommand(cmd *cobra.Command) {
	opts := &listOpts{}
	subcmd := &cobra.Command{
		Use:   "history <image> [<revision range>]",
		Short: "history of images",
		RunE:  opts.run,
	}
	opts.addRegistryFlags(subcmd)
	opts.addGitFlags(subcmd)
	cmd.AddCommand(subcmd)
}
