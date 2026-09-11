package ui

import (
	"net"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/openshift/agent-installer-utils/tools/agent_tui/checks"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestSetFocusToSelectIPRefreshesHostIPList(t *testing.T) {
	originalGetInterfaceAddrs := getInterfaceAddrs
	t.Cleanup(func() {
		getInterfaceAddrs = originalGetInterfaceAddrs
	})

	addresses := []net.Addr{&net.IPNet{IP: net.ParseIP("192.0.2.1")}}
	getInterfaceAddrs = func() ([]net.Addr, error) {
		return addresses, nil
	}

	ui := NewUI(tview.NewApplication(), checks.Config{}, logrus.New(), "")
	assert.Equal(t, []string{"192.0.2.1", EMPTY_OPTION, BACK_BUTTON}, selectIPListItems(ui.selectIPList))

	addresses = []net.Addr{&net.IPNet{IP: net.ParseIP("192.0.2.2")}}
	ui.setFocusToSelectIP()

	assert.Equal(t, []string{"192.0.2.2", EMPTY_OPTION, BACK_BUTTON}, selectIPListItems(ui.selectIPList))
}

// Test the InputCapture works correctly
func TestSelectIPListNavigation(t *testing.T) {
	list := tview.NewList()

	list.AddItem("IP0", "", '0', nil)
	list.AddItem("IP1", "", '1', nil)
	list.AddItem("IP2", "", '2', nil)
	list.AddItem("", "", '3', nil)
	list.AddItem(BACK_BUTTON, "", '4', nil)
	list.AddItem(RENDEZVOUS_CONFIGURE_NETWORK_BUTTON, "", '5', nil)

	list.SetInputCapture(getSelectIPListInputCapture(list))

	assert.Equal(t, 0, list.GetCurrentItem())

	// Press KeyDown
	applyKeyToList(list, tcell.KeyDown, 1)
	assert.Equal(t, 1, list.GetCurrentItem())
	applyKeyToList(list, tcell.KeyDown, 1)
	assert.Equal(t, 2, list.GetCurrentItem())
	applyKeyToList(list, tcell.KeyDown, 1)
	assert.Equal(t, 4, list.GetCurrentItem()) // should skip blank line at position 3 and go to <Back> button
	applyKeyToList(list, tcell.KeyDown, 1)
	assert.Equal(t, 5, list.GetCurrentItem())
	applyKeyToList(list, tcell.KeyDown, 1)
	assert.Equal(t, 0, list.GetCurrentItem()) // back at top of the list

	// Press KeyUP
	applyKeyToList(list, tcell.KeyUp, 1)
	assert.Equal(t, 5, list.GetCurrentItem()) // should go to the bottom to <Configure Network> button
	applyKeyToList(list, tcell.KeyUp, 1)
	assert.Equal(t, 4, list.GetCurrentItem())
	applyKeyToList(list, tcell.KeyUp, 1)
	assert.Equal(t, 2, list.GetCurrentItem()) // should skip blank line at position 3
	applyKeyToList(list, tcell.KeyUp, 1)
	assert.Equal(t, 1, list.GetCurrentItem())
	applyKeyToList(list, tcell.KeyUp, 1)
	assert.Equal(t, 0, list.GetCurrentItem())
}

func TestSelectIPListNavigation1Address(t *testing.T) {
	list := tview.NewList()

	list.AddItem("IP0", "", '0', nil)
	list.AddItem("", "", '1', nil)
	list.AddItem(BACK_BUTTON, "", '2', nil)
	list.AddItem(RENDEZVOUS_CONFIGURE_NETWORK_BUTTON, "", '3', nil)

	list.SetInputCapture(getSelectIPListInputCapture(list))

	assert.Equal(t, 0, list.GetCurrentItem())

	// Press KeyDown
	applyKeyToList(list, tcell.KeyDown, 1)
	assert.Equal(t, 2, list.GetCurrentItem()) // should skip blank line at position 1 and go to <Back> button
	applyKeyToList(list, tcell.KeyDown, 1)
	assert.Equal(t, 3, list.GetCurrentItem())
	applyKeyToList(list, tcell.KeyDown, 1)
	assert.Equal(t, 0, list.GetCurrentItem()) // back at top of the list

	// Press KeyUP
	applyKeyToList(list, tcell.KeyUp, 1)
	assert.Equal(t, 3, list.GetCurrentItem()) // should go to the bottom to <Configure Network> button
	applyKeyToList(list, tcell.KeyUp, 1)
	assert.Equal(t, 2, list.GetCurrentItem())
	applyKeyToList(list, tcell.KeyUp, 1)
	assert.Equal(t, 0, list.GetCurrentItem()) // should skip blank line at position 1
}

func applyKeyToList(list *tview.List, key tcell.Key, numKeyPresses int) {
	for i := 0; i < numKeyPresses; i++ {
		list.InputHandler()(tcell.NewEventKey(key, 0, tcell.ModNone), func(p tview.Primitive) {})
	}
}

func selectIPListItems(list *tview.List) []string {
	items := make([]string, list.GetItemCount())
	for i := range items {
		items[i], _ = list.GetItemText(i)
	}
	return items
}
