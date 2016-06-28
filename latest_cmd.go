package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

type latestOpts struct {
	registry string
	gitOpts
}

func addLatestCommand(top *cobra.Command) {
	opts := latestOpts{}
	cmd := &cobra.Command{
		Use:   "latest",
		Short: "output the name of the latest image, relative to the git revision given",
		RunE:  opts.run,
	}
	opts.addGitFlags(cmd)
	top.AddCommand(cmd)
}

func (opts *latestOpts) run(_ *cobra.Command, args []string) error {
	fmt.Println("latest")
	return nil
}
