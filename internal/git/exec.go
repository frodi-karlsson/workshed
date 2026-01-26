package git

import (
	"context"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type RealGit struct{}

func (RealGit) Clone(ctx context.Context, url, dir string, opts CloneOptions) error {
	args := []string{"clone"}
	if opts.Depth > 0 {
		args = append(args, "--depth", strconv.Itoa(opts.Depth))
	}
	if opts.Mirror {
		args = append(args, "--mirror")
	}
	args = append(args, url, dir)

	cmd := exec.CommandContext(ctx, "git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return ClassifyError("clone", err, output)
	}

	return nil
}

func (RealGit) Checkout(ctx context.Context, dir, ref string) error {
	cmd := exec.CommandContext(ctx, "git", "checkout", ref)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return ClassifyError("checkout", err, output)
	}
	return nil
}

func (RealGit) GetRemoteURL(ctx context.Context, dir string) (string, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}

	cmd := exec.CommandContext(ctx, "git", "remote", "get-url", "origin")
	cmd.Dir = absDir
	output, err := cmd.Output()
	if err != nil {
		return "", ClassifyError("get-url", err, output)
	}

	return strings.TrimSpace(string(output)), nil
}

func (RealGit) CurrentBranch(ctx context.Context, dir string) (string, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}

	cmd := exec.CommandContext(ctx, "git", "branch", "--show-current")
	cmd.Dir = absDir
	output, err := cmd.Output()
	if err != nil {
		return "", ClassifyError("current-branch", err, output)
	}

	return strings.TrimSpace(string(output)), nil
}

func (RealGit) RevParse(ctx context.Context, dir, ref string) (string, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}

	cmd := exec.CommandContext(ctx, "git", "rev-parse", ref)
	cmd.Dir = absDir
	output, err := cmd.Output()
	if err != nil {
		return "", ClassifyError("rev-parse", err, output)
	}

	return strings.TrimSpace(string(output)), nil
}

func (RealGit) StatusPorcelain(ctx context.Context, dir string) (string, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}

	cmd := exec.CommandContext(ctx, "git", "status", "--porcelain")
	cmd.Dir = absDir
	output, err := cmd.Output()
	if err != nil {
		return "", ClassifyError("status", err, output)
	}

	return strings.TrimSpace(string(output)), nil
}
