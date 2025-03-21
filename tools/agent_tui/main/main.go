package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"

	"github.com/openshift/agent-installer-utils/tools/agent_tui"
	"github.com/openshift/agent-installer-utils/tools/agent_tui/checks"
	"github.com/openshift/agent-installer-utils/tools/agent_tui/ui"
)

const (
	RENDEZVOUS_IP_TEMPLATE_VALUE = "{{.RendezvousIP}}"
)

func main() {
	releaseImage := os.Getenv("RELEASE_IMAGE")
	logPath := os.Getenv("AGENT_TUI_LOG_PATH")

	if releaseImage == "" {
		fmt.Println("RELEASE_IMAGE environment variable is not specified.")
		fmt.Println("Unable to perform connectivity checks.")
		fmt.Println("Exiting agent-tui.")
		os.Exit(1)
	}
	if logPath == "" {
		logPath = "/tmp/agent_tui.log"
		fmt.Printf("AGENT_TUI_LOG_PATH is unspecified, logging to: %v\n", logPath)
	}
	config := checks.Config{
		ReleaseImageURL: releaseImage,
		LogPath:         logPath,
	}
	agent_tui.App(nil, getRendezvousIP(), config)
}

// The rendezvous IP address can be passed into AGENT_TUI
// through the NODE_ZERO_IP environment variable.
// If NODE_ZERO_IP is unset through the environment variable,
// then this function reads /etc/assisted/rendezvous-host.env
// for the value..
func getRendezvousIP() string {
	envMap, err := godotenv.Read(ui.RENDEZVOUS_HOST_ENV_PATH)
	if err != nil {
		return ""
	}
	nodeZeroIP := envMap["NODE_ZERO_IP"]
	if nodeZeroIP == RENDEZVOUS_IP_TEMPLATE_VALUE {
		return ""
	}

	return nodeZeroIP
}
