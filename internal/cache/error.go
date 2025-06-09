package cache

import "fmt"

type DBErrorType = string

const (
	DBErrorInvalidRequest  DBErrorType = "invalid request"
	DBErrorKeyDoesNotExist             = "key does not exist"
	DBNoError                          = "no error"
	DBErrorExpiredKey                  = "expired key"
)

type DBError struct {
	kind    DBErrorType
	message string
}

func (e DBError) Error() string {
	return fmt.Sprintf("%s: %s", e.kind, e.message)
}

func (e DBError) Kind() DBErrorType {
	return e.kind
}
