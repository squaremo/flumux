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
	if len(args) != 1 {
		return fmt.Errorf("expected argument <image>")
	}
	image := args[0]

	var (
		repo *git.Repository
		err  error
	)

	repo, err = opts.openRepository()
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

	images := makeImageHistory(tags, repo)

	out := tabwriter.NewWriter(os.Stdout, 0, 4, 1, ' ', 0)
	for _, entry := range images {
		fmt.Fprint(out, entry.tag)
		fmt.Fprint(out, "\t")
		fmt.Fprint(out, entry.msg)
		fmt.Fprint(out, "\n")
	}
	out.Flush()

	return nil
}

func addListCommand(cmd *cobra.Command) {
	opts := &listOpts{}
	subcmd := &cobra.Command{
		Use:   "list <image>",
		Short: "list images",
		RunE:  opts.run,
	}
	opts.addRegistryFlags(subcmd)
	opts.addGitFlags(subcmd)
	cmd.AddCommand(subcmd)
}
