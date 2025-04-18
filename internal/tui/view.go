package tui

import "fmt"

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
		s += fmt.Sprintf(
			"Please enter the folder path where you want to search for .cdslck files:\n\n%s\n\n",
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
