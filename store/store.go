package store

import (
	"os"
	"path/filepath"
	"crypto/sha1"
	"io"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"github.com/satori/go.uuid"
	"database/sql"
	"github.com/jmoiron/sqlx"
)

type Store struct {
	root string
	db *sqlx.DB
	Err error
}

func New(rootDir string) (*Store, error) {
	dbPath := filepath.Join(rootDir, "snb.db")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("Error opening database as %s: %s", dbPath, err)
	}

	if err := os.MkdirAll(filepath.Join(rootDir, "objects"), os.ModePerm); err != nil {
		return nil, err
	}

	sqlxDB := sqlx.NewDb(db, "sqlite3")

	sqlxDB.MustExec(`
	create table if not exists objects (
	  id integer not null primary key,
	  path text not null,
	  uuid text not null,
	  sha text not null,
	  updated_at timestamp default current_timestamp not null
	);
	`)

	sqlxDB.MustExec(`
	create table if not exists steps (
	  id integer not null primary key,
	  definition text not null,
	  number integer not null,
	  updated_at timestamp default current_timestamp not null
	);
	`)

	return &Store{
		root: rootDir,
		db: sqlxDB,
	}, nil
}

func (s *Store) SaveStep(step string, index int, modifiedFiles []string) error {
	for _, file := range modifiedFiles {
		err := s.SaveBlob(file)
		if err != nil {
			return err
		}
	}

	_, err := s.db.Exec(`
	insert into steps
	 (definition, number) VALUES
	 ($1, $2)`,
		step, index,
	)

	return err
}


func (s *Store) SaveBlob(srcPath string) error {
	sha, err := s.sha(srcPath)
	if err != nil {
		return err
	}

	u1 := uuid.NewV4().String()

	err = s.copyFile(srcPath, filepath.Join(s.root, "objects", u1))
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`
	insert into objects
	 (path, uuid, sha) VALUES
	 ($1, $2, $3)`,
		srcPath, u1, sha,
	)

	return err
}

func (s *Store) copyFile(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

func (s *Store) sha(srcPath string) (string, error) {
	f, err := os.Open(srcPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}