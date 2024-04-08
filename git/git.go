package git

import (
	"os/exec"
	"strings"
)

func Root() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	gitRoot := strings.TrimSpace(string(output))
	return gitRoot, nil
}
