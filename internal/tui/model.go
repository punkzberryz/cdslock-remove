package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	confirmButtonStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("255")).
				Background(lipgloss.Color("205")).
				Padding(0, 1)
	cancelButtonStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("255")).
				Background(lipgloss.Color("240")).
				Padding(0, 1)
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

// SetFolderPath sets the folder path and skips the input state
func (m *Model) SetFolderPath(path string) {
	m.folderPath = path
	m.skipInput = true
}
