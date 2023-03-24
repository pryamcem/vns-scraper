package main

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

var dbschema = ` CREATE TABLE IF NOT EXISTS qa (
id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
test_num INTEGER,
question TEXT, 
rightanswer TEXT,
UNIQUE(question, rightanswer)
);`

// TODO: create table for each test or create another table which contains number of tables with tests
//CREATE TABLE IF NOT EXISTS tests (

//)`

var dbconfig = `PRAGMA foreign_keys = ON`

// New creates new SQLite storage.
func New(path string) (*Storage, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("Can't open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("Can't connect to database: %w", err)
	}
	return &Storage{db: db}, nil
}

func (s *Storage) Close() {
	s.db.Close()
}

// Init schama
func (s *Storage) Init() error {
	// Exec schema.
	_, err := s.db.Exec(dbschema)
	if err != nil {
		return fmt.Errorf("Can't create tables: %w", err)
	}
	// Exec config.
	_, err = s.db.Exec(dbconfig)
	if err != nil {
		return fmt.Errorf("Can't configure database: %w", err)
	}
	return nil
}

func (s *Storage) PutQA(data QA) error {
	_, rightanser, ok := strings.Cut(data.rightanswer, "Правильна відповідь: ")
	if !ok {
		return errors.New("Anser cut error")
	}
	fmt.Println("PUT", rightanser)
	q := `INSERT OR IGNORE INTO qa (test_num, question, rightanswer) VALUES (?, ?, ?);`
	_, err := s.db.Exec(q, data.testnum, data.question, rightanser)
	if err != nil {
		return err
	}
	return nil
}

func (s *Storage) GetRightanswer(question string) (answer string, err error) {
	q := `SELECT rightanswer FROM qa WHERE question = ? LIMIT 1;`
	err = s.db.QueryRow(q, question).Scan(&answer)
	//if err == sql.ErrNoRows {
	//return "", errors.New("No questions like that in database.")
	//}
	if err != nil && err != sql.ErrNoRows {
		return "", err
	}
	return answer, nil
}
