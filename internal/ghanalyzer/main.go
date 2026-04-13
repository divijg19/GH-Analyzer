package ghanalyzer

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func Main() {
	if err := run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.ExitCode())
		}

		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cmd, err := commandForArgs(os.Args[1:])
	if err != nil {
		return err
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err == nil {
		return nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return err
	}

	return fmt.Errorf("unable to run gh-analyzer: %w", err)
}

func commandForArgs(args []string) (*exec.Cmd, error) {
	if _, err := exec.LookPath("gh-analyzer"); err == nil {
		return exec.Command("gh-analyzer", args...), nil
	}

	if _, err := exec.LookPath("go"); err != nil {
		return nil, fmt.Errorf("unable to run gh-analyzer: executable file not found in $PATH")
	}

	cmd := exec.Command("go", append([]string{"run", "./cmd/gh-analyzer"}, args...)...)
	if root, ok := moduleRoot(); ok {
		cmd.Dir = root
	}

	return cmd, nil
}

func moduleRoot() (string, bool) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", false
	}

	root := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", ".."))
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		return "", false
	}

	return root, true
}
