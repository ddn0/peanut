package cmd

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	errNotFound = errors.New("not found")
)

var wdCmd = &cobra.Command{
	Use:   "wd [package]",
	Short: "print working directory of package",
	RunE:  runWd,
}

func runWd(cmd *cobra.Command, args []string) error {
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return err
	}

	cfg, err := readConf()
	if err != nil {
		return err
	}

	repos := cfg.RepoPaths()
	score := make(map[string]int)
	for _, dir := range repos {
		for _, arg := range args {
			if strings.Index(dir, arg) >= 0 {
				score[dir] += 1
			}
		}
		for len(dir) > 0 {
			b := filepath.Base(dir)
			for _, arg := range args {
				if b == arg {
					score[dir] += len(arg) + 1
				}
			}
			next := filepath.Dir(dir)
			if dir == next {
				break
			}
			dir = next
		}
	}

	sort.Slice(repos, func(i, j int) bool {
		return score[repos[i]] > score[repos[j]]
	})

	for _, r := range repos {
		fmt.Println(r)
		return nil
	}

	return errNotFound
}

func init() {
	c := wdCmd

	RootCmd.AddCommand(c)
}
