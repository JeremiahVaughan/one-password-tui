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

	if m.data.validationMsg != "" {
		display.WriteString(getErrorStyle(m.data.validationMsg))
	}
	if m.loading {
		display.WriteString(m.spinner.View())
	} else {
		switch m.data.activeView {
		case activeViewEnterPassword:
			display.WriteString(m.data.thePassword.View())
		case activeViewListItems:
			display.WriteString(m.items.View())
		case activeViewItem:
		}
	}

	return docStyle.Render(display.String())
}

func getErrorStyle(errMsg string) string {
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true).Width(80).MarginLeft(4)
	return fmt.Sprintf("\n\n%v", errorStyle.Render(errMsg))
}
