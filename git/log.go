package git

import (
	"bytes"
	"fmt"
	"runtime/debug"
	"strconv"
	"time"
)

// A Log is the data associated with a commit.
type Log struct {
	Commit        string
	Tree          string
	Parents       []string
	Author        string
	AuthorDate    time.Time
	Committer     string
	CommitterDate time.Time
	Subject       string
	Body          string
}

func mustConsume(bs []byte, target []byte) []byte {
	next := consume(bs, target)
	if len(next) == len(bs) {
		panic("could not find target")
	}
	return next
}

func consume(bs []byte, target []byte) []byte {
	if bytes.HasPrefix(bs, target) {
		return bs[len(target):]
	}
	return bs
}

func parseLine(bs []byte) ([]byte, []byte) {
	idx := bytes.IndexByte(bs, '\n')
	if idx < 0 {
		panic("could not parse line")
	}
	return bs[:idx], bs[idx+1:]
}

func unmarshalSha(bs []byte, l *Log) []byte {
	c, rest := parseLine(bs)
	l.Commit = string(c)
	return rest
}

func unmarshalTree(bs []byte, l *Log) []byte {
	bs = mustConsume(bs, []byte("tree "))
	t, rest := parseLine(bs)
	l.Tree = string(t)
	return rest
}

func unmarshalParent(bs []byte, l *Log) []byte {
	bs0 := consume(bs, []byte("parent "))
	if len(bs0) == len(bs) {
		return bs
	}

	p, rest := parseLine(bs0)
	l.Parents = append(l.Parents, string(p))
	return rest
}

func unmarshalUserTime(bs []byte, user *string, t *time.Time) []byte {
	eol := bytes.IndexByte(bs, '\n')
	if eol < 0 {
		panic("could not parse user time")
	}
	zone := bytes.LastIndexByte(bs[:eol], ' ')
	if zone < 0 {
		panic("could not parse user time")
	}
	ts := bytes.LastIndexByte(bs[:zone], ' ')
	if ts < 0 {
		panic("could not parse user time")
	}
	i, err := strconv.ParseInt(string(bs[ts+1:zone]), 10, 64)
	if err != nil {
		panic(err)
	}

	*t = time.Unix(i, 0)
	*user = string(bs[:ts])

	return bs[eol+1:]
}

func unmarshalAuthor(bs []byte, l *Log) []byte {
	bs = mustConsume(bs, []byte("author "))

	return unmarshalUserTime(bs, &l.Author, &l.AuthorDate)
}

func unmarshalCommitter(bs []byte, l *Log) []byte {
	bs = mustConsume(bs, []byte("committer "))

	return unmarshalUserTime(bs, &l.Committer, &l.CommitterDate)
}

func unmarshalSubject(bs []byte, l *Log) []byte {
	bs0 := consume(bs, []byte("\n"))
	if len(bs0) == len(bs) {
		return bs
	}
	s, rest := parseLine(bs0)
	l.Subject = string(s)
	return rest
}

func unmarshalBody(bs []byte, l *Log) []byte {
	bs0 := consume(bs, []byte("\n"))
	if len(bs0) == len(bs) {
		return bs
	}
	l.Body = string(bs0)
	return nil
}

// Raw log format:
//   <commit sha>
//   tree <sha>
//   parent <sha>
//   parent <sha>
//   author <username> <email> <timestamp> <timezone>
//   committer <username> <email> <timestamp> <timezone>
//   <empty line>
//   <4 spaces><subject>
//   <empty line>
//   <4 spaces><body>
func unmarshalLog(bs []byte) (log *Log, err error) {
	log = &Log{}

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("parse error: %s from %s", r, debug.Stack())
		}
	}()

	bs = unmarshalSha(bs, log)
	bs = unmarshalTree(bs, log)

	for {
		bs0 := unmarshalParent(bs, log)
		if len(bs0) == len(bs) {
			break
		}
		bs = bs0
	}

	bs = unmarshalAuthor(bs, log)
	bs = unmarshalCommitter(bs, log)
	bs = unmarshalSubject(bs, log)
	bs = unmarshalBody(bs, log)

	return
}

// RevListNot returns the commit string representing the negation of the commit
// set for RevList.
func RevListNot(commit string) string {
	return fmt.Sprintf("^%s", commit)
}

// FirstParent returns the commit string representing the first parent of the
// commit.
func FirstParent(commit string) string {
	return fmt.Sprintf("%s^", commit)
}

// Logs returns the logs for a set of commits. See git rev-list for syntax of
// commits.
func (a *Client) Logs(repo string, commits ...string) ([]Log, error) {
	var args []string
	args = append(args, "rev-list", "--header")
	args = append(args, commits...)
	out, err := output(repo, a.gitPath, args...)
	if err != nil {
		return nil, err
	}
	objs := bytes.Split(out, []byte{'\x00'})

	var logs []Log
	for _, obj := range objs {
		if len(obj) == 0 {
			continue
		}
		l, err := unmarshalLog(obj)
		if err != nil {
			return nil, err
		}
		logs = append(logs, *l)
	}
	return logs, nil
}
