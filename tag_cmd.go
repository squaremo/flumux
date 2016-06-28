package main

import (
	"fmt"
	"os"
	"text/template"

	"github.com/spf13/cobra"
)

type tagOpts struct {
	template string
	ref      string
	gitOpts
}

type repoState struct {
	Commit string
}

func (opts *tagOpts) run(_ *cobra.Command, args []string) error {
	tmpl, err := template.New("tag").Parse(opts.template)
	if err != nil {
		return err
	}
	repo, err := opts.openRepository()
	if err != nil {
		return err
	}
	ref, err := repo.References.Dwim(opts.ref)
	if err != nil {
		return err
	}
	ref, err = ref.Resolve()
	if err != nil {
		return err
	}
	err = tmpl.Execute(os.Stdout, repoState{
		Commit: ref.Target().String(),
	})
	fmt.Println("")
	return err
}

func addTagCommand(cmd *cobra.Command) {
	opts := &tagOpts{}
	subcmd := &cobra.Command{
		Use:   "tag",
		Short: "generate an image tag from git state",
		RunE:  opts.run,
	}
	subcmd.Flags().StringVarP(&opts.template, "template", "t",
		"{{.Commit}}", "template for generating a tag")
	subcmd.Flags().StringVar(&opts.ref, "ref", "HEAD",
		`git ref to generate tag for; e.g., "master"`)
	opts.addGitFlags(subcmd)
	cmd.AddCommand(subcmd)
}
