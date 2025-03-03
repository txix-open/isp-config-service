package repository

import "isp-config-service/service/rqlite/db"

type Variable struct {
	db db.DB
}

func NewVariable(db db.DB) Variable {
	return Variable{
		db: db,
	}
}
