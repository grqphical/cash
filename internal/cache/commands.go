package cache

import (
	"errors"
	"slices"
	"strings"
)

type Operation = string

const (
	OperationSet     Operation = "SET"
	OperationGet     Operation = "GET"
	OperationDelete  Operation = "DELETE"
	OperationExpires Operation = "EXPIRES"
)

var validOperations []Operation = []Operation{OperationGet, OperationSet, OperationDelete, OperationExpires}

// A command to run on the database
type Command struct {
	operation Operation
	args      []string
}

func NewCommand(operation Operation, args []string) Command {
	return Command{
		operation,
		args,
	}
}

// Parses a string into commands it contains
func ParseCommandsFromString(commandsString string) ([]Command, error) {
	commandsStringSplit := strings.Split(commandsString, ";")
	var commands []Command = make([]Command, 0, len(commandsStringSplit))

	for _, command := range commandsStringSplit {
		command = strings.TrimSpace(command)
		if command == "" {
			continue // skip empty command fragments
		}

		args := strings.Split(command, " ")
		if len(args) == 0 {
			return nil, errors.New("empty command")
		}

		cmd := Operation(strings.ToUpper(args[0]))

		if !slices.Contains(validOperations, cmd) {
			return nil, errors.New("invalid operation")
		}

		commands = append(commands, Command{
			operation: cmd,
			args:      args[1:],
		})
	}

	return commands, nil
}

func (c Command) String() string {
	builder := strings.Builder{}

	builder.WriteString(c.operation)
	builder.WriteRune(' ')

	for _, arg := range c.args {
		builder.WriteString(arg)
		builder.WriteRune(' ')
	}

	builder.WriteRune(';')

	return builder.String()
}
