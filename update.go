package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type OnePasswordItem struct {
	ID                    string    `json:"id,omitempty"`
	TheTitle              string    `json:"title,omitempty"`
	Version               int       `json:"version,omitempty"`
	Vault                 Vault     `json:"vault,omitempty"`
	Category              string    `json:"category,omitempty"`
	LastEditedBy          string    `json:"last_edited_by,omitempty"`
	CreatedAt             time.Time `json:"created_at,omitempty"`
	UpdatedAt             time.Time `json:"updated_at,omitempty"`
	AdditionalInformation string    `json:"additional_information,omitempty"`
	Urls                  []Urls    `json:"urls,omitempty"`
}
type Vault struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}
type Urls struct {
	Label   string `json:"label,omitempty"`
	Primary bool   `json:"primary,omitempty"`
	Href    string `json:"href,omitempty"`
}

type OnePasswordItemDetails struct {
	ID                    string    `json:"id,omitempty"`
	Title                 string    `json:"title,omitempty"`
	Version               int       `json:"version,omitempty"`
	Vault                 Vault     `json:"vault,omitempty"`
	Category              string    `json:"category,omitempty"`
	LastEditedBy          string    `json:"last_edited_by,omitempty"`
	CreatedAt             time.Time `json:"created_at,omitempty"`
	UpdatedAt             time.Time `json:"updated_at,omitempty"`
	AdditionalInformation string    `json:"additional_information,omitempty"`
	Sections              []Section `json:"sections,omitempty"`
	Fields                []Fields  `json:"fields,omitempty"`
}

type PasswordDetails struct {
	Entropy   int    `json:"entropy,omitempty"`
	Generated bool   `json:"generated,omitempty"`
	Strength  string `json:"strength,omitempty"`
}

type Section struct {
	ID string `json:"id,omitempty"`
}

type FieldType string

const (
	FieldTypeConcealed   FieldType = "CONCEALED"
	FieldTypeString      FieldType = "STRING"
	FieldTypeEmail       FieldType = "EMAIL"
	FieldTypeUrl         FieldType = "URL"
	FieldTypeDate        FieldType = "DATE"
	FieldTypeMonthYear   FieldType = "MONTH_YEAR"
	FieldTypePhoneNumber FieldType = "PHONE"
	FieldTypeOtp         FieldType = "OTP"
	FieldTypeNa          FieldType = "N/A"
)

type Fields struct {
	ID              string          `json:"id,omitempty"`
	Type            FieldType       `json:"type,omitempty"`
	Purpose         string          `json:"purpose,omitempty"`
	Label           string          `json:"label,omitempty"`
	Value           string          `json:"value,omitempty"`
	Reference       string          `json:"reference,omitempty"`
	Entropy         float64         `json:"entropy,omitempty"`
	PasswordDetails PasswordDetails `json:"password_details,omitempty"`
	Section         Section         `json:"section,omitempty"`
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !m.loading {
			// reset any errors or validation messages on key press if not loading
			m.data.err = nil
			m.data.validationMsg = ""

			if msg.Type == tea.KeyCtrlC {
				return m, tea.Quit
			}

			switch m.data.activeView {
			case activeViewEnterPassword:
				switch msg.Type {
				case tea.KeyEnter:
					if !m.loading {
						m.loading = true
						go func() {
							var md modelData
							passwordValue := m.data.thePassword.Value()
							if passwordValue == "" {
								md.validationMsg = "must provide the password"
							} else {
								password := fmt.Sprintf("'%s'", passwordValue)
								theCommand := exec.Command("op", "signin")
								theCommand.Stdin = bytes.NewBufferString(password)
								output, err := theCommand.CombinedOutput()
								if err != nil {
									m.data.err = fmt.Errorf("error, during signin. Error: %v. Output: %s", err, output)
								} else {
									listItems, err := fetchItems()
									if err != nil {
										m.data.err = fmt.Errorf("error, when fetchItems() for Update(). Error: %v. Output: %s", err, output)
									} else {
										md.err = nil
										md.validationMsg = ""
										md.activeView = activeViewListItems
										md.items = listItems
									}
								}
							}
							loadingFinished <- md
						}()
						return m, m.spinner.Tick
					}
				}
			case activeViewListItems:
				switch msg.Type {
				case tea.KeyEnter:

				}
			case activeViewItem:
			}
		}
	case spinner.TickMsg:
		select {
		case md := <-loadingFinished:
			m.resetSpinner()
			m.loading = false
			m.data = md
			switch m.data.activeView {
			case activeViewListItems:
				itemsToSet := make([]list.Item, len(m.data.items))
				for i, it := range m.data.items {
					itemsToSet[i] = it
				}
				m.items.SetItems(itemsToSet)
			case activeViewItem:
			}
		default:
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.items.SetSize(msg.Width-h, msg.Height-v)
	case error:
		m.data.err = msg
		return m, nil
	}

	// if I don't do this down here the updates don't work properly, seems casting to type is causing an issue
	switch m.data.activeView {
	case activeViewEnterPassword:
		m.data.thePassword, cmd = m.data.thePassword.Update(msg)
	case activeViewListItems:
		m.items, cmd = m.items.Update(msg)
	case activeViewItem:
	}
	return m, cmd
}

func fetchItems() ([]OnePasswordItem, error) {
	theCommand := exec.Command("op", "item", "list", "--format", "json")
	output, err := theCommand.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error, during signin. Error: %v. Output: %s", err, output)
	}
	var onePasswordItems []OnePasswordItem
	err = json.Unmarshal(output, &onePasswordItems)
	if err != nil {
		return nil, fmt.Errorf("error, when decoding items from one password. Error: %v. Command output: %s", err, output)
	}
	return onePasswordItems, nil
}
