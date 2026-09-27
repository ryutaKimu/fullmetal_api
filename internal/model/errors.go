package model

import "errors"

var (
	ErrNotFound        = errors.New("narration not found")
	ErrInvalidEpisodes = errors.New("narration episodes out of range")
)
