package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/squaremo/flumux/list"
	"github.com/squaremo/flumux/tag"
)

func main() {
	topCmd := &cobra.Command{
		Use:   "flumux",
		Short: "container platform multitool",
	}
	tag.AddSubcommandTo(topCmd)
	list.AddSubcommandTo(topCmd)

	if err := topCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
