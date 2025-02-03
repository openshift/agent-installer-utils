package agent_tui

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/openshift/agent-installer-utils/tools/agent_tui/checks"
	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
)

const (
	waitTimeout = 35 * time.Second
)

// AppTester is a test helper class to allow writing integration tests.
// It works by injecting a SimulationScreen, a normal Screen object with
// some additional testing features.
// The class offers some methods to interact with the Agent TUI screen, so
// that it would be possible to mimick exactly the user actions required to
// accomplish a given activity.
// In addition, a number of verification methods are present to check the results
// of the action performed (WaitFor...)
type AppTester struct {
	t      *testing.T
	screen tcell.SimulationScreen
	app    *tview.Application

	checkResults map[string]string
	wrapper      checks.CheckFunction
}

// Creates a new instance of AppTester
func NewAppTester(t *testing.T, debug ...bool) *AppTester {

	s := tcell.NewSimulationScreen("")
	if s == nil {
		t.Fatalf("Unable to create simulation screen")
	}
	if err := s.Init(); err != nil {
		t.Fatalf("Failed to initialize screen: %v", err)
	}

	app := &AppTester{
		t:            t,
		screen:       s,
		app:          tview.NewApplication().SetScreen(s),
		checkResults: make(map[string]string),
	}
	app.wrapper = func(checkType string, config checks.Config) ([]byte, error) {
		res, found := app.checkResults[checkType]
		if !found || res == "" {
			return []byte("Ok"), nil
		}
		return []byte(res), errors.New(res)
	}
	return app
}

// Starts a new Agent TUI in background
func (a *AppTester) Start(config checks.Config) *AppTester {
	go App(a.app, "192.168.111.80", config, checks.CheckFunctions{
		checks.CheckTypeReleaseImageHostDNS:  a.wrapper,
		checks.CheckTypeReleaseImageHostPing: a.wrapper,
		checks.CheckTypeReleaseImageHttp:     a.wrapper,
		checks.CheckTypeReleaseImagePull:     a.wrapper,
	})
	return a
}

// Releases all the resources and stop the app
func (a *AppTester) Stop() {
	a.app.Stop()
}

// FocusItem loops over the current focusable items until
// it will find a primitive matching the specified caption
func (a *AppTester) FocusItem(caption string) *AppTester {
	a.t.Helper()
	ok := assert.Eventually(a.t, func() bool {
		p := a.app.GetFocus()
		switch v := p.(type) {
		case *tview.Button:
			if v.GetLabel() == caption {
				return true
			}
		case *tview.InputField:
			if v.GetLabel() == caption {
				return true
			}
		default:
			a.t.Logf("Item type %T not managed, skipping", v)
		}
		// Move to the next focusable item
		a.ScreenPressTab()
		return false
	}, waitTimeout, 2*time.Millisecond)

	if !ok {
		assert.FailNow(a.t, fmt.Sprintf("widget with caption '%s' not found", caption))
	}

	return a
}

// SelectItem loops over the current focusable items until
// it will find a primitive matching the specified caption.
// If found, the item will be selected by pressing the Enter key
func (a *AppTester) SelectItem(caption string) *AppTester {
	a.t.Helper()
	a.FocusItem(caption)
	return a.ScreenPressEnter()
}

// Moves the current cursor to the right
func (a *AppTester) ScreenMoveCursorRight() *AppTester {
	return a.screenPressKey(tcell.KeyRight)
}

// Press the Tab key
func (a *AppTester) ScreenPressTab() *AppTester {
	return a.screenPressKey(tcell.KeyTAB)
}

// Press the Enter key
func (a *AppTester) ScreenPressEnter() *AppTester {
	return a.screenPressKey(tcell.KeyEnter)
}

func (a *AppTester) screenPressKey(key tcell.Key) *AppTester {
	a.screen.InjectKey(key, rune(0), tcell.ModNone)
	time.Sleep(1 * time.Millisecond)
	return a
}

// Types a string at the current screen position and then press enter
func (a *AppTester) ScreenTypeText(text string) *AppTester {
	for _, c := range text {
		a.screen.InjectKey(tcell.KeyRune, rune(c), tcell.ModNone)
		time.Sleep(1 * time.Millisecond)
	}
	return a
}

func (a *AppTester) fetchScreenContent() []string {
	cells, w, h := a.screen.GetContents()
	lines := []string{}

	for y := 0; y < h; y++ {
		line := ""
		for x := 0; x < w; x++ {
			if len(cells[x+y*w].Runes) == 0 {
				continue
			}
			r := cells[x+y*w].Runes[0]
			if !unicode.IsSymbol(r) ||
				r == '✓' ||
				r == '✖' {
				line += string(r)
			}
		}
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, strings.TrimSpace(line))
		}
	}
	return lines
}

// Wait until the current screen buffer contains the specified labels, or timeout
func (a *AppTester) WaitForScreenContent(labels ...string) *AppTester {
	a.t.Helper()
	ok := assert.Eventually(a.t, func() bool {
		lines := a.fetchScreenContent()
		for _, label := range labels {
			found := false
			for _, l := range lines {
				if strings.Contains(l, label) {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
		return true
		//Some tasks may take a while to display their output in the screen
	}, 20*time.Second, 10*time.Millisecond)

	if !ok {
		a.DumpScreen()
		assert.FailNow(a.t, fmt.Sprintf("Screen does not contain '%s'", labels))
	}

	return a
}

// Print the content of the current screen to the terminal in a raw format.
// Just useful for debugging.
func (a *AppTester) DumpScreen() *AppTester {
	cells, w, h := a.screen.GetContents()

	rows := []string{"\n"}
	for y := 0; y < h; y++ {
		row := ""
		for x := 0; x < w; x++ {
			c := cells[x+y*w]
			row += string(c.Bytes)
		}
		if strings.TrimSpace(row) != "" {
			row += "\n"
			rows = append(rows, row)
		}
	}
	a.t.Log(rows)

	return a
}

func (a *AppTester) setCheckResult(cType string, res string) *AppTester {
	a.checkResults[cType] = res
	return a
}

// Set the error for the next pull image checks.
func (a *AppTester) SetPullCheckError(res string) *AppTester {
	return a.setCheckResult(checks.CheckTypeReleaseImagePull, res)
}

// Reset pull check results.
func (a *AppTester) SetPullCheckOk() *AppTester {
	return a.setCheckResult(checks.CheckTypeReleaseImagePull, "")
}

// Set the error for the next http get checks.
func (a *AppTester) SetHttpCheckError(res string) *AppTester {
	return a.setCheckResult(checks.CheckTypeReleaseImageHttp, res)
}

// Reset http get check results.
func (a *AppTester) SetHttpCheckOk() *AppTester {
	return a.setCheckResult(checks.CheckTypeReleaseImageHttp, "")
}

// Set the error for the next ping checks.
func (a *AppTester) SetPingCheckError(res string) *AppTester {
	return a.setCheckResult(checks.CheckTypeReleaseImageHostPing, res)
}

// Reset ping check results.
func (a *AppTester) SetPingCheckOk() *AppTester {
	return a.setCheckResult(checks.CheckTypeReleaseImageHostPing, "")
}

// Set the error for the next DNS checks.
func (a *AppTester) SetDNSCheckError(res string) *AppTester {
	return a.setCheckResult(checks.CheckTypeReleaseImageHostDNS, res)
}

// Reset DNS check results.
func (a *AppTester) SetDNSCheckOk() *AppTester {
	return a.setCheckResult(checks.CheckTypeReleaseImageHostDNS, "")
}
