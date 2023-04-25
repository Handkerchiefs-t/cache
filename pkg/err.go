package pkg

import "errors"

var (
	ErrKeyNotFound  = errors.New("key not found")
	ErrOverCapacity = errors.New("key not found")
	ErrSetFailed    = errors.New("set failed")
)
