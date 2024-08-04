package git

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
)

const (
	HookPrepareCommitMessageFile     = "prepare-commit-msg"
	HookPrepareCommitMessageTemplate = `#!/bin/sh

if [[ "$2" != "message" && "$2" != "commit" ]]; then
  ai commit --file $1 --preview --no_confirm
fi
`
)

var excludeFromDiff = []string{
	"package-lock.json",
	"pnpm-lock.yaml",
	"*.lock",
	"go.sum",
}

type Command struct {
	// Generate diffs with <n> lines of context, instead of the usual three
	diffUnified int
	excludeList []string
	isAmend     bool
}

func New(opts ...Option) *Command {
	cfg := &config{}

	for _, o := range opts {
		o.apply(cfg)
	}

	cmd := &Command{
		diffUnified: cfg.diffUnified,
		// Append the user-defined excludeList to the default excludeFromDiff
		excludeList: append(excludeFromDiff, cfg.excludeList...),
		isAmend:     cfg.isAmend,
	}

	return cmd
}

func (c *Command) Commit(val string) (string, error) {
	output, err := c.commit(val).Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

// GitDir to show the (by default, absolute) path of the git directory of the working tree.
func (c *Command) GitDir() (string, error) {
	output, err := c.gitDir().Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

// DiffFiles compares the differences between two sets of data.
func (c *Command) DiffFiles() (string, error) {
	output, err := c.diffNames().Output()
	if err != nil {
		return "", err
	}
	if string(output) == "" {
		return "", errors.New("please add your staged changes using git add <files...>")
	}

	output, err = c.diffFiles().Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

func (c *Command) InstallHook() error {
	hookPath, err := c.hookPath().Output()
	if err != nil {
		return err
	}

	target := path.Join(strings.TrimSpace(string(hookPath)), HookPrepareCommitMessageFile)

	return os.WriteFile(target, []byte(HookPrepareCommitMessageTemplate), 0o600)
}

func (c *Command) UninstallHook() error {
	hookPath, err := c.hookPath().Output()
	if err != nil {
		return err
	}

	target := path.Join(strings.TrimSpace(string(hookPath)), HookPrepareCommitMessageFile)
	if err := os.Remove(target); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return errors.New("hook file prepare-commit-msg is not exist")
		}
		return err
	}

	return nil
}

func (c *Command) excludeFiles() []string {
	var excludedFiles []string
	for _, f := range c.excludeList {
		excludedFiles = append(excludedFiles, ":(exclude,top)"+f)
	}
	return excludedFiles
}

func (c *Command) diffNames() *exec.Cmd {
	args := []string{
		"diff",
		"--name-only",
	}

	if c.isAmend {
		args = append(args, "HEAD^", "HEAD")
	} else {
		args = append(args, "--staged")
	}

	excludedFiles := c.excludeFiles()
	args = append(args, excludedFiles...)

	return exec.Command(
		"git",
		args...,
	)
}

func (c *Command) diffFiles() *exec.Cmd {
	args := []string{
		"diff",
		"--ignore-all-space",
		"--diff-algorithm=minimal",
		"--unified=" + strconv.Itoa(c.diffUnified),
	}

	if c.isAmend {
		args = append(args, "HEAD^", "HEAD")
	} else {
		args = append(args, "--staged")
	}

	excludedFiles := c.excludeFiles()
	args = append(args, excludedFiles...)

	return exec.Command(
		"git",
		args...,
	)
}

func (c *Command) hookPath() *exec.Cmd {
	args := []string{
		"rev-parse",
		"--git-path",
		"hooks",
	}

	return exec.Command(
		"git",
		args...,
	)
}

func (c *Command) gitDir() *exec.Cmd {
	args := []string{
		"rev-parse",
		"--git-dir",
	}

	return exec.Command(
		"git",
		args...,
	)
}

func (c *Command) commit(val string) *exec.Cmd {
	args := []string{
		"commit",
		"--no-verify",
		"--signoff",
		fmt.Sprintf("--message=%s", val),
	}

	if c.isAmend {
		args = append(args, "--amend")
	}

	return exec.Command(
		"git",
		args...,
	)
}
