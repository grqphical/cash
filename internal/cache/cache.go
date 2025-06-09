package cache

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"
)

type Cache struct {
	mu               *sync.Mutex
	expirations      map[string]time.Time
	values           map[string]string
	compressionIndex map[string]interface{}
	persistanceFile  *os.File
}

func New(persistenceFileName string) *Cache {
	var persistanceFile *os.File
	var err error

	if persistenceFileName != "" {
		persistanceFile, err = os.OpenFile(persistenceFileName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			slog.Error("unable to open persistance file", "error", err)
			os.Exit(1)
		}
	}

	return &Cache{
		values:           make(map[string]string),
		compressionIndex: make(map[string]interface{}),
		expirations:      make(map[string]time.Time),
		mu:               &sync.Mutex{},
		persistanceFile:  persistanceFile,
	}

}

func (c *Cache) Get(args []string, errChan chan DBError, outputChan chan string) bool {
	if len(args) != 1 {
		errChan <- DBError{
			kind:    DBErrorInvalidRequest,
			message: "missing parameter 'key'",
		}
		return false
	}

	key := args[0]

	expiration, expirationExists := c.expirations[key]
	if expirationExists && time.Now().After(expiration) {
		errChan <- DBError{
			kind:    DBErrorExpiredKey,
			message: fmt.Sprintf("'%s' has expired", key),
		}

		delete(c.values, key)
		delete(c.expirations, key)
		return false
	}

	value, exists := c.values[key]

	if !exists {
		errChan <- DBError{
			kind:    DBErrorKeyDoesNotExist,
			message: fmt.Sprintf("key '%s' does not exist", key),
		}
		return false
	}

	errChan <- DBError{
		kind: DBNoError,
	}

	if _, ok := c.compressionIndex[key]; ok {
		buf := bytes.NewBuffer([]byte(value))
		gz, err := gzip.NewReader(buf)
		if err != nil {
			errChan <- DBError{
				kind:    DBErrorCompressionFailure,
				message: err.Error(),
			}
			return false
		}
		defer gz.Close()

		var out bytes.Buffer
		_, err = io.Copy(&out, gz)
		if err != nil {
			errChan <- DBError{
				kind:    DBErrorCompressionFailure,
				message: err.Error(),
			}
			return false
		}

		value = out.String()

	}

	outputChan <- value
	return true
}

func (c *Cache) Set(args []string, errChan chan DBError, outputChan chan string) bool {
	if len(args) < 2 || len(args) > 3 {
		errChan <- DBError{
			kind:    DBErrorInvalidRequest,
			message: "invalid number of parameters provided",
		}
		return false
	}

	key := args[0]
	var value string

	if len(args) > 3 {
		compression := args[2]
		if strings.ToUpper(compression) == "COMPRESS" {
			var buf bytes.Buffer
			gz := gzip.NewWriter(&buf)
			_, err := gz.Write([]byte(args[1]))
			if err != nil {
				errChan <- DBError{
					kind:    DBErrorCompressionFailure,
					message: err.Error(),
				}
				return false
			}

			if err := gz.Close(); err != nil {
				errChan <- DBError{
					kind:    DBErrorCompressionFailure,
					message: err.Error(),
				}
				return false

			}

			value = buf.String()
			c.compressionIndex[key] = struct{}{}
		}
	} else {
		value = args[1]
	}

	c.values[key] = value
	errChan <- DBError{
		kind: DBNoError,
	}
	outputChan <- "OK"
	return true
}

func (c *Cache) Delete(args []string, errChan chan DBError, outputChan chan string) bool {
	if len(args) != 1 {
		errChan <- DBError{
			kind:    DBErrorInvalidRequest,
			message: "missing parameter 'key'",
		}
		return false
	}

	key := args[0]
	delete(c.values, key)

	errChan <- DBError{
		kind: DBNoError,
	}
	outputChan <- "OK"
	return true
}

func (c *Cache) Expires(args []string, errChan chan DBError, outputChan chan string) bool {
	if len(args) != 2 {
		errChan <- DBError{
			kind:    DBErrorInvalidRequest,
			message: "invalid number of parameters provided",
		}
		return false
	}

	key := args[0]
	duration, err := time.ParseDuration(args[1] + "s")
	if err != nil {
		errChan <- DBError{
			kind:    DBErrorInvalidRequest,
			message: fmt.Sprintf("invalid duration: %s", err),
		}
		return false
	}

	c.expirations[key] = time.Now().Add(duration)
	errChan <- DBError{
		kind: DBNoError,
	}
	outputChan <- "OK"
	return true
}

func (c *Cache) runDBService(cmdChan chan Command, outputChan chan string, errChan chan DBError) {
	if c.persistanceFile != nil {
		c.mu.Lock()
		c.persistanceFile.Seek(0, io.SeekStart)
		persistedData, err := io.ReadAll(c.persistanceFile)
		if err != nil {
			slog.Error("unable to read persistance file", "error", err)
		} else {
			persistedCommands, err := ParseCommandsFromString(string(persistedData))
			if err != nil {
				slog.Error("unable to parse persisted commands", "error", err)
			}

			for _, cmd := range persistedCommands {
				switch cmd.operation {
				case OperationGet:
					c.Get(cmd.args, errChan, outputChan)
					break
				case OperationSet:
					c.Set(cmd.args, errChan, outputChan)
					break
				case OperationDelete:
					c.Delete(cmd.args, errChan, outputChan)
					break
				case OperationExpires:
					c.Expires(cmd.args, errChan, outputChan)
					break
				}

				err := <-errChan
				if err.kind != DBNoError {
					slog.Error("error while running persisted command", "error", err)
				}
				<-outputChan

			}
		}
		c.mu.Unlock()
	}

	for {
		cmd := <-cmdChan

		var success bool
		c.mu.Lock()
		switch cmd.operation {
		case OperationGet:
			success = c.Get(cmd.args, errChan, outputChan)
			break
		case OperationSet:
			success = c.Set(cmd.args, errChan, outputChan)
			break
		case OperationDelete:
			success = c.Delete(cmd.args, errChan, outputChan)
			break
		case OperationExpires:
			success = c.Expires(cmd.args, errChan, outputChan)
			break
		}
		c.mu.Unlock()

		if success {
			c.persistanceFile.WriteString(cmd.String())
		}
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
	cmdChan = make(chan Command, 1)
	resultChan = make(chan string, 1)
	errChan = make(chan DBError, 1)

	go c.runDBService(cmdChan, resultChan, errChan)
	go c.runExpirationCleanup(time.NewTicker(time.Second))

	return
}

func (c *Cache) Cleanup() {
	c.persistanceFile.Close()
}
