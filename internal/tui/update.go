package tui

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mitchellh/go-wordwrap"
	s "github.com/punkzberryz/cdslock-remove/internal/search"
)

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
					files := s.SearchFilesParallel(folderPath)
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
