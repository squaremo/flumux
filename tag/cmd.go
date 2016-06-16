package tag

import (
	//	"fmt"
	"os"
	"text/template"

	"github.com/spf13/cobra"
	git "gopkg.in/libgit2/git2go.v22"
)

type tagOpts struct {
	template string
	ref      string
	repo     string
}

type repoState struct {
	Commit string
}

func (opts *tagOpts) run(_ *cobra.Command, args []string) error {
	tmpl, err := template.New("tag").Parse(opts.template)
	if err != nil {
		return err
	}
	repo, err := git.OpenRepository(opts.repo)
	if err != nil {
		return err
	}
	ref, err := repo.LookupReference(opts.ref)
	if err != nil {
		return err
	}
	ref, err = ref.Resolve()
	if err != nil {
		return err
	}
	return tmpl.Execute(os.Stdout, repoState{
		Commit: ref.Target().String(),
	})
}

func AddSubcommandTo(cmd *cobra.Command) {
	opts := tagOpts{}
	subcmd := &cobra.Command{
		Use:   "tag",
		Short: "generate an image tag from git state",
		RunE:  opts.run,
	}
	subcmd.Flags().StringVarP(&opts.template, "template", "t",
		"{{.Commit}}", "template for generating a tag")
	subcmd.Flags().StringVar(&opts.ref, "ref", "HEAD",
		`git ref to generate tag for; e.g., "master"`)
	subcmd.Flags().StringVarP(&opts.repo, "repository", "r", ".",
		"git repository, as a path")
	cmd.AddCommand(subcmd)
}
