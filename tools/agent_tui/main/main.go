package main

import (
	"fmt"
	"os"

	"github.com/openshift/agent-installer-utils/tools/agent_tui"
)

func main() {
	releaseImage := os.Getenv("RELEASE_IMAGE")
	nodeZeroIP := os.Getenv("NODE_ZERO_IP")
	if releaseImage == "" || nodeZeroIP == "" {
		if releaseImage == "" {
			fmt.Println("RELEASE_IMAGE environment variable is not specified.")
		}
		if nodeZeroIP == "" {
			fmt.Println("NODE_ZERO_IP environment variable is not specified.")
		}
		fmt.Println("Unable to perform connectivity checks.")
		fmt.Println("Exiting agent-tui.")
		os.Exit(1)
	}
	config := agent_tui.Config{
		ReleaseImageURL:  releaseImage,
		RendezvousHostIP: nodeZeroIP,
	}
	agent_tui.App(nil, config)
}
