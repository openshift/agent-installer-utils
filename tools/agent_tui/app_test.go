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

func TestRendezvousIP(t *testing.T) {
	cases := []struct {
		name  string
		steps func(app *AppTester)
	}{
		{
			name: "invalid ip",
			steps: func(app *AppTester) {
				app.Start().
					// Move to the node form
					SelectItem("No").

					// Insert an invalid ip
					FocusItem("Rendezvous IP Address").
					ScreenTypeText("256.256.256.256").ScreenPressTab().
					WaitForScreenContent("The specified Rendezvous IP is not a valid IP Address")
			},
		},
	}
	for _, tc := range cases {
		steps := tc.steps
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			app := NewAppTester(t)
			defer app.Stop()

			steps(app)
		})
	}
}

func TestCheckConnectivity(t *testing.T) {
	cases := []struct {
		name  string
		steps func(app *AppTester)
	}{
		{
			name: "connectivity ok",
			steps: func(app *AppTester) {
				app.Start().
					// Move to the node form
					SelectItem("No").

					// Wait for the node form, and insert an invalid ip
					FocusItem("Rendezvous IP Address").
					ScreenTypeText("127.0.0.1").

					// Press "Check connectivity" button
					SelectItem("Check connectivity").
					WaitForScreenContent("Connectivity check successful")
			},
		},
		{
			name: "connectivity failure",
			steps: func(app *AppTester) {
				app.Start().
					// Move to the node form
					SelectItem("No").

					// Wait for the node form, and insert an invalid ip
					FocusItem("Rendezvous IP Address").
					ScreenTypeText("196.0.0.1").

					// Press "Check connectivity" button
					SelectItem("Check connectivity").
					WaitForScreenContent("Failed to connect to 196.0.0.1 (exit status 1)")
			},
		},
	}
	for _, tc := range cases {
		steps := tc.steps
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			app := NewAppTester(t)
			defer app.Stop()

			steps(app)
		})
	}
}
