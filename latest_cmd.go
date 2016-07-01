package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	git "gopkg.in/libgit2/git2go.v24"
)

type latestOpts struct {
	registryOpts
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
	opts.addRegistryFlags(cmd)
	top.AddCommand(cmd)
}

func (opts *latestOpts) run(_ *cobra.Command, args []string) error {
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

	result := &result{repo, make([]*resultEntry, len(tags))}
	i := 0
	for tag, _ := range tags {
		entry := &resultEntry{tag: tag}
		result.entries[i] = entry
		i++

		additional := ""
		// hard-code tag format for now
		if strings.HasSuffix(tag, "-WIP") {
			additional = " (uncommitted changes)"
			tag = tag[:len(tag)-4]
		}
		commit, otherwise := commitFromTag(repo, tag)
		if otherwise != "" {
			entry.msg = otherwise
		} else {
			entry.commitID = commit.Id()
			entry.msg = strings.Split(commit.Message(), "\n")[0]
		}
		entry.msg = entry.msg + additional
	}
	sort.Sort(result)

	if result.Len() > 0 {
		fmt.Printf("%s:%s\n", image, result.entries[0].tag)
	} else {
		return fmt.Errorf("no result (no images, or no images that correspond to a commit)")
	}

	return nil
}
