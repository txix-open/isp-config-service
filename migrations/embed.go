package migrations

import (
	"embed"
)

var (
	//go:embed *
	Migrations embed.FS
)
