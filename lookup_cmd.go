package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

type lookupOpts struct {
	gitOpts
	registryOpts
}

func addLookupCommand(top *cobra.Command) {
	opts := &lookupOpts{}
	cmd := &cobra.Command{
		Use:   "lookup",
		Short: "find an image for the commit(s) supplied",
		RunE:  opts.run,
	}
	opts.addGitFlags(cmd)
	opts.addRegistryFlags(cmd)
	top.AddCommand(cmd)
}

func (opts *lookupOpts) run(_ *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("expected argument <image>")
	}
	image := args[0]

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
	commits := repo.tagsToCommits(tags)

	lines := bufio.NewScanner(os.Stdin)
	for lines.Scan() {
		line := lines.Text()
		fields := strings.Fields(line)
		if len(fields) < 1 {
			continue
		}
		rev := fields[0]
		commitObj, err := repo.RevparseSingle(rev)
		if err != nil {
			fmt.Fprintf(os.Stderr, "commit '%s' not found\n", rev)
		}
		if tag, found := commits[commitObj.Id().String()]; found {
			fmt.Println(imageName(image, tag), line[len(rev):])
		}
	}
	return nil
}
