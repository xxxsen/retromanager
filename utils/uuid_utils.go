package utils

import "github.com/google/uuid"

func NewUUIDStr() string {
	return uuid.NewString()
}
