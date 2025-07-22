package notify

import (
	"os/exec"
)

func Send(title, memo, url string) error {
	args := []string {"-title", title}
	if memo != "" {
		args = append(args, "-message", memo)
	}
	if url != "" {
		args = append(args, "-open", url)
	}
	cmd := exec.Command("terminal-notifier", args...)
	return cmd.Run()
}