package git

import (
	"os/exec"
	"strings"
)

type AbsolutePath = string

func Root() (AbsolutePath, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")

	o, err := cmd.Output()
	if err != nil {
		return "", err
	}

	r := strings.TrimSpace(string(o))
	return r, nil
}
