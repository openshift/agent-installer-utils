package net

import (
	"os"
	"os/exec"
)

func ValidateConnectivity() ([]byte, error) {
	output, err := CheckRegistryConnectivity()
	if err != nil {
		return output, err
	}
	return CheckRendezvousHostConnectivity()
}

func GetRendezvousHostIP() string {
	return os.Getenv("NODE_ZERO_IP")
}

func CheckRendezvousHostConnectivity() ([]byte, error) {
	return exec.Command("ping", "-c", "4", GetRendezvousHostIP()).CombinedOutput()
}

func GetReleaseImageURL() string {
	return os.Getenv("RELEASE_IMAGE")
}

func CheckRegistryConnectivity() ([]byte, error) {
	return exec.Command("podman", "pull", GetReleaseImageURL()).CombinedOutput()
}
