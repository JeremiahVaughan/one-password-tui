package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m model) View() string {
	if m.data.err != nil {
		return getErrorStyle(m.data.err.Error())
	}

	var display strings.Builder

	switch m.data.activeView {
	case activeViewListItems:
		if m.loading {
			display.WriteString(m.spinner.View())
		} else {
			display.WriteString(m.items.View())
		}
	case activeViewItem:
		display.WriteString(m.itemDetails.View())
		if m.clipboardCopyTriggered {
			display.WriteRune('\t')
			display.WriteString(fmt.Sprintf("%s fetching otp", m.spinner.View()))
		} else if m.loading {
			display.WriteRune('\t')
			display.WriteString(fmt.Sprintf("%s downloading to %s", m.spinner.View(), m.downloadTarget))
		} else if m.clipboardLifeMeter != nil && m.clipboardLifeMeter.Running() {
			// timers start off as running even if they have not started, so using
			// a nil check to get around this issue
			timerStyle := lipgloss.NewStyle()
			secondsElapsed := m.clipboardLifeMeter.Timeout.Seconds()
			if secondsElapsed > clipboardLifeInSeconds*.66 {
				timerStyle = timerStyle.Foreground(lipgloss.Color("#00ff00"))
			} else if secondsElapsed > clipboardLifeInSeconds*.33 {
				timerStyle = timerStyle.Foreground(lipgloss.Color("#ffff00"))
			} else {
				timerStyle = timerStyle.Foreground(lipgloss.Color("#ff0000"))
			}
			display.WriteRune('\t')
			display.WriteString(fmt.Sprintf("clipboard cleanup in %s", timerStyle.Render(m.clipboardLifeMeter.View())))
		} else if m.data.fileDownloaded {
			display.WriteRune('\t')
			display.WriteString(fmt.Sprintf("download complete %s", m.downloadTarget))
		}
	}
	if m.data.validationMsg != "" {
		display.WriteString(getErrorStyle(m.data.validationMsg))
	}

	return docStyle.Render(display.String())
}

func getErrorStyle(errMsg string) string {
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true).Width(80).MarginLeft(4)
	return fmt.Sprintf("\n\n%v", errorStyle.Render(errMsg))
}
