package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

type closestOpts struct {
	gitOpts
	registryOpts
	head string
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
	cmd.Flags().StringVar(&opts.head, "head", "master",
		"a ref of which the image commit must be an ancestor; e.g,. used to specify a branch into which it must have been merged")
	top.AddCommand(cmd)
}

func (opts *closestOpts) run(_ *cobra.Command, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("expected arguments <image> and <ref>")
	}
	image := args[0]
	ref := args[1]
	fmt.Printf("closest %s %s\n", image, ref)
	return nil
}
