package cli

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	parser "github.com/holeyko/make-helper/internal/make-helper/makefile/parser"
)

type Action string

const (
	typing         = Action("")
	typeSearchText = Action("typeSearchText")
	searchText     = Action("searchText")
)

type cliModel struct {
	choices []string
	numLine int
	output  string
	buffer  string
	action  Action
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
	view := fmt.Sprintf("%v to quit.\n", applyCommandStyle("Press q"))
	view += fmt.Sprintf("%v to drop output.\n\n", applyCommandStyle("Press r"))

	if model.action == typeSearchText || model.action == searchText {
		view += fmt.Sprintf("%s %s\n\n", applySearchTextTitleStyle("Search text:"), model.buffer)
	}

	view += "Makefile's targets:\n"

	for i, choice := range model.choices {
		if model.numLine == i {
			view += fmt.Sprintf("> %s", applyChosenTargetStyle(choice))
		} else {
			view += fmt.Sprintf("  %s", applyTargetStyle(choice))
		}

		view += "\n"
	}

	if model.output != "" {
		view += "\n"
		view += model.output
		view += "\n"
	}

	return view
}

func (model cliModel) getSelectedTarget() string {
	return model.choices[model.numLine]
}

func handleKeyPress(model cliModel, message tea.KeyMsg) (cliModel, tea.Cmd) {
	var cmd tea.Cmd

	switch message.String() {

	// Go to default action
	case "esc":
		model.action = typing
		model.buffer = ""

	default:
		if model.action == typing {
			model, cmd = handleTypingAction(model, message)
		} else if model.action == typeSearchText {
			model, cmd = handleTypeSearchingText(model, message)
		} else if model.action == searchText {
			model, cmd = handleSearchText(model, message)
		}
	}

	return model, cmd
}

func handleTypingAction(model cliModel, message tea.KeyMsg) (cliModel, tea.Cmd) {
	var cmd tea.Cmd
	clearBuffer := true

	switch message.String() {

	// Quit from CLI
	case "ctrl+c", "q":
		cmd = tea.Quit

	// Move cursor up
	case "up", "k":
		if model.numLine > 0 {
			model.numLine = max(0, model.numLine-countFromBuffer(model))
		}

	// Move cursor down
	case "down", "j":
		if model.numLine < len(model.choices)-1 {
			model.numLine = min(len(model.choices)-1, model.numLine+countFromBuffer(model))
		}

	case "r":
		model.output = ""

	// Execute Makefile's target
	case "enter":
		target := model.getSelectedTarget()
		output, _ := executeMakeTarget(target)

		text := output
		model.output = text

	// Searching
	case "/":
		model.action = typeSearchText
		model.buffer = ""

	default:
		model.buffer += message.String()
		clearBuffer = false
	}

	if clearBuffer {
		model.buffer = ""
	}

	return model, cmd
}

func handleTypeSearchingText(model cliModel, message tea.KeyMsg) (cliModel, tea.Cmd) {
	var cmd tea.Cmd

	switch message.String() {
	case "enter":
		model.action = searchText
		model.numLine = findNextMatchedLine(model)

	case "backspace":
		if model.buffer != "" {
			model.buffer = model.buffer[:len(model.buffer)-1]
		}

	default:
		model.buffer += message.String()
	}

	return model, cmd
}

func handleSearchText(model cliModel, message tea.KeyMsg) (cliModel, tea.Cmd) {
	var cmd tea.Cmd

	switch message.String() {
	case "n":
		model.numLine = findNextMatchedLine(model)

	case "N":
		model.numLine = findPreviousMatchedLine(model)

	case "/":
		model.action = typeSearchText
		model.buffer = ""

	case "enter":
		target := model.getSelectedTarget()
		output, _ := executeMakeTarget(target)

		text := output
		model.output = text
	}

	return model, cmd
}

func findNextMatchedLine(model cliModel) int {
	for i := model.numLine + 1; i < model.numLine+len(model.choices); i++ {
		idx := i % len(model.choices)
		choice := model.choices[idx]
		if strings.Contains(choice, model.buffer) {
			return idx
		}
	}

	return model.numLine
}

func findPreviousMatchedLine(model cliModel) int {
	for i := model.numLine - 1; i > model.numLine-len(model.choices); i-- {
		idx := i
		if idx < 0 {
			idx += len(model.choices)
		}

		choice := model.choices[idx]
		if strings.Contains(choice, model.buffer) {
			return idx
		}
	}

	return model.numLine
}

func initCliModel() *cliModel {
	targets := parseMakefileTargets()
	targetsNames := make([]string, len(targets))
	for i, target := range targets {
		targetsNames[i] = target.name
	}

	return &cliModel{
		choices: targetsNames,
		action:  typing,
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

func countFromBuffer(model cliModel) int {
	n, err := strconv.Atoi(model.buffer)
	if err != nil {
		n = 1
	}

	return n
}
