package cache

import (
	"errors"
)

var (
	ErrNotFound      = errors.New("cache: key not found")
	ErrExpired       = errors.New("cache: item expired")
	ErrCacheFull     = errors.New("cache: insufficient space")
	ErrInvalidValue  = errors.New("cache: invalid value")
	ErrValueTooLarge = errors.New("cache: value too large")
	ErrTooManyKeys   = errors.New("cache: too many keys in shard")
)
