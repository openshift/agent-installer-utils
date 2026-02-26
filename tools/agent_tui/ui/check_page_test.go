package ui

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/openshift/agent-installer-utils/tools/agent_tui/checks"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestCheckPageNavigation(t *testing.T) {
	config := checks.Config{
		ReleaseImageURL: "",
		LogPath:         "/tmp/agent-tui.log",
	}

	logger := logrus.New()
	ui := NewUI(tview.NewApplication(), config, logger, "")

	// There are two buttons that the user can navigate between
	// <Configure Network> and <Quit>
	// <Configure Network> is position 0
	// <Quit> is position 1
	assert.Equal(t, 0, ui.focusedItem)

	// Note: Disabling the ESC key in check_page mainFlex has
	// no affect on the results of this test. It does have
	// an affect when agent-tui is executed.
	applyKeyToChecks(ui, tcell.KeyESC, 1)
	assert.Equal(t, 0, ui.focusedItem)
	applyKeyToChecks(ui, tcell.KeyTAB, 1) // TAB from <Configure Network> to <Quit>
	assert.Equal(t, 1, ui.focusedItem)
	applyKeyToChecks(ui, tcell.KeyESC, 1) // ESC should have no affect
	assert.Equal(t, 1, ui.focusedItem)
	applyKeyToChecks(ui, tcell.KeyTAB, 1) // TAB from <Quit> to <Configure Network>
	assert.Equal(t, 0, ui.focusedItem)
}

func TestRendezvousIPPageNavigation(t *testing.T) {
	config := checks.Config{
		ReleaseImageURL: "",
		LogPath:         "/tmp/agent-tui.log",
	}

	logger := logrus.New()
	ui := NewUI(tview.NewApplication(), config, logger, "")

	// Focus on rendezvous IP page
	ui.setFocusToRendezvousIP()

	// There are four focusable items on the rendezvous IP page:
	// 0: Input field (Rendezvous IP)
	// 1: Save rendezvous IP button
	// 2: This is the rendezvous node button
	// 3: Configure Network button
	assert.Equal(t, 0, ui.focusedItem)

	// Test TAB navigation (forward)
	applyKeyToRendezvousIPPage(ui, tcell.KeyTab, 1)
	assert.Equal(t, 1, ui.focusedItem) // Save button
	applyKeyToRendezvousIPPage(ui, tcell.KeyTab, 1)
	assert.Equal(t, 2, ui.focusedItem) // This is rendezvous node button
	applyKeyToRendezvousIPPage(ui, tcell.KeyTab, 1)
	assert.Equal(t, 3, ui.focusedItem) // Configure Network button
	applyKeyToRendezvousIPPage(ui, tcell.KeyTab, 1)
	assert.Equal(t, 0, ui.focusedItem) // Back to input field

	// Test BACKTAB navigation (backward)
	applyKeyToRendezvousIPPage(ui, tcell.KeyBacktab, 1)
	assert.Equal(t, 3, ui.focusedItem) // Configure Network button
	applyKeyToRendezvousIPPage(ui, tcell.KeyBacktab, 1)
	assert.Equal(t, 2, ui.focusedItem) // This is rendezvous node button
	applyKeyToRendezvousIPPage(ui, tcell.KeyBacktab, 1)
	assert.Equal(t, 1, ui.focusedItem) // Save button
	applyKeyToRendezvousIPPage(ui, tcell.KeyBacktab, 1)
	assert.Equal(t, 0, ui.focusedItem) // Input field
}

func TestRendezvousIPPageNavigationWithPrefilled(t *testing.T) {
	config := checks.Config{
		ReleaseImageURL: "",
		LogPath:         "/tmp/agent-tui.log",
	}

	logger := logrus.New()
	prefilledIP := "192.168.111.80"
	ui := NewUI(tview.NewApplication(), config, logger, prefilledIP)

	// Verify the input field has the prefilled IP
	inputField := ui.rendezvousIPForm.GetFormItemByLabel(FIELD_ENTER_RENDEZVOUS_IP)
	assert.NotNil(t, inputField)

	// Verify initial rendezvous IP is set
	assert.Equal(t, prefilledIP, ui.initialRendezvousIP)
}

func TestInteractiveUIModeWithPrefilledIP(t *testing.T) {
	config := checks.Config{
		ReleaseImageURL: "",
		LogPath:         "/tmp/agent-tui.log",
	}

	logger := logrus.New()
	prefilledIP := "192.168.111.80"
	ui := NewUI(tview.NewApplication(), config, logger, prefilledIP)
	controller := NewController(ui)

	// Initialize with interactive mode and prefilled IP
	controller.Init(1, prefilledIP, true)

	// Verify timeout modal is active
	assert.True(t, ui.IsRendezvousIPTimeoutActive())
}

func TestInteractiveUIModeWithoutPrefilledIP(t *testing.T) {
	config := checks.Config{
		ReleaseImageURL: "",
		LogPath:         "/tmp/agent-tui.log",
	}

	logger := logrus.New()
	ui := NewUI(tview.NewApplication(), config, logger, "")
	controller := NewController(ui)

	// Initialize with interactive mode but no prefilled IP
	controller.Init(1, "", true)

	// Verify timeout modal is NOT active
	assert.False(t, ui.IsRendezvousIPTimeoutActive())
}

func TestNonInteractiveUIMode(t *testing.T) {
	config := checks.Config{
		ReleaseImageURL: "",
		LogPath:         "/tmp/agent-tui.log",
	}

	logger := logrus.New()
	ui := NewUI(tview.NewApplication(), config, logger, "")
	controller := NewController(ui)

	// Initialize with non-interactive mode
	controller.Init(1, "", false)

	// Verify splash screen is shown and timeout modal is NOT active
	assert.False(t, ui.IsRendezvousIPTimeoutActive())
}

func TestTimeoutModalCancellation(t *testing.T) {
	config := checks.Config{
		ReleaseImageURL: "",
		LogPath:         "/tmp/agent-tui.log",
	}

	logger := logrus.New()
	prefilledIP := "192.168.111.80"
	ui := NewUI(tview.NewApplication(), config, logger, prefilledIP)

	// Show timeout modal
	ui.ShowRendezvousIPTimeoutDialog(prefilledIP)
	assert.True(t, ui.IsRendezvousIPTimeoutActive())

	// Cancel the timeout modal
	ui.cancelRendezvousIPTimeout()
	assert.False(t, ui.IsRendezvousIPTimeoutActive())
}

func applyKeyToChecks(u *UI, key tcell.Key, numKeyPresses int) {
	for i := 0; i < numKeyPresses; i++ {
		u.mainFlex.InputHandler()(tcell.NewEventKey(key, 0, tcell.ModNone), func(p tview.Primitive) {})
	}
}

func applyKeyToRendezvousIPPage(u *UI, key tcell.Key, numKeyPresses int) {
	// Get the rendezvous IP page from pages
	page, _ := u.pages.GetFrontPage()
	if page == PAGE_RENDEZVOUS_IP {
		for i := 0; i < numKeyPresses; i++ {
			// Apply key event through the main flex which has the input capture logic
			u.rendezvousIPMainFlex.InputHandler()(tcell.NewEventKey(key, 0, tcell.ModNone), func(p tview.Primitive) {})
		}
	}
}
