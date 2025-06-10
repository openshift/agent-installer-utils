package ui

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
)

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
