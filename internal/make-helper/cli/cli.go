package cli

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	parser "github.com/holeyko/make-helper/internal/make-helper/makefile/parser"
)

type cliModel struct {
	choices []string
	numLine int
	output  string
}

type MakefileTarget struct {
	name string
}

func RunCLI() {
	checkMakefileExists()

	p := tea.NewProgram(initCliModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Exception during run: %v", err)
		os.Exit(1)
	}
}

const (
	makefileName = "Makefile"
	makeCommand  = "make"
)

func checkMakefileExists() {
	_, err := os.Stat(makefileName)
	if err == nil {
		return // File exists
	}

	if os.IsNotExist(err) {
		log.Fatalf("Can't find Makefile here")
	}
}

func (model cliModel) Init() tea.Cmd {
	return nil // Nothing need to Init
}

func (model cliModel) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	var updatedModel cliModel
	var cmd tea.Cmd

	switch typedMessage := message.(type) {
	case tea.KeyMsg:
		updatedModel, cmd = handleKeyPress(model, typedMessage)
	default:
		updatedModel, cmd = model, nil
	}

	return updatedModel, cmd
}

func (model cliModel) View() string {
	view := "\nPress q to quit.\n"
	view += "Press r to drop output.\n\n"

	view += "Makefile's targets:\n"

	for i, choice := range model.choices {
		cursor := " " // no cursor
		if model.numLine == i {
			cursor = ">"
		}

		view += fmt.Sprintf("%s [%v] %s\n", cursor, i+1, choice)
	}

	if model.output != "" {
		view += "\nOutput target:\n"
		view += model.output
		view += "\n"
	}

	return view
}

func (model cliModel) getSelectedTarget() string {
	return model.choices[model.numLine]
}

func handleKeyPress(model cliModel, message tea.KeyMsg) (cliModel, tea.Cmd) {
	switch message.String() {

	// Quit from CLI
	case "ctrl+c", "q":
		return model, tea.Quit

	// Move cursor up
	case "up", "k":
		if model.numLine > 0 {
			model.numLine--
		}

	// Move cursor down
	case "down", "j":
		if model.numLine < len(model.choices)-1 {
			model.numLine++
		}

	case "r":
		model.output = ""

	// Execute Makefile's target
	case "enter":
		target := model.getSelectedTarget()
		output, _ := executeMakeTarget(target)

		text := output
		// if err != nil {
		// 	text = err.Error()
		// }

		model.output = text
	}

	return model, nil
}

func initCliModel() *cliModel {
	targets := parseMakefileTargets()
	targetsNames := make([]string, len(targets))
	for i, target := range targets {
		targetsNames[i] = target.name
	}

	return &cliModel{
		choices: targetsNames,
	}
}

func parseMakefileTargets() []MakefileTarget {
	makefileInfo, err := parser.Parse(makefileName)
	if err != nil {
		log.Fatalf("Can't parse Makefile due to error: %v", err)
	}

	if len(makefileInfo.Rules) == 0 {
		log.Fatalln("Makefile doesn't contain any targets")
	}

	targets := make([]MakefileTarget, len(makefileInfo.Rules))
	for i, rule := range makefileInfo.Rules {
		targets[i] = MakefileTarget{name: rule.Target}
	}

	return targets
}

func executeMakeTarget(target string) (string, error) {
	cmd := exec.Command(makeCommand, target)
	output, err := cmd.CombinedOutput()

	return string(output), err
}
