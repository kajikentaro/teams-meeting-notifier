package utils

import (
	"log"
	"os/exec"
)

func ExecCommand(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	err := cmd.Start()
	if err != nil {
		log.Printf("failed to execute command: %s %v: %v", command, args, err)
	}
	return err
}
