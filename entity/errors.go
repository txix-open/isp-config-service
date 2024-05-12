package entity

import (
	"github.com/pkg/errors"
)

var (
	ErrConfigNotFound         = errors.New("no active config")
	ErrModuleNotFound         = errors.New("module not found")
	ErrSchemaNotFound         = errors.New("config schema not found")
	ErrConfigNotFoundOrActive = errors.New("config not found or markers as active")
	ErrConfigConflictUpdate   = errors.New("config conflict update")
)
