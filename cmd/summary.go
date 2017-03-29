package cmd

import (
	"fmt"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/ddn0/peanut/git"
	"github.com/dustin/go-humanize"
	"github.com/mattn/go-colorable"
	"github.com/mgutz/ansi"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "show brief status for working directories",
	RunE:  runSummary,
}

func prettySummary(status []Status) error {
	out := colorable.NewColorableStdout()
	printStatus := func(status []Status, color string) {
		for _, s := range status {
			last := git.Log{
				Commit: "None",
			}

			if len(s.LastN) >= 0 {
				last = s.LastN[0]
			}

			sha := last.Commit
			if len(sha) > 7 {
				sha = sha[:7]
			}
			subject := strings.TrimSpace(last.Subject)
			htime := humanize.Time(last.AuthorDate)
			dir, fn := path.Split(s.Repo)
			srepo := path.Join(path.Base(dir), fn)
			fmt.Fprintf(out, "    %s [%s] %s (%s)\n",
				ansi.Color(sha, color),
				ansi.Color(srepo, "cyan"),
				subject,
				htime)
		}
	}

	var master []Status
	var dirty []Status
	var other []Status

	for _, s := range status {
		switch {
		case s.Dirty:
			dirty = append(dirty, s)
		case len(s.Unmerged) > 0:
			dirty = append(dirty, s)
		case len(s.Unpushed) > 0:
			other = append(other, s)
		case s.Commit.Branch == "master":
			master = append(master, s)
		case s.Commit.Branch != "master":
			other = append(other, s)
		default:
			// Shouldn't happen...
			other = append(other, s)
		}
	}

	if len(master) > 0 {
		fmt.Fprintf(out, "on master branch and up-to-date\n")
		printStatus(master, "green")
	}
	if len(other) > 0 {
		fmt.Fprintf(out, "on another branch or unpushed\n")
		printStatus(other, "yellow")
	}
	if len(dirty) > 0 {
		fmt.Fprintf(out, "dirty or out of date\n")
		printStatus(dirty, "red")
	}

	return nil
}

func runSummary(cmd *cobra.Command, args []string) error {
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return err
	}

	cfg, err := readConf()
	if err != nil {
		return err
	}

	seen := make(map[string]bool)
	var status []Status
	for _, dir := range cfg.RepoPaths() {
		s, err := newStatus(dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warn: error reading git status of %q: %s\n", dir, err)
			continue
		}
		if seen[s.Repo] {
			continue
		}
		seen[s.Repo] = true

		status = append(status, *s)
	}

	sort.Sort(StatusSlice(status))

	return prettySummary(status)
}

func init() {
	c := summaryCmd

	RootCmd.AddCommand(c)
}
