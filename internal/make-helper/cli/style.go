package cli

import "github.com/charmbracelet/lipgloss"

var (
	targetStyle = lipgloss.NewStyle().
			Bold(true)

	chosenTargetStyle = lipgloss.NewStyle().
				Bold(true).
				Underline(true)

	commandStyle = lipgloss.NewStyle().
			Bold(true)
)

func applyTargetStyle(target string) string {
	return targetStyle.Render(target)
}

func applyChosenTargetStyle(chosenTarget string) string {
	return chosenTargetStyle.Render(chosenTarget)
}

func applyCommandStyle(cmd string) string {
	return commandStyle.Render(cmd)
}
