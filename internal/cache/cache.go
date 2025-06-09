package cache

import (
	"fmt"
	"sync"
	"time"
)

type Cache struct {
	mu          *sync.Mutex
	expirations map[string]time.Time
	values      map[string]string
}

func New() *Cache {
	return &Cache{
		values:      make(map[string]string),
		expirations: make(map[string]time.Time),
		mu:          &sync.Mutex{},
	}
}

func (c *Cache) runDBService(cmdChan chan Command, outputChan chan string, errChan chan DBError) {
	for {
		cmd := <-cmdChan

		c.mu.Lock()
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

			expiration, expirationExists := c.expirations[key]
			if expirationExists && time.Now().After(expiration) {
				errChan <- DBError{
					kind:    DBErrorExpiredKey,
					message: fmt.Sprintf("'%s' has expired", key),
				}

				delete(c.values, key)
				delete(c.expirations, key)
				break
			}

			value, exists := c.values[key]

			if !exists {
				errChan <- DBError{
					kind:    DBErrorKeyDoesNotExist,
					message: fmt.Sprintf("key '%s' does not exist", key),
				}
				break
			}

			errChan <- DBError{
				kind: DBNoError,
			}
			outputChan <- value
			break
		case OperationSet:
			if len(cmd.args) != 2 {
				errChan <- DBError{
					kind:    DBErrorInvalidRequest,
					message: "invalid number of parameters provided",
				}
				break
			}

			key := cmd.args[0]
			value := cmd.args[1]

			c.values[key] = value
			errChan <- DBError{
				kind: DBNoError,
			}
			outputChan <- "OK"
			break
		case OperationDelete:
			if len(cmd.args) != 1 {
				errChan <- DBError{
					kind:    DBErrorInvalidRequest,
					message: "missing parameter 'key'",
				}
				break
			}

			key := cmd.args[0]
			delete(c.values, key)

			errChan <- DBError{
				kind: DBNoError,
			}
			outputChan <- "OK"
		case OperationExpires:
			if len(cmd.args) != 2 {
				errChan <- DBError{
					kind:    DBErrorInvalidRequest,
					message: "invalid number of parameters provided",
				}
				break
			}

			key := cmd.args[0]
			duration, err := time.ParseDuration(cmd.args[1] + "s")
			if err != nil {
				errChan <- DBError{
					kind:    DBErrorInvalidRequest,
					message: fmt.Sprintf("invalid duration: %s", err),
				}
				break
			}

			c.expirations[key] = time.Now().Add(duration)
			errChan <- DBError{
				kind: DBNoError,
			}
			outputChan <- "OK"
		}
		c.mu.Unlock()
	}
}

func (c *Cache) runExpirationCleanup(ticker *time.Ticker) {
	defer ticker.Stop()
	for range ticker.C {
		c.mu.Lock()
		for key, expiration := range c.expirations {
			if time.Now().After(expiration) {
				delete(c.expirations, key)
				delete(c.values, key)
			}
		}
		c.mu.Unlock()
	}
}

func (c *Cache) Run() (cmdChan chan Command, resultChan chan string, errChan chan DBError) {
	cmdChan = make(chan Command)
	resultChan = make(chan string)
	errChan = make(chan DBError)

	go c.runDBService(cmdChan, resultChan, errChan)
	go c.runExpirationCleanup(time.NewTicker(time.Second))

	return
}
