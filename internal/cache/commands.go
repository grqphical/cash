package cache

type Operation = string

const (
	OperationSet    Operation = "SET"
	OperationGet    Operation = "GET"
	OperationDelete Operation = "DELETE"
)

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
