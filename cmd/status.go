package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/ddn0/peanut/git"
	"github.com/dustin/go-humanize"
	"github.com/mattn/go-colorable"
	"github.com/mgutz/ansi"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "show status for working directories",
	RunE:  runStatus,
}

type Status struct {
	Repo             string
	Commit           *git.Commit
	Dirty            bool
	DirtyFiles       []string
	LastN            []git.Log
	Unpushed         []git.Log
	Unmerged         []git.Log
	Missing          []git.Log
	UnmergedBranches []string
}

type StatusSlice []Status

func (a StatusSlice) Len() int {
	return len(a)
}

func (a StatusSlice) Less(i, j int) bool {
	return a[i].Repo < a[j].Repo
}

func (a StatusSlice) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func newStatus(dir string) (*Status, error) {
	gc := git.NewClient(nil)
	wt, err := gc.WorkTree(dir)
	if err != nil {
		return nil, err
	}
	upstream := wt.Commit.Upstream
	if len(upstream) == 0 {
		upstream = "origin/master"
	}
	unpushed, _ := gc.Logs(wt.Repo, wt.Commit.Sha, git.RevListNot(upstream))
	unmerged, _ := gc.Logs(wt.Repo, git.RevListNot(wt.Commit.Sha), upstream)
	var missing []git.Log
	if upstream != "origin/master" {
		missing, _ = gc.Logs(wt.Repo, git.RevListNot(wt.Commit.Sha), "origin/master")
	}
	unmergedB, _ := wt.UnmergedBranches("origin/master")
	last, _ := gc.Logs(wt.Repo, wt.Commit.Sha, git.RevListNot(git.FirstParent(wt.Commit.Sha)))

	return &Status{
		Repo:             wt.Repo,
		Commit:           wt.Commit,
		Dirty:            len(wt.DirtyFiles) > 0,
		DirtyFiles:       wt.DirtyFiles,
		LastN:            last,
		Unpushed:         unpushed,
		Unmerged:         unmerged,
		Missing:          missing,
		UnmergedBranches: unmergedB,
	}, nil
}

func prettyStatus(status []Status) error {
	out := colorable.NewColorableStdout()
	printLog := func(ls []git.Log, heading, color string) {
		if len(ls) == 0 {
			return
		}

		fmt.Fprintf(out, "  %s\n", heading)
		for _, l := range ls {
			sha := l.Commit
			if len(sha) > 7 {
				sha = sha[:7]
			}
			subject := strings.TrimSpace(l.Subject)
			htime := humanize.Time(l.AuthorDate)
			fmt.Fprintf(out, "    %s %s (%s)\n", ansi.Color(sha, color), subject, htime)
		}
	}
	printBranches := func(bs []string, heading, color string) {
		if len(bs) == 0 {
			return
		}

		fmt.Fprintf(out, "  %s\n", heading)
		fmt.Fprintf(out, "    ")
		for _, b := range bs {
			fmt.Fprintf(out, "%s ", ansi.Color(b, color))
		}
		fmt.Fprintf(out, "\n")

	}

	for _, s := range status {
		var branch string
		if s.Commit.Branch != "master" {
			branch = ansi.Color(fmt.Sprintf("(%s)", s.Commit.Branch), "170")
		}
		fmt.Fprintln(out, ansi.Color(s.Repo, "cyan"), branch)
		if s.Dirty {
			fmt.Fprintf(out, "  Dirty:\n")
			for _, f := range s.DirtyFiles {
				fmt.Fprintln(out, "    ", ansi.Color(f, "red"))
			}
		}
		printLog(s.Unmerged, "Unmerged:", "blue")
		printLog(s.Missing, "Missing:", "blue")
		printLog(s.Unpushed, "Unpushed:", "yellow")
		printBranches(s.UnmergedBranches, "Unmerged Branches:", "blue")
	}

	return nil
}

func runStatus(cmd *cobra.Command, args []string) error {
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

	if format := viper.GetString("format"); format == "pretty" {
		return prettyStatus(status)
	} else {
		return print(status, format, viper.GetString("filter"))
	}
}

func init() {
	c := statusCmd
	flags := c.Flags()

	RootCmd.AddCommand(c)
	flags.String("format", "pretty", "Output format {pretty,json,text,yaml}")
	flags.String("filter", "", "Filter text format using go package template")
}
