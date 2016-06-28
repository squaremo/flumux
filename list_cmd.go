package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	git "gopkg.in/libgit2/git2go.v24"
)

type listOpts struct {
	gitOpts
	registryOpts
}

type resultEntry struct {
	tag      string
	msg      string
	commitID *git.Oid
}

type result struct {
	repo    *git.Repository
	entries []*resultEntry
}

func (result *result) Less(i, j int) bool {
	if result.entries[i].commitID == nil {
		return false
	} else if result.entries[j].commitID == nil {
		return true
	}
	// Define: A < B iff A is a descendant of B, i.e., comes after it in git history
	res, err := result.repo.DescendantOf(result.entries[i].commitID, result.entries[j].commitID)
	// assume an error indicates no relative ordering
	return (err == nil) && res
}

func (result *result) Swap(i, j int) {
	t := result.entries[i]
	result.entries[i] = result.entries[j]
	result.entries[j] = t
}

func (result *result) Len() int {
	return len(result.entries)
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

	out := tabwriter.NewWriter(os.Stdout, 0, 4, 1, ' ', 0)
	for _, entry := range result.entries {
		fmt.Fprint(out, entry.tag)
		fmt.Fprint(out, "\t")
		fmt.Fprint(out, entry.msg)
		fmt.Fprint(out, "\n")
	}
	out.Flush()

	return nil
}

func commitFromTag(repo *git.Repository, tag string) (*git.Commit, string) {
	bits := strings.Split(tag, "-")
	if len(bits) != 2 {
		return nil, "tag does not correspond to a commit"
	}
	commitRev, err := repo.RevparseSingle(bits[1])
	if err != nil {
		return nil, err.Error()
	}
	commit, err := commitRev.AsCommit()
	if err != nil {
		return nil, err.Error()
	}
	return commit, ""
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
