package arch

import (
	"os/exec"
)

func ExtExecution() string {
	return ".exe"
}

func ExtScript() string {
	return ".bat"
}

func CommandExecuteFix(commands ...string) *exec.Cmd {
	cmms := append([]string{"/C"}, commands...)
	return exec.Command("cmd", cmms...)
}
