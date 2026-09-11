package applecontainer

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// Runner is the injectable Apple Container command boundary.
type Runner interface {
	Run(ctx context.Context, args ...string) ([]byte, error)
	Stream(ctx context.Context, args ...string) (io.ReadCloser, error)
}

// CommandRunner executes the Apple container CLI without a shell.
type CommandRunner struct {
	Bin    string
	Stderr io.Writer
}

func (r CommandRunner) binary() string {
	if r.Bin != "" {
		return r.Bin
	}
	return "container"
}

func (r CommandRunner) Run(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, r.binary(), args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if r.Stderr != nil && stderr.Len() > 0 {
			_, _ = io.Copy(r.Stderr, bytes.NewReader(stderr.Bytes()))
		}
		return nil, fmt.Errorf("container %v: %w: %s", args, err, bytes.TrimSpace(stderr.Bytes()))
	}
	return stdout.Bytes(), nil
}

func (r CommandRunner) Stream(ctx context.Context, args ...string) (io.ReadCloser, error) {
	cmd := exec.CommandContext(ctx, r.binary(), args...)
	if r.Stderr != nil {
		cmd.Stderr = r.Stderr
	} else {
		cmd.Stderr = os.Stderr
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return &commandStream{ReadCloser: stdout, wait: cmd.Wait}, nil
}

type commandStream struct {
	io.ReadCloser
	wait func() error
}

func (s *commandStream) Close() error {
	closeErr := s.ReadCloser.Close()
	waitErr := s.wait()
	if closeErr != nil {
		return closeErr
	}
	return waitErr
}
