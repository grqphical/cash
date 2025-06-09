package cache

import (
	"errors"
	"slices"
	"strings"
)

type Operation = string

const (
	OperationSet    Operation = "SET"
	OperationGet    Operation = "GET"
	OperationDelete Operation = "DELETE"
)

var validOperations []Operation = []Operation{OperationGet, OperationSet, OperationDelete}

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

func ParseCommandFromString(command string) (Command, error) {
	args := strings.Split(strings.TrimSpace(command), " ")
	if len(args) == 0 {
		return Command{}, errors.New("empty command")
	}

	cmd := Operation(strings.ToUpper(args[0]))

	if !slices.Contains(validOperations, cmd) {
		return Command{}, errors.New("invalid operation")
	}

	return Command{
		operation: cmd,
		args:      args[1:],
	}, nil
}
