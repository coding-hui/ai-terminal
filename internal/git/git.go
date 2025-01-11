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
  ai commit --file $1 --preview --no-confirm
fi
`
)

var excludeFromDiff = []string{
	"package-lock.json",
	"pnpm-lock.yaml",
	// yarn.lock, Cargo.lock, Gemfile.lock, Pipfile.lock, etc.
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

func (c *Command) AddFiles(files []string) error {
	for _, file := range files {
		output, err := exec.Command("git", "add", file).CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to add file %s: %w, output: %s", file, err, string(output))
		}
	}
	return nil
}

func (c *Command) Commit(val string) (string, error) {
	output, err := c.commit(val).Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

// RollbackLastCommit rolls back the most recent commit, leaving changes staged.
func (c *Command) RollbackLastCommit() error {
	output, err := exec.Command("git", "reset", "--hard", "HEAD~1").CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to rollback last commit: %w, output: %s", err, string(output))
	}
	return nil
}

// GitDir to show the (by default, absolute) path of the git directory of the working tree.
func (c *Command) GitDir() (string, error) {
	output, err := c.gitDir().Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

func (c *Command) ListAllFiles() ([]string, error) {
	cmd := exec.Command("git", "ls-files")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	fileNames := strings.Split(strings.TrimSpace(string(output)), "\n")
	return fileNames, nil
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
	excludedFiles := []string{}
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

// FormatDiff formats git diff output with color and stats
func (c *Command) FormatDiff(diffOutput string) string {
	lines := strings.Split(diffOutput, "\n")
	var formattedLines []string
	var stats struct {
		added   int
		removed int
		total   int
	}

	// First pass to calculate stats
	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "+"):
			stats.added++
			stats.total++
		case strings.HasPrefix(line, "-"):
			stats.removed++
			stats.total++
		case strings.HasPrefix(line, "@@"):
			stats.total++
		}
	}

	// Create progress bar
	progress := c.createProgressBar(stats.added, stats.removed, stats.total)

	// Second pass to format lines
	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "+"):
			formattedLines = append(formattedLines,
				fmt.Sprintf("\x1b[32m%s\x1b[0m", line)) // Green for additions
		case strings.HasPrefix(line, "-"):
			formattedLines = append(formattedLines,
				fmt.Sprintf("\x1b[31m%s\x1b[0m", line)) // Red for deletions
		case strings.HasPrefix(line, "@@"):
			formattedLines = append(formattedLines,
				fmt.Sprintf("\x1b[33m%s\x1b[0m", line)) // Yellow for headers
		default:
			formattedLines = append(formattedLines, line)
		}
	}

	// Add header with stats
	header := fmt.Sprintf("Changes (%d added, %d removed):\n%s\n",
		stats.added, stats.removed, progress)

	return header + strings.Join(formattedLines, "\n")
}

// createProgressBar generates a visual progress bar for diff stats
func (c *Command) createProgressBar(added, removed, total int) string {
	const barWidth = 30
	if total == 0 {
		return ""
	}

	addedBlocks := int(float64(added) / float64(total) * barWidth)
	removedBlocks := int(float64(removed) / float64(total) * barWidth)
	unchangedBlocks := barWidth - addedBlocks - removedBlocks

	bar := strings.Repeat("█", addedBlocks) +
		strings.Repeat("░", removedBlocks) +
		strings.Repeat(" ", unchangedBlocks)

	return fmt.Sprintf("[%s] %d changes", bar, total)
}
