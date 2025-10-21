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
	ui := NewUI(tview.NewApplication(), config, logger)

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

func applyKeyToChecks(u *UI, key tcell.Key, numKeyPresses int) {
	for i := 0; i < numKeyPresses; i++ {
		u.mainFlex.InputHandler()(tcell.NewEventKey(key, 0, tcell.ModNone), func(p tview.Primitive) {})
	}
}
