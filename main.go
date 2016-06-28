package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	topCmd := &cobra.Command{
		Use:   "flumux",
		Short: "container platform multitool",
	}
	addTagCommand(topCmd)
	addListCommand(topCmd)
	addLatestCommand(topCmd)

	if err := topCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
