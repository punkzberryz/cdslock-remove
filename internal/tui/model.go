package tui

import (
	"github.com/charmbracelet/bubbles/spinner"
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
	spinner       spinner.Model //Spinner ui
	spinnerFrame  int
}

// InitialModel returns a new model with default values
func InitialModel() Model {
	s := spinner.New()
	s.Spinner = spinner.Monkey
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
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
		spinner:       s,
		spinnerFrame:  0,
	}
}

// Define message types for our Bubble Tea program
type findFilesMsg struct{ files []string }

type deleteResultMsg struct {
	success bool
	path    string
	index   int
}

type searchImmediatelyMsg struct{}

func (m Model) Init() tea.Cmd {
	// If folder path was provided as a flag, skip input state
	if m.skipInput {
		return tea.Batch(
			m.spinner.Tick,
			func() tea.Msg {
				return searchImmediatelyMsg{}
			},
		)
	}
	// return textinput.Blink
	return tea.Batch(
		textinput.Blink,
		m.spinner.Tick, //Start the spinner ticking
	)
}

// SetFolderPath sets the folder path and skips the input state
func (m *Model) SetFolderPath(path string) {
	m.folderPath = path
	m.skipInput = true
}

// Set default folder, we will not skip input but have input filled by default
func (m *Model) SetDefaultFolderPath(path string) {
	m.folderPath = path
	m.skipInput = false
	m.textInput.SetValue(path)
}
