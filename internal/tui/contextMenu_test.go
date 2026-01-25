//go:build !integration
// +build !integration

package tui

import (
	"testing"

	"github.com/charmbracelet/x/exp/teatest"
)

func TestContextMenuOpens(t *testing.T) {
	m := newTestContextMenuModel("test-workspace")

	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(testTermWidth, testTermHeight))

	quitAndWait(tm, t)
}

func TestContextMenuOptions(t *testing.T) {
	m := newTestContextMenuModel("test-workspace")

	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(testTermWidth, testTermHeight))

	quitAndWait(tm, t)
}

func TestContextMenuNavigate(t *testing.T) {
	m := newTestContextMenuModel("test-workspace")

	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(testTermWidth, testTermHeight))

	navigateDown(tm)

	quitAndWait(tm, t)
}

func TestContextMenuCancelEsc(t *testing.T) {
	m := newTestContextMenuModel("test-workspace")

	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(testTermWidth, testTermHeight))

	pressEsc(tm)

	tm.WaitFinished(t, teatest.WithFinalTimeout(testTimeout))
}

func TestContextMenuCancelCtrlC(t *testing.T) {
	m := newTestContextMenuModel("test-workspace")

	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(testTermWidth, testTermHeight))

	pressCtrlC(tm)

	tm.WaitFinished(t, teatest.WithFinalTimeout(testTimeout))
}

func TestContextMenuNavigationWithJAndK(t *testing.T) {
	m := newTestContextMenuModel("test-workspace")

	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(testTermWidth, testTermHeight))

	pressKey(tm, "j")

	quitAndWait(tm, t)
}

func TestContextMenuAllDescriptions(t *testing.T) {
	m := newTestContextMenuModel("test-workspace")

	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(testTermWidth, testTermHeight))

	quitAndWait(tm, t)
}
