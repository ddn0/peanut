package git

import (
	"bytes"
	"fmt"
	"os/exec"
)

// A Commit represents a git commit.
type Commit struct {
	Sha      string // Commit sha
	Branch   string // Branch name
	Upstream string // Upstream branch name
	Repo     string
	client   *Client
}

func output(dir, prog string, args ...string) ([]byte, error) {
	cmd := exec.Command(prog, args...)
	cmd.Dir = dir
	return cmd.Output()
}

// Head returns the git HEAD commit
func (a *Client) Head(repo string) (*Commit, error) {
	branch, err := output(repo, a.gitPath, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return nil, err
	}

	branchStr := string(bytes.TrimSpace(branch))
	sha, err := output(repo, a.gitPath, "rev-parse", branchStr)
	if err != nil {
		return nil, err
	}

	upstream, _ := output(repo, a.gitPath, "rev-parse", "--abbrev-ref", fmt.Sprintf("%s@{upstream}", branchStr))

	return &Commit{
		Sha:      string(bytes.TrimSpace(sha)),
		Branch:   branchStr,
		Upstream: string(bytes.TrimSpace(upstream)),
		Repo:     repo,
		client:   a,
	}, nil
}

// UpstreamMerge returns the merge of this commit with its upstream branch.
func (a *Commit) UpstreamMerge() (*Merge, error) {
	if len(a.Upstream) == 0 {
		return nil, fmt.Errorf("no upstream branch")
	}

	upstreamSha, err := output(a.Repo, a.client.gitPath, "rev-parse", a.Upstream)
	if err != nil {
		return nil, err
	}

	upstreamShaStr := string(bytes.TrimSpace(upstreamSha))
	mergeSha, err := output(a.Repo, a.client.gitPath, "merge-base", a.Sha, upstreamShaStr)
	if err != nil {
		return nil, err
	}

	base := &Commit{
		Sha:  string(bytes.TrimSpace(mergeSha)),
		Repo: a.Repo,
	}
	topic := &Commit{
		Sha:    upstreamShaStr,
		Branch: a.Upstream,
		Repo:   a.Repo,
	}

	return &Merge{
		Base:    base,
		Topic:   topic,
		Current: a,
	}, nil
}
