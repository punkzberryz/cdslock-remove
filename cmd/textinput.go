package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mitchellh/go-wordwrap"
)

type inputState string

const (
	INPUT     inputState = "input"
	CONFIRM   inputState = "confirm"
	SEARCHING inputState = "searching"
	DELETING  inputState = "deleting"
	SUMMARY   inputState = "SUMMARY"
	QUIT      inputState = "QUIT"
	ERROR     inputState = "ERROR"
)

// Styles for the TUI
var (
	titleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	// subtitleStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
	highlightStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("211"))
	warningStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
	successStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("40"))
	fileStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	confirmButtonStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Background(lipgloss.Color("205")).Padding(0, 1)
	cancelButtonStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Background(lipgloss.Color("240")).Padding(0, 1)
)

// Model represents the state of our application
type Model struct {
	textInput     textinput.Model
	files         []string
	state         inputState // "input", "confirm", "processing", "summary"
	folderPath    string
	selectedIndex int // For confirmation yes/no
	deletedCount  int
	failedFiles   []string
	skipInput     bool // Flag to skip input state if folder path is provided
	errorMessage  string
	statusMessage string
	quitMessage   string
}

// SetFolderPath sets the folder path and skips the input state
func (m *Model) SetFolderPath(path string) {
	m.folderPath = path
	m.skipInput = true
}

// InitialModel returns a new model with default values
func InitialModel() Model {
	ti := textinput.New()
	ti.Placeholder = "Enter folder path (e.g. /kang/project/)"
	ti.Focus()
	ti.Width = 60

	vp := viewport.New(80, 20)
	vp.Style = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder())
	return Model{
		textInput:     ti,
		state:         INPUT,
		selectedIndex: 0,
		skipInput:     false,
		errorMessage:  "",
		statusMessage: "",
		quitMessage:   "",
	}
}

// Define message types for our Bubble Tea program
type findFilesMsg struct{ files []string }

type deleteResultMsg struct {
	success bool
	path    string
	index   int
}

func (m Model) Init() tea.Cmd {
	// If folder path was provided as a flag, skip input state
	if m.skipInput {
		return func() tea.Msg {
			//TODO: add function to process this...
			return nil
		}
	}
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case INPUT:
			if msg.String() == "ctrl+c" {
				m.state = QUIT
				return m, tea.Quit
			}
			//user pressed Enter key
			if msg.Type == tea.KeyEnter {
				folderPath := strings.TrimSpace(m.textInput.Value())
				if folderPath == "" {
					m.state = ERROR
					m.errorMessage = "Error: No folder path provided."
					return m, nil
				}

				//Check if folder exists
				info, err := os.Stat(folderPath)
				if err != nil || !info.IsDir() {
					m.state = ERROR
					m.errorMessage = fmt.Sprintf("Error: The folder '%s' does not exist.", folderPath)
					return m, nil
				}

				m.folderPath = folderPath
				m.state = SEARCHING
				m.statusMessage = fmt.Sprintf("Searching for .cdslck files in '%s'...", folderPath)

				// Start searching for .cdslck files
				return m, func() tea.Msg {
					files := SearchFilesParallel(folderPath)
					return findFilesMsg{files: files}
				}
			}

		case CONFIRM:
			switch msg.String() {
			case "left", "h":
				m.selectedIndex = 0 // Yes
				// Update the view to reflect the selection
				m.updateConfirmView()
				return m, nil
			case "right", "l":
				m.selectedIndex = 1 // No
				// Update the view to reflect the selection
				m.updateConfirmView()
				return m, nil
			case "enter":
				if m.selectedIndex == 0 {
					// Yes selected
					m.state = DELETING
					m.deletedCount = 0
					m.statusMessage = "Deleting files...\n\n"
					return m, func() tea.Msg {
						file := m.files[0]
						err := os.Remove(file)
						return deleteResultMsg{
							success: err == nil, // or perform actual deletion logic
							path:    file,
							index:   0,
						}
					}
				} else {
					//No selected
					m.quitMessage = "Operation cancelled."
					m.state = QUIT
					return m, nil
				}
			case "q", "ctrl+c":
				return m, tea.Quit
			}
		case ERROR, SEARCHING, QUIT, SUMMARY, DELETING:
			//we wait until user hit q
			if msg.String() == "q" || msg.String() == "ctrl+c" || msg.Type == tea.KeyEnter {
				return m, tea.Quit
			}
		}

	case findFilesMsg:
		m.files = msg.files

		if len(m.files) == 0 {
			m.state = QUIT
			m.quitMessage = fmt.Sprintf("No .cdslck files found in '%s'.", m.folderPath)
			return m, nil
		}

		m.state = CONFIRM
		m.updateConfirmView()
		return m, nil

	case deleteResultMsg:
		if msg.success {
			wrappedText := wordwrap.WrapString(fmt.Sprintf("Deleted: %s\n", msg.path), 50) //incase the text is too long
			m.statusMessage += fmt.Sprintln(successStyle.Render(wrappedText))
			m.deletedCount++
		} else {
			wrappedText := wordwrap.WrapString(fmt.Sprintf("Failed to delete: %s\n", msg.path), 50) //incase the text is too long
			m.statusMessage += fmt.Sprintln(warningStyle.Render(wrappedText))
			m.failedFiles = append(m.failedFiles, msg.path)
		}
		if msg.index == len(m.files)-1 {
			//last file processed...
			m.state = SUMMARY
			summary := fmt.Sprintf("\nOperation complete. %d out of %d files were deleted.", m.deletedCount, len(m.files))
			if len(m.failedFiles) > 0 {
				summary += "\n\nFailed to delete the following files:\n"
				for _, file := range m.failedFiles {
					summary += warningStyle.Render(fmt.Sprintf("  %s\n", file))
				}
			}
			summary += "\n\nPress q to quit."
			m.statusMessage += summary
			return m, nil
		}

		return m, func() tea.Msg {
			nextfile := m.files[msg.index+1]
			err := os.Remove(nextfile)
			return deleteResultMsg{
				index:   msg.index + 1,
				success: err == nil,
				path:    nextfile,
			}
		}
	}

	// Handle text input updates during INPUT state
	if m.state == INPUT {
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}
	// Handle viewport updates
	return m, cmd
}

// updateConfirmView updates the viewport with the confirmation view
func (m *Model) updateConfirmView() {
	content := fmt.Sprintf("Found %d .cdslck files in '%s':\n\n", len(m.files), m.folderPath)
	for _, file := range m.files {
		shortenedFile := strings.TrimPrefix(file, m.folderPath)
		if !strings.HasPrefix(shortenedFile, "/") {
			shortenedFile = "/" + shortenedFile
		}
		wrappedText := wordwrap.WrapString(shortenedFile, 50) //incase the text is too long
		text := fmt.Sprintln(fileStyle.Render(wrappedText))
		content += text
	}

	content += "\n" + highlightStyle.Render("Do you want to proceed with deletion?") + "\n\n"
	var yesButton string
	var noButton string

	if m.selectedIndex == 0 {
		yesButton = confirmButtonStyle.Bold(true).Background(lipgloss.Color("205")).Render("Yes")
		noButton = cancelButtonStyle.Bold(true).Background(lipgloss.Color("240")).Render("No")
	} else {
		yesButton = confirmButtonStyle.Bold(true).Background(lipgloss.Color("240")).Render("Yes")
		noButton = cancelButtonStyle.Bold(true).Background(lipgloss.Color("205")).Render("No")
	}

	content += fmt.Sprintf("%s    %s", yesButton, noButton)
	m.statusMessage = content
}

func (m Model) View() string {
	s := titleStyle.Render("CDSLOCK File Cleanup Tool\n\n") //header
	if m.state == QUIT {
		if m.quitMessage != "" {
			s += "\n" + m.quitMessage
		}
		s += "\n See you later!\n\n"
	}
	if m.state == INPUT {
		//INPUT mode view
		s += fmt.Sprintf("Please enter the folder path where you want to search for .cdslck files:\n\n%s\n\n",
			m.textInput.View(),
		)
	}
	if m.state == SEARCHING {
		s += m.statusMessage
	}

	if m.state == CONFIRM {
		s += m.statusMessage
	}

	if m.state == DELETING {
		s += m.statusMessage
	}

	if m.state == SUMMARY {
		s += m.statusMessage
	}

	if m.state == ERROR {
		//ERROR mode view
		s += m.errorMessage + "\n\n"
		s += "\n(Press enter | q | ctrl+c to quit.)"
		return s
	}

	s += "\n(Press ctrl+c to quit.)"
	return s

}
