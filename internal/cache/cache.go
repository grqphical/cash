package cache

import "fmt"

type Cache struct {
	values map[string]string
}

func New() *Cache {
	return &Cache{
		values: make(map[string]string),
	}
}

func (c *Cache) runDBService(cmdChan chan Command, outputChan chan string, errChan chan DBError) {
	for {
		cmd := <-cmdChan

		switch cmd.operation {
		case OperationGet:
			if len(cmd.args) != 1 {
				errChan <- DBError{
					kind:    DBErrorInvalidRequest,
					message: "missing parameter 'key'",
				}
				break
			}

			key := cmd.args[0]

			value, exists := c.values[key]

			if !exists {
				errChan <- DBError{
					kind:    DBErrorKeyDoesNotExist,
					message: fmt.Sprintf("key '%s' does not exist", key),
				}
				break
			}

			outputChan <- value
			break
		}
	}
}

func (c *Cache) Run() (cmdChan chan Command, resultChan chan string, errChan chan DBError) {
	cmdChan = make(chan Command)
	resultChan = make(chan string)
	errChan = make(chan DBError)

	go c.runDBService(cmdChan, resultChan, errChan)

	return
}
