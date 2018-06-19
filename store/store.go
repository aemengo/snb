package store

import (
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"database/sql"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"strings"
)

type Store struct {
	root string
	db   *sqlx.DB
}

type Object struct {
	Path string
	Sha  string
}

func New(rootDir string) (*Store, error) {
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

	return &Store{
		root: rootDir,
		db:   sqlxDB,
	}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) AnalyzeStep(step string, index int, paths []string) ([]Object, bool, error) {
	var (
		stepExists     = false
		pathsFlags     = make([]bool, len(paths))
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
		return nil, false, err
	}

	var objects []Object

	for index, path := range paths {
		sha, err := s.getSha(path)
		if err != nil {
			return nil, false, err
		}

		err = s.db.Get(&pathsFlags[index], "select count(*) == 1 from objects where path = ? and sha = ? limit 1", path, sha)
		if err != nil {
			return nil, false, err
		}

		objects = append(objects, Object{
			Path: path,
			Sha:  sha,
		})
	}

	return objects, allPathsExists(), nil
}

//func (s *Store) SaveStep(step string, index int, modifiedFiles []string) error {
//	for _, file := range modifiedFiles {
//		err := s.SaveBlob(file)
//		if err != nil {
//			return err
//		}
//	}
//
//	_, err := s.db.Exec(`
//	insert into steps
//	 (definition, number) VALUES
//	 ($1, $2)`,
//		step, index,
//	)
//
//	return err
//}

//func (s *Store) SaveBlob(srcPath string) error {
//	sha, err := s.sha(srcPath)
//	if err != nil {
//		return err
//	}
//
//	u1 := uuid.NewV4().String()
//
//	err = s.copyFile(srcPath, filepath.Join(s.root, "objects", u1))
//	if err != nil {
//		return err
//	}
//
//	_, err = s.db.Exec(`
//	insert into objects
//	 (path, uuid, sha) VALUES
//	 ($1, $2, $3)`,
//		srcPath, u1, sha,
//	)
//
//	return err
//}

//func (s *Store) copyFile(src, dest string) error {
//	in, err := os.Open(src)
//	if err != nil {
//		return err
//	}
//	defer in.Close()
//
//	out, err := os.Create(dest)
//	if err != nil {
//		return err
//	}
//	defer out.Close()
//
//	if _, err := io.Copy(out, in); err != nil {
//		return err
//	}
//	return out.Sync()
//}

func (s *Store) getSha(srcPath string) (string, error) {
	info, err := os.Stat(srcPath)
	if err != nil {
		return "", err
	}

	if info.IsDir() {
		return s.shaDir(srcPath)
	} else {
		return s.shaFile(srcPath)
	}
}

func (s *Store) shaDir(srcPath string) (string, error) {
	h := sha1.New()

	filepath.Walk(srcPath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() || isHiddenFile(path) {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(h, f)
		return err
	})

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func (s *Store) shaFile(srcPath string) (string, error) {
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

func isHiddenFile(path string) bool {
	names := strings.Split(path, string(os.PathSeparator))
	for _, element := range names {
		if strings.HasPrefix(element, ".") {
			return true
		}
	}

	return false
}
