package net

import (
	"os/exec"
)

func ValidateConnectivity(address string) ([]byte, error) {
	return exec.Command("ping", "-c", "4", address).CombinedOutput()
}
