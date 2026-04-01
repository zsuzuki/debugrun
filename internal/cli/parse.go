package cli

import (
	"fmt"
	"strconv"
	"strings"
)

type Action string

const (
	ActionExec     Action = "exec"
	ActionShow     Action = "show"
	ActionList     Action = "list"
	ActionParams   Action = "params"
	ActionComplete Action = "complete"
	ActionLast     Action = "last"
	ActionEditLast Action = "edit-last"
	ActionRepeat   Action = "repeat"
	ActionHistory  Action = "history"
)

type RawParam struct {
	Name  string
	Value string
}

type Parsed struct {
	Action      Action
	ProfileName string
	ParamName   string
	Topic       string
	RawParams   []RawParam
	ExtraArgs   []string
	RepeatIndex int
}

func Parse(args []string) (*Parsed, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("usage: run <profile> [name=value ...] [-- extra args...]")
	}

	switch args[0] {
	case "list":
		return &Parsed{Action: ActionList}, nil
	case "history":
		return &Parsed{Action: ActionHistory}, nil
	case "last":
		return &Parsed{Action: ActionLast}, nil
	case "edit-last":
		return parseIndexedAction(ActionEditLast, "usage: run edit-last [n]", args[1:])
	case "repeat":
		return parseIndexedAction(ActionRepeat, "usage: run repeat [n]", args[1:])
	case "params":
		return parseProfileAction(ActionParams, args[1:])
	case "complete":
		return parseComplete(args[1:])
	case "show":
		return parseProfileAction(ActionShow, args[1:])
	case "exec":
		return parseProfileAction(ActionExec, args[1:])
	default:
		return parseProfileAction(ActionExec, args)
	}
}

func parseComplete(args []string) (*Parsed, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("usage: run complete profiles | params <profile> | values <profile> <param>")
	}

	switch args[0] {
	case "profiles":
		if len(args) != 1 {
			return nil, fmt.Errorf("usage: run complete profiles")
		}
		return &Parsed{Action: ActionComplete, Topic: "profiles"}, nil
	case "params":
		if len(args) != 2 {
			return nil, fmt.Errorf("usage: run complete params <profile>")
		}
		return &Parsed{Action: ActionComplete, Topic: "params", ProfileName: args[1]}, nil
	case "values":
		if len(args) != 3 {
			return nil, fmt.Errorf("usage: run complete values <profile> <param>")
		}
		return &Parsed{Action: ActionComplete, Topic: "values", ProfileName: args[1], ParamName: args[2]}, nil
	default:
		return nil, fmt.Errorf("unknown complete topic %q", args[0])
	}
}

func parseIndexedAction(action Action, usage string, args []string) (*Parsed, error) {
	parsed := &Parsed{Action: action, RepeatIndex: 1}
	if len(args) == 0 {
		return parsed, nil
	}
	if len(args) > 1 {
		return nil, fmt.Errorf("%s", usage)
	}
	n, err := strconv.Atoi(args[0])
	if err != nil || n < 1 {
		return nil, fmt.Errorf("history index must be a positive integer")
	}
	parsed.RepeatIndex = n
	return parsed, nil
}

func parseProfileAction(action Action, args []string) (*Parsed, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("profile is required")
	}

	parsed := &Parsed{
		Action:      action,
		ProfileName: args[0],
	}

	rest := args[1:]
	dashdash := len(rest)
	for i, token := range rest {
		if token == "--" {
			dashdash = i
			break
		}
	}

	for _, token := range rest[:dashdash] {
		name, value, ok := strings.Cut(token, "=")
		if !ok || name == "" {
			return nil, fmt.Errorf("expected name=value before --, got %q", token)
		}
		if value == "" {
			return nil, fmt.Errorf("empty value is not allowed for %q", name)
		}
		parsed.RawParams = append(parsed.RawParams, RawParam{Name: name, Value: value})
	}

	if dashdash < len(rest) {
		parsed.ExtraArgs = append(parsed.ExtraArgs, rest[dashdash+1:]...)
	}

	return parsed, nil
}
