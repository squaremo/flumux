package list

import (
	"github.com/spf13/cobra"
)

type listOpts struct {
}

func (opts *listOpts) run(_ *cobra.Command, args []string) error {
	return nil
}

func AddSubcommandTo(cmd *cobra.Command) {
	opts := listOpts{}
	subcmd := &cobra.Command{
		Use:   "list",
		Short: "list images",
		RunE:  opts.run,
	}
	cmd.AddCommand(subcmd)
}
