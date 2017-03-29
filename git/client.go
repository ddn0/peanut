package git

var defaultClientOpt = &ClientOpt{
	GitPath: "git",
}

// A Client represents a user of git.
type Client struct {
	gitPath string
}

type ClientOpt struct {
	// Path to git program
	GitPath string
}

// NewClient creates a new git client.
func NewClient(opt *ClientOpt) *Client {
	if opt == nil {
		opt = defaultClientOpt
	}

	return &Client{
		gitPath: opt.GitPath,
	}
}
