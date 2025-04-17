package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles for the TUI
var (
	titleStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	subtitleStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
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
	viewport      viewport.Model
	files         []string
	state         string // "input", "confirm", "processing", "summary"
	folderPath    string
	selectedIndex int // For confirmation yes/no
	deletedCount  int
	failedFiles   []string
	skipInput     bool // Flag to skip input state if folder path is provided
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
		viewport:      vp,
		state:         "input",
		selectedIndex: 0,
		skipInput:     false,
	}
}

// Define message types for our Bubble Tea program
type findFilesMsg struct{ files []string }

type deleteResultMsg struct {
	success bool
	path    string
}

func (m Model) Init() tea.Cmd {
	// If folder path was provided as a flag, skip input state
	if m.skipInput {
		return func() tea.Msg {
			//do-something...
			return nil
		}
	}
	//cursor blinking.
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case "input":
			//user pressed Enter key
			if msg.Type == tea.KeyEnter {
				folderPath := strings.TrimSpace(m.textInput.Value())
				if folderPath == "" {
					m.viewport.SetContent(
						warningStyle.Render("Error: No folder path provided."))
					return m, nil
				}

				//Check if folder exists
				info, err := os.Stat(folderPath)
				if err != nil || !info.IsDir() {
					m.viewport.SetContent(
						warningStyle.Render(
							fmt.Sprintf("Error: The folder '%s' does not exist.", folderPath)))
				}

				m.folderPath = folderPath
				m.state = "processing"
				m.viewport.SetContent(
					subtitleStyle.Render(
						fmt.Sprintf("Searching for .cdslck files in '%s'...", folderPath)))

				// Start searching for .cdslck files
				return m, func() tea.Msg {
					var files []string
					filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
						if err != nil {
							return nil
						}
						if !info.IsDir() && strings.HasSuffix(path, ".cdslck") {
							files = append(files, path)
						}
						return nil
					})
					return findFilesMsg{files: files}
				}
			}

		case "confirm":
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
					m.state = "deleting"
					content := "Deleting files...\n\n"
					m.viewport.SetContent(content)

					// Return a command to delete the files
					return m, func() tea.Msg {
						m.deletedCount = 0
						for _, file := range m.files {
							err := os.Remove(file)
							if err != nil {
								m.failedFiles = append(m.failedFiles, file)
								return deleteResultMsg{success: false, path: file}
							} else {
								m.deletedCount++
								return deleteResultMsg{success: true, path: file}
							}
						}
						m.state = "summary"
						return nil
					}
				} else {
					//No selected
					content := "Operation cancelled."
					m.viewport.SetContent(content)
					return m, tea.Quit
				}
			case "q", "ctrl+c":
				return m, tea.Quit
			}
		case "summary", "deleting":
			if msg.String() == "q" || msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
		}

	case findFilesMsg:
		m.files = msg.files

		if len(m.files) == 0 {
			m.viewport.SetContent(subtitleStyle.Render(fmt.Sprintf("No .cdslck files found in '%s'.", m.folderPath)))
			return m, tea.Quit
		}

		m.state = "confirm"
		m.updateConfirmView()
		return m, nil

	case deleteResultMsg:
		content := m.viewport.View()
		if msg.success {
			content += successStyle.Render(fmt.Sprintf("Deleted: %s\n", msg.path))
		} else {
			content += warningStyle.Render(fmt.Sprintf("Failed to delete: %s\n", msg.path))
		}
		m.viewport.SetContent(content)

		// If all files have been processed, show the summary
		if m.deletedCount+len(m.failedFiles) >= len(m.files) {
			m.state = "summary"
			summary := fmt.Sprintf("\nOperation complete. %d out of %d files were deleted.",
				m.deletedCount, len(m.files))
			if len(m.failedFiles) > 0 {
				summary += "\n\nFailed to delete the following files:\n"
				for _, file := range m.failedFiles {
					summary += warningStyle.Render(fmt.Sprintf("  %s\n", file))
				}
			}
			summary += "\n\nPress q to quit."
			m.viewport.SetContent(m.viewport.View() + summary)
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 4

	}

	// Handle text input updates
	if m.state == "input" {
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}
	// Handle viewport updates
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// updateConfirmView updates the viewport with the confirmation view
func (m *Model) updateConfirmView() {
	content := fmt.Sprintf("Found %d .cdslck files in '%s':\n\n", len(m.files), m.folderPath)
	for _, file := range m.files {
		content += fileStyle.Render(fmt.Sprintf("  %s\n", file))
	}

	content += "\n" + highlightStyle.Render("Do you want to proceed with deletion?") + "\n\n"
	yesButton := confirmButtonStyle.Render("Yes")
	noButton := cancelButtonStyle.Render("No")

	if m.selectedIndex == 0 {
		yesButton = confirmButtonStyle.Bold(true).Render("Yes")
	} else {
		noButton = cancelButtonStyle.Bold(true).Render("No")
	}

	content += fmt.Sprintf("%s    %s", yesButton, noButton)
	m.viewport.SetContent(content)
}

func (m Model) View() string {
	if m.state == "input" {
		return fmt.Sprintf(
			"\n%s\n\n%s\n\n%s\n\n",
			titleStyle.Render("CDSLCK File Cleanup Tool"),
			subtitleStyle.Render("Please enter the folder path where you want to search for .cdslck files:"),
			m.textInput.View(),
		)
	}

	return fmt.Sprintf(
		"\n%s\n\n%s\n",
		titleStyle.Render("CDSLCK File Cleanup Tool"),
		m.viewport.View(),
	)
}
