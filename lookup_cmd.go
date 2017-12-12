package main

import (
	"context"
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
	imageStr := args[0]
	image, err := parseImage(imageStr)
	if err != nil {
		return err
	}

	repo, err := opts.openRepository()
	if err != nil {
		return err
	}

	regClient, err := opts.newRegistryClient(image)
	if err != nil {
		return err
	}

	ctx := context.Background()
	tags, err := regClient.Tags(ctx)
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
