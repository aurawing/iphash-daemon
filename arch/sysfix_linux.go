package arch

import (
	"os/exec"
)

func ExtExecution() string {
	return ""
}

func ExtScript() string {
	return ".sh"
}

func CommandExecuteFix(commands ...string) *exec.Cmd {
	return exec.Command(commands[0], commands[1:])
}
