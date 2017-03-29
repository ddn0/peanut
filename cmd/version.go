package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "print version",
	RunE:  runVersion,
}

func runVersion(cmd *cobra.Command, args []string) error {
	fmt.Printf("version: %s\n", strings.TrimSpace(version))
	fmt.Printf("commit: %s\n", strings.TrimSpace(commit))
	return nil
}

func init() {
	c := versionCmd
	RootCmd.AddCommand(c)
}
