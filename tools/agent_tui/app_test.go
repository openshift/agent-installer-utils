package agent_tui

import (
	"testing"
)

func TestInitialScreen(t *testing.T) {
	app := NewAppTester(t)
	defer app.Stop()

	app.Start()
	app.WaitForButton("Yes")

	app.WaitForScreenContent(
		"Do you wish for this node",
		"to be the one that runs",
		"the installation service",
		"(only one node may perform",
		"this function)?",
		"Yes     No")

	//app.DumpScreen()
}

func TestInsertInvalidRendezvousIP(t *testing.T) {
	app := NewAppTester(t)
	defer app.Stop()

	app.Start()
	app.WaitForButton("Yes")
	//app.DumpScreen()

	// Select the "No" button
	app.ScreenPressTab()
	app.ScreenPressEnter()

	// Wait for the node form, and insert an invalid ip
	app.WaitForInputField("Rendezvous IP Address")
	//app.DumpScreen()

	app.ScreenTypeText("256.256.256.256")
	app.ScreenPressEnter()

	app.WaitForScreenContent("Tzhe specified Rendezvous IP is not a valid IP Address")
	//app.DumpScreen()
}
