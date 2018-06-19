package db

import (
	"fmt"
	"os"
	"path/filepath"

	"database/sql"
	"github.com/aemengo/snb/fs"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	root string
	db   *sqlx.DB
}

func New(rootDir string) (*DB, error) {
	if err := os.MkdirAll(rootDir, os.ModePerm); err != nil {
		return nil, err
	}

	dbPath := filepath.Join(rootDir, "snb.db")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("Error opening database as %s: %s", dbPath, err)
	}

	sqlxDB := sqlx.NewDb(db, "sqlite3")

	sqlxDB.MustExec(`
	create table if not exists objects (
	  id integer not null primary key,
	  path text not null unique,
	  sha text not null,
	  updated_at timestamp default current_timestamp not null
	);
	`)

	sqlxDB.MustExec(`
	create table if not exists steps (
	  id integer not null primary key,
	  definition text not null,
	  number integer not null unique,
	  updated_at timestamp default current_timestamp not null
	);
	`)

	return &DB{
		root: rootDir,
		db:   sqlxDB,
	}, nil
}

func (s *DB) Close() error {
	return s.db.Close()
}

func (s *DB) IsCached(step string, index int, objects []fs.Object) (bool, error) {
	var (
		stepExists     = false
		pathsFlags     = make([]bool, len(objects))
		allPathsExists = func() bool {
			if len(pathsFlags) == 0 {
				return false
			}

			for _, pflag := range pathsFlags {
				if !pflag {
					return false
				}
			}

			return true
		}
	)

	err := s.db.Get(&stepExists, "select count(*) == 1 from steps where definition = ? and number = ? limit 1", step, index)
	if err != nil {
		return false, err
	}

	for index, obj := range objects {
		err = s.db.Get(&pathsFlags[index], "select count(*) == 1 from objects where path = ? and sha = ? limit 1", obj.Path, obj.Sha)
		if err != nil {
			return false, err
		}
	}

	return stepExists && allPathsExists(), nil
}

func (s *DB) Save(step string, index int, objects []fs.Object) error {
	_, err := s.db.Exec(`
		insert or replace into steps
		 (definition, number) VALUES
		 ($1, $2)`,
		step, index,
	)

	if err != nil {
		return err
	}

	for _, obj := range objects {
		_, err := s.db.Exec(`
		insert or replace into objects
		  (path, sha) VALUES
		  ($1, $2)`,
			obj.Path, obj.Sha,
		)

		if err != nil {
			return err
		}
	}

	return nil
}
