package storage

import "errors"

var ErrKeyNotFound = errors.New("key not found")
var ErrEmptyKey = errors.New("key cannot be empty")

type Storage interface {
	Set(key, value string) error
	Get(key string) (string, error)
	Delete(key string) error
	Exists(key string) bool
}