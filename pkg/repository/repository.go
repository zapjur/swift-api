package repository

import (
	"database/sql"
)

type Repository interface {
}

type repo struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repo{db: db}
}
