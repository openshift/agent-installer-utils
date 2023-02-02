package agent_tui

import (
	"testing"
)

func TestInitialScreen(t *testing.T) {
	app := NewAppTester(t)
	defer app.Stop()

	app.Start().
		WaitForScreenContent(
			"Do you wish for this node",
			"to be the one that runs",
			"the installation service",
			"(only one node may perform",
			"this function)?",
			"Yes     No")
}

func TestInsertInvalidRendezvousIP(t *testing.T) {
	app := NewAppTester(t)
	defer app.Stop()

	app.Start().
		// Move to the node form
		SelectItem("No").

		// Insert an invalid ip
		FocusItem("Rendezvous IP Address").
		ScreenTypeText("256.256.256.256").ScreenPressTab().
		WaitForScreenContent("The specified Rendezvous IP is not a valid IP Address")
}

func TestCheckConnectivity(t *testing.T) {
	app := NewAppTester(t)
	defer app.Stop()

	app.Start().
		// Move to the node form
		SelectItem("No").

		// Wait for the node form, and insert an invalid ip
		FocusItem("Rendezvous IP Address").
		ScreenTypeText("127.0.0.1").

		// Press "Check connectivity" button
		SelectItem("Check connectivity").
		WaitForScreenContent("Connectivity check successful")
}

func TestCheckConnectivityFailure(t *testing.T) {
	app := NewAppTester(t)
	defer app.Stop()

	app.Start().
		// Move to the node form
		SelectItem("No").

		// Wait for the node form, and insert an invalid ip
		FocusItem("Rendezvous IP Address").
		ScreenTypeText("196.0.0.1").

		// Press "Check connectivity" button
		SelectItem("Check connectivity").
		WaitForScreenContent("Failed to connect to 196.0.0.1 (exit status 1)")
}
