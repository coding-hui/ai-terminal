package git

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/coding-hui/ai-terminal/internal/ui/console"
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

func (c *Command) CommitWithAuthor(val, authorName, authorEmail string) (string, error) {
	output, err := c.commitWithAuthor(val, authorName, authorEmail).Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

func (c *Command) CommitWithAuthorAndCommitter(val, authorName, authorEmail, committerName, committerEmail string) (string, error) {
	output, err := c.commitWithAuthorAndCommitter(val, authorName, authorEmail, committerName, committerEmail).Output()
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

	err = os.WriteFile(target, []byte(HookPrepareCommitMessageTemplate), 0o600)
	if err != nil {
		return err
	}

	// Set executable permission
	err = os.Chmod(target, 0o755)
	if err != nil {
		return fmt.Errorf("failed to set executable permission: %w", err)
	}

	return nil
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

func (c *Command) commitWithAuthor(val, authorName, authorEmail string) *exec.Cmd {
	args := []string{
		"commit",
		"--no-verify",
		"--signoff",
		fmt.Sprintf("--message=%s", val),
		fmt.Sprintf("--author=%s <%s>", authorName, authorEmail),
	}

	if c.isAmend {
		args = append(args, "--amend")
	}

	return exec.Command(
		"git",
		args...,
	)
}

func (c *Command) commitWithAuthorAndCommitter(val, authorName, authorEmail, committerName, committerEmail string) *exec.Cmd {
	args := []string{
		"commit",
		"--no-verify",
		"--signoff",
		fmt.Sprintf("--message=%s", val),
	}

	if authorName != "" && authorEmail != "" {
		args = append(args, fmt.Sprintf("--author=%s <%s>", authorName, authorEmail))
	}

	if c.isAmend {
		args = append(args, "--amend")
	}

	cmd := exec.Command("git", args...)

	// Set committer via environment variables
	if committerName != "" && committerEmail != "" {
		cmd.Env = append(os.Environ(),
			fmt.Sprintf("GIT_COMMITTER_NAME=%s", committerName),
			fmt.Sprintf("GIT_COMMITTER_EMAIL=%s", committerEmail),
		)
	}

	return cmd
}

// DiffStats holds statistics about the diff
type DiffStats struct {
	Added      int
	Removed    int
	Total      int
	Context    int
	FileHeader string
}

// FormatDiff formats git diff output with color and stats
func (c *Command) FormatDiff(diffOutput string) string {
	if diffOutput == "" {
		return ""
	}

	lines := strings.Split(diffOutput, "\n")
	stats := c.calculateDiffStats(lines)
	formattedLines := c.formatDiffLines(lines)

	// Add separator before file header
	if stats.FileHeader != "" {
		separator := strings.Repeat("─", 80)
		formattedLines = append([]string{
			console.StdoutStyles().DiffContext.Render(separator),
			"",
		}, formattedLines...)
	}

	header := c.createDiffHeader(stats)
	return header + "\n" + strings.Join(formattedLines, "\n")
}

// calculateDiffStats calculates statistics from diff lines
func (c *Command) calculateDiffStats(lines []string) *DiffStats {
	stats := &DiffStats{}
	var origLineNum, updatedLineNum int
	var lastNonDeleted int

	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "+++"):
			stats.FileHeader = line
		case strings.HasPrefix(line, "+"):
			stats.Added++
			stats.Total++
			updatedLineNum++
			lastNonDeleted = updatedLineNum
		case strings.HasPrefix(line, "-"):
			stats.Removed++
			stats.Total++
			origLineNum++
		case strings.HasPrefix(line, "@@"):
			stats.Total++
			if matches := regexp.MustCompile(`@@ -(\d+),\d+ \+(\d+),\d+ @@`).FindStringSubmatch(line); len(matches) > 0 {
				origLineNum, _ = strconv.Atoi(matches[1])
				updatedLineNum, _ = strconv.Atoi(matches[2])
				lastNonDeleted = updatedLineNum
			}
		case strings.HasPrefix(line, " "):
			stats.Context++
			stats.Total++
			origLineNum++
			updatedLineNum++
			lastNonDeleted = updatedLineNum
		}
	}

	stats.Context = lastNonDeleted
	return stats
}

// formatDiffLines formats diff lines with proper styling
func (c *Command) formatDiffLines(lines []string) []string {
	var formattedLines []string
	var origLineNum, updatedLineNum int

	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---"):
			formattedLines = append(formattedLines,
				console.StdoutStyles().DiffFileHeader.Render(line))
		case strings.HasPrefix(line, "@@"):
			formattedLines = append(formattedLines,
				console.StdoutStyles().DiffHunkHeader.Render(line))
			if matches := regexp.MustCompile(`@@ -(\d+),\d+ \+(\d+),\d+ @@`).FindStringSubmatch(line); len(matches) > 0 {
				origLineNum, _ = strconv.Atoi(matches[1])
				updatedLineNum, _ = strconv.Atoi(matches[2])
			}
		case strings.HasPrefix(line, "+"):
			formattedLines = append(formattedLines,
				console.StdoutStyles().DiffAdded.Render(fmt.Sprintf("%4d +%s", updatedLineNum, line[1:])))
			updatedLineNum++
		case strings.HasPrefix(line, "-"):
			formattedLines = append(formattedLines,
				console.StdoutStyles().DiffRemoved.Render(fmt.Sprintf("%4d -%s", origLineNum, line[1:])))
			origLineNum++
		case strings.HasPrefix(line, " "):
			formattedLines = append(formattedLines,
				console.StdoutStyles().DiffContext.Render(fmt.Sprintf("%4d  %s", origLineNum, line[1:])))
			origLineNum++
			updatedLineNum++
		default:
			formattedLines = append(formattedLines, line)
		}
	}

	return formattedLines
}

// createDiffHeader creates the header with stats and progress
func (c *Command) createDiffHeader(stats *DiffStats) string {
	progress := c.createProgressBar(stats.Context, stats.Total)

	return console.StdoutStyles().DiffHeader.Render(
		fmt.Sprintf("Changes (%d added, %d removed, %d unchanged):\n%s\n",
			stats.Added, stats.Removed, stats.Context, progress))
}

// createProgressBar generates a visual progress bar with line numbers
func (c *Command) createProgressBar(lastNonDeleted, totalLines int) string {
	const barWidth = 30

	// Handle edge cases
	if totalLines <= 0 || lastNonDeleted < 0 {
		return ""
	}

	// Ensure lastNonDeleted doesn't exceed totalLines
	if lastNonDeleted > totalLines {
		lastNonDeleted = totalLines
	}

	// Calculate progress based on last non-deleted line
	progress := float64(lastNonDeleted) / float64(totalLines)

	// Calculate bar segments ensuring non-negative values
	filledBlocks := int(progress * barWidth)
	if filledBlocks < 0 {
		filledBlocks = 0
	}
	if filledBlocks > barWidth {
		filledBlocks = barWidth
	}
	emptyBlocks := barWidth - filledBlocks

	bar := strings.Repeat("█", filledBlocks) +
		strings.Repeat(" ", emptyBlocks)

	percentage := progress * 100
	return fmt.Sprintf("%3d / %3d lines [%s] %3.0f%%",
		lastNonDeleted, totalLines, bar, percentage)
}
