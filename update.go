package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/timer"

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
	Fields                []Field   `json:"fields,omitempty"`
	Files                 []File    `json:"files"`
}

type Field struct {
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

func (f Field) Title() string {
	return f.Label
}
func (f Field) Description() string {
	switch f.Type {
	case FieldTypeConcealed, FieldTypeOtp:
		return strings.Repeat("*", len(f.Value))
	}
	return f.Value
}
func (f Field) FilterValue() string {
	return f.Label
}

type File struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Size        int    `json:"size"`
	ContentPath string `json:"content_path"`
}

func (f File) Title() string {
	return f.Name
}
func (f File) Description() string {
	return fmt.Sprintf("file size: %s", formatFileSize(f.Size))
}
func (f File) FilterValue() string {
	return f.Name
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

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !m.loading {
			// reset any errors or validation messages on key press if not loading
			m.data.err = nil
			m.data.validationMsg = ""

			switch msg.Type {
			case tea.KeyCtrlC:
				return m, tea.Quit
			}

			switch m.data.activeView {
			case activeViewEnterPassword:
				switch msg.Type {
				case tea.KeyEnter:
					if !m.loading {
						passwordValue := m.data.thePassword.Value()
						if passwordValue == "" {
							m.data.validationMsg = "must provide the password"
						} else {
							m.loading = true
							md := m.data
							go func() {
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
								loadingFinished <- md
							}()
						}
						return m, m.spinner.Tick
					}
				}
			case activeViewListItems:
				switch msg.Type {
				case tea.KeyEnter:
					if m.items.FilterState() != list.Filtering {
						if !m.loading {
							theItem, ok := m.items.SelectedItem().(OnePasswordItem)
							if !ok {
								m.data.err = fmt.Errorf("error, invalid one password item type selected. Selected: %v", m.items.SelectedItem())
							}
							// todo figure out how to fetch files, secret files are not showing up
							theCommand := exec.Command("op", "item", "get", theItem.ID, "--format", "json")
							m.loading = true
							md := m.data
							go func() {
								var output []byte
								output, md.err = theCommand.CombinedOutput()
								if md.err != nil {
									md.err = fmt.Errorf("error, during signin. Error: %v. Output: %s", md.err, output)
								} else {
									var theItemDetails OnePasswordItemDetails
									md.err = json.Unmarshal(output, &theItemDetails)
									if md.err != nil {
										md.err = fmt.Errorf("error, when decoding one password item details. Error: %v", md.err)
									} else {
										md.err = nil
										md.validationMsg = ""
										md.activeView = activeViewItem
										md.selectedItem = theItemDetails
										md.fileDownloaded = false
									}
								}
								loadingFinished <- md
							}()
							return m, m.spinner.Tick
						}
					}
				}
			case activeViewItem:
				switch msg.Type {
				case tea.KeyEsc:
					m.data.activeView = activeViewListItems
				case tea.KeyEnter:
					if m.itemDetails.FilterState() != list.Filtering {
						if !m.loading {
							selectedField, ok := m.itemDetails.SelectedItem().(Field)
							if !ok {
								var selectedFile File
								selectedFile, ok = m.itemDetails.SelectedItem().(File)
								if !ok {
									m.data.err = errors.New("error, when casting selected item field")
								} else {
									var targetDir string
									targetDir, m.data.err = getDownloadTargetDirectory()
									if m.data.err != nil {
										m.data.err = fmt.Errorf("error, when getDownloadTargetDirectory() for Update(). Error: %v", m.data.err)
									} else {
										m.downloadTarget = fmt.Sprintf("%s/%s", targetDir, selectedFile.Name)
										m.loading = true
										md := m.data
										go func() {
											md.err = downloadFileToDownloads(
												m.downloadTarget,
												m.data.selectedItem,
												selectedFile,
											)
											if md.err != nil {
												md.err = fmt.Errorf("error, when downloadFileToDownloads() for Update(). Error: %v", md.err)
											} else {
												md.fileDownloaded = true
											}
											loadingFinished <- md
										}()
										return m, m.spinner.Tick
									}
								}
							} else {
								m.data.fileDownloaded = false
								m.downloadTarget = ""
								m.data.err = copyPasswordToClipboard(selectedField.Type, selectedField.Value)
								if m.data.err != nil {
									m.data.err = fmt.Errorf("error, when copyPasswordToClipboard() for Update(). Error: %v", m.data.err)
								} else {
									newTimer := timer.NewWithInterval(clipboardLifeInSeconds*time.Second, time.Millisecond)
									m.clipboardLifeMeter = &newTimer
									cmd = m.clipboardLifeMeter.Init()
								}
							}
						}
					}
				}
			}
		}
	case timer.TickMsg:
		var theTimer timer.Model
		theTimer, cmd = m.clipboardLifeMeter.Update(msg)
		m.clipboardLifeMeter = &theTimer
		return m, cmd
	case spinner.TickMsg:
		select {
		case md := <-loadingFinished:
			m.resetSpinner()
			m.loading = false
			m.data = md
			if m.data.activeView == "" {
				m.data.err = errors.New("error, activeView is not set during spinner.TickMsg evaluation")
			}
			switch m.data.activeView {
			case activeViewListItems:
				itemsToSet := make([]list.Item, len(m.data.items))
				for i, it := range m.data.items {
					itemsToSet[i] = it
				}
				m.items.SetItems(itemsToSet)
			case activeViewItem:
				var itemsToSet []list.Item
				for _, f := range md.selectedItem.Fields {
					if f.Value != "" {
						itemsToSet = append(itemsToSet, f)
					}
				}
				for _, f := range md.selectedItem.Files {
					itemsToSet = append(itemsToSet, f)
				}
				m.itemDetails.Title = md.selectedItem.Title
				m.itemDetails.SetItems(itemsToSet)
			}
		default:
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.items.SetSize(msg.Width-h, msg.Height-v)
		m.itemDetails.SetSize(msg.Width-h, msg.Height-v)
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
		var cmd2 tea.Cmd
		var cmd3 tea.Cmd
		m.itemDetails, cmd2 = m.itemDetails.Update(msg)
		if m.clipboardLifeMeter != nil {
			var theTimer timer.Model
			theTimer, cmd3 = m.clipboardLifeMeter.Update(msg)
			m.clipboardLifeMeter = &theTimer
		}
		cmd = tea.Batch(cmd, cmd2, cmd3)
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

func copyPasswordToClipboard(theType FieldType, password string) error {
	err := clipboard.WriteAll(password)
	if err != nil {
		return fmt.Errorf("failed to copy password to clipboard: %v", err)
	}
	switch theType {
	case FieldTypeConcealed, FieldTypeOtp:
		time.AfterFunc(clipboardLifeInSeconds*time.Second, func() {
			err := clipboard.WriteAll("")
			if err != nil {
				log.Fatalf("error, failed to clear clipboard:", err)
			}
		})
	}
	return nil
}

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*1, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func formatFileSize(size int) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	} else if size < 1024*1024 {
		return fmt.Sprintf("%.2f KB", float64(size)/1024)
	} else if size < 1024*1024*1024 {
		return fmt.Sprintf("%.2f MB", float64(size)/(1024*1024))
	} else {
		return fmt.Sprintf("%.2f GB", float64(size)/(1024*1024*1024))
	}
}

func downloadFileToDownloads(fileTarget string, itemDetails OnePasswordItemDetails, theFile File) error {
	fileReference := fmt.Sprintf("op://%s/%s/%s", itemDetails.Vault.ID, itemDetails.ID, theFile.ID)
	theCommand := exec.Command("op", "read", fileReference)
	var err error
	theCommand.Dir, err = getDownloadTargetDirectory()
	if err != nil {
		return fmt.Errorf("error, getDownloadTargetDirectory() for downloadFileToDownloads(). Error: %v", err)
	}
	var output []byte
	output, err = theCommand.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error, when running the read file command for downloadFileToDownloads(). Error: %v. Output: %s", err, output)
	}
	err = os.WriteFile(fileTarget, output, 0600)
	if err != nil {
		return fmt.Errorf("error, when writing the file for downloadFileToDownloads(). Error: %v", err)
	}
	return nil
}

func getDownloadTargetDirectory() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error, when getting the users home directory for getDownloadTarget(). Error: %v", err)
	}
	return fmt.Sprintf("%s/Downloads", homeDir), nil
}
