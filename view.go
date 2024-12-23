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

	// titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true).Background(lipgloss.Color("4")).PaddingTop(1).PaddingBottom(1).PaddingLeft(3).PaddingRight(3)
	// title := "One Password"
	// display.WriteString(titleStyle.Render(title))
	// display.WriteRune('\n')
	// display.WriteRune('\n')
	// display.WriteRune('\n')

	if m.loading {
		display.WriteString(m.spinner.View())
	} else {
		switch m.data.activeView {
		case activeViewEnterPassword:
			password := m.data.thePassword.Value()
			maskedPassword := strings.Repeat("●", len(password))
			view := fmt.Sprintf(
				"Enter your password:\n\n %s\n\n(press enter to submit)",
				maskedPassword,
			)
			display.WriteString(view)
		case activeViewListItems:
			display.WriteString(m.items.View())
		case activeViewItem:
			display.WriteString(m.itemDetails.View())
			// timers start off as running even if they have not started, so using
			// a nil check to get around this issue
			if m.clipboardLifeMeter != nil && m.clipboardLifeMeter.Running() {
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
				display.WriteString(fmt.Sprintf("Clipboard cleanup in: %s", timerStyle.Render(m.clipboardLifeMeter.View())))
			}
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
