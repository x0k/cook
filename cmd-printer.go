package main

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
)

type cmdReceiptLinesPrinter struct {
	cmd *exec.Cmd
}

func NewCmdReceiptLinesPrinter(command string) (ReceiptLinesPrinter, error) {
	parts := strings.Fields(command)
	if len(parts) < 1 {
		return nil, errors.New("invalid command line")
	}
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	return &cmdReceiptLinesPrinter{
		cmd: cmd,
	}, nil
}

// Print implements ReceiptLinesPrinter.
func (p *cmdReceiptLinesPrinter) Print(lines string) error {
	// Create a pipe for stdin.
	stdin, err := p.cmd.StdinPipe()
	if err != nil {
		return err
	}
	defer stdin.Close()
	if err := p.cmd.Start(); err != nil {
		return err
	}
	_, err = io.WriteString(stdin, lines)
	if err != nil {
		return err
	}
	err = stdin.Close()
	if err != nil {
		return err
	}
	return p.cmd.Wait()
}