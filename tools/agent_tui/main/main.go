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
	INTERACTIVE_UI_SENTINEL_PATH = "/etc/assisted/interactive-ui"
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
	rendezvousIP := getRendezvousIP()
	interactiveUIMode := IsInteractiveUIEnabled()

	ctx := agent_tui.AppContext{
		App:               nil,
		RendezvousIP:      rendezvousIP,
		InteractiveUIMode: interactiveUIMode,
		Config: checks.Config{
			ReleaseImageURL: releaseImage,
			LogPath:         logPath,
		},
	}
	agent_tui.App(ctx)
}

// getRendezvousIP reads NODE_ZERO_IP from /etc/assisted/rendezvous-host.env.
func getRendezvousIP() (nodeZeroIP string) {
	envMap, err := godotenv.Read(ui.RENDEZVOUS_HOST_ENV_PATH)
	if err != nil {
		return ""
	}
	nodeZeroIP = envMap["NODE_ZERO_IP"]
	if nodeZeroIP == RENDEZVOUS_IP_TEMPLATE_VALUE {
		nodeZeroIP = ""
	}

	return nodeZeroIP
}

// IsInteractiveUIEnabled checks if the interactive UI sentinel file exists.
// Returns true if /etc/assisted/interactive-ui exists, false otherwise.
func IsInteractiveUIEnabled() bool {
	_, err := os.Stat(INTERACTIVE_UI_SENTINEL_PATH)
	return err == nil
}
