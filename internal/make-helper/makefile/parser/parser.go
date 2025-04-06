package parser

import (
	"errors"
	"regexp"
	"strings"
)

type Makefile struct {
	FileName  string
	Rules     RuleList
	Variables VariableList
}

type Rule struct {
	Target       string
	Dependencies []string
	Body         []string
	FileName     string
	LineNumber   int
}

type RuleList []Rule

type Variable struct {
	Name            string
	SimplyExpanded  bool
	Assignment      string
	SpecialVariable bool
	FileName        string
	LineNumber      int
}

type VariableList []Variable

var (
	reFindRule             = regexp.MustCompile("^([a-zA-Z-]+):(.*)")
	reFindRuleBody         = regexp.MustCompile("^\t+(.*)")
	reFindSimpleVariable   = regexp.MustCompile("^([a-zA-Z]+) ?:=(.*)")
	reFindExpandedVariable = regexp.MustCompile("^([a-zA-Z]+) ?=(.*)")
	reFindSpecialVariable  = regexp.MustCompile("^\\.([a-zA-Z_]+):(.*)")
)

func Parse(filepath string) (ret Makefile, err error) {

	ret.FileName = filepath
	var scanner *MakefileScanner
	scanner, err = NewMakefileScanner(filepath)
	if err != nil {
		return ret, err
	}

	for {
		switch {
		case strings.HasPrefix(scanner.Text(), "#"):
			// parse comments here, ignoring them for now
			scanner.Scan()
		case strings.HasPrefix(scanner.Text(), "."):
			if matches := reFindSpecialVariable.FindStringSubmatch(scanner.Text()); matches != nil {
				specialVar := Variable{
					Name:            strings.TrimSpace(matches[1]),
					Assignment:      strings.TrimSpace(matches[2]),
					SpecialVariable: true,
					FileName:        filepath,
					LineNumber:      scanner.LineNumber}
				ret.Variables = append(ret.Variables, specialVar)
			}
			scanner.Scan()
		default:
			ruleOrVariable, parseError := parseRuleOrVariable(scanner)
			if parseError != nil {
				return ret, parseError
			}
			switch ruleOrVariable.(type) {
			case Rule:
				rule, found := ruleOrVariable.(Rule)
				if found != true {
					return ret, errors.New("Parse error")
				}
				ret.Rules = append(ret.Rules, rule)
			case Variable:
				variable, found := ruleOrVariable.(Variable)
				if found != true {
					return ret, errors.New("Parse error")
				}
				ret.Variables = append(ret.Variables, variable)
			}
		}

		if scanner.Finished == true {
			return
		}
	}
}

func parseRuleOrVariable(scanner *MakefileScanner) (ret interface{}, err error) {
	line := scanner.Text()

	if matches := reFindRule.FindStringSubmatch(line); matches != nil {
		// we found a rule so we need to advance the scanner to figure out if
		// there is a body
		beginLineNumber := scanner.LineNumber - 1
		scanner.Scan()
		bodyMatches := reFindRuleBody.FindStringSubmatch(scanner.Text())
		ruleBody := make([]string, 0, 20)
		for bodyMatches != nil {

			ruleBody = append(ruleBody, strings.TrimSpace(bodyMatches[1]))

			// done parsing the rule body line, advance the scanner and potentially
			// go into the next loop iteration
			scanner.Scan()
			bodyMatches = reFindRuleBody.FindStringSubmatch(scanner.Text())
		}
		// trim whitespace from all dependencies
		deps := strings.Split(matches[2], " ")
		filteredDeps := make([]string, 0, cap(deps))

		for idx := range deps {
			item := strings.TrimSpace(deps[idx])
			if item != "" {
				filteredDeps = append(filteredDeps, item)
			}
		}
		ret = Rule{
			Target:       strings.TrimSpace(matches[1]),
			Dependencies: filteredDeps,
			Body:         ruleBody,
			FileName:     scanner.FileHandle.Name(),
			LineNumber:   beginLineNumber}
	} else if matches := reFindSimpleVariable.FindStringSubmatch(line); matches != nil {
		ret = Variable{
			Name:           strings.TrimSpace(matches[1]),
			Assignment:     strings.TrimSpace(matches[2]),
			SimplyExpanded: true,
			FileName:       scanner.FileHandle.Name(),
			LineNumber:     scanner.LineNumber}
		scanner.Scan()
	} else if matches := reFindExpandedVariable.FindStringSubmatch(line); matches != nil {
		ret = Variable{
			Name:           strings.TrimSpace(matches[1]),
			Assignment:     strings.TrimSpace(matches[2]),
			SimplyExpanded: false,
			FileName:       scanner.FileHandle.Name(),
			LineNumber:     scanner.LineNumber}
		scanner.Scan()
	} else {
		scanner.Scan()
	}

	return
}
