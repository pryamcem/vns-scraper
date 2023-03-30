package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

const (
	dbschema = ` CREATE TABLE IF NOT EXISTS test_%d (
id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
question TEXT, 
rightanswer TEXT,
UNIQUE(question, rightanswer)
);`

	dbconfig = `PRAGMA foreign_keys = ON`
)

var ErrNoSavedQA = errors.New("This question is not exists in databese")

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

// Init schema
func (s *Storage) InitByNum(testNum int) error {
	// Exec schema.
	_, err := s.db.Exec(fmt.Sprintf(dbschema, testNum))
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

func (s *Storage) PutQA(testNum int, data QA) error {
	_, rightanser, ok := strings.Cut(data.rightanswer, "Правильна відповідь: ")
	if !ok {
		return errors.New("Anser cut error")
	}
	fmt.Println("PUT", rightanser)
	q := `INSERT OR IGNORE INTO test_%d (question, rightanswer) VALUES ( ?, ?);`
	_, err := s.db.Exec(fmt.Sprintf(q, testNum), data.question, rightanser)
	if err != nil {
		return err
	}
	return nil
}

func (s *Storage) PickRightanswer(testNum int, question string) (answer string, err error) {
	q := `SELECT rightanswer FROM test_%d WHERE question = ? LIMIT 1;`
	err = s.db.QueryRow(fmt.Sprintf(q, testNum), question).Scan(&answer)
	//if err == sql.ErrNoRows {
	//return "", ErrNoSavedQA
	//}
	if err != nil && err != sql.ErrNoRows {
		return "", err
	}
	return answer, nil
}

func (s *Storage) ParseToFile(testNum int) error {
	//Create file
	path := fmt.Sprintf("answers/test_%d.txt", testNum)
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("Cant't create file: %w", err)
	}
	defer file.Close()

	//Get data from database
	qa := QA{}
	q := `SELECT question, rightanswer FROM test_%d;`
	rows, err := s.db.Query(fmt.Sprintf(q, testNum))
	if err != nil {
		return fmt.Errorf("Cant't get data from database: %w", err)
	}
	defer rows.Close()

	//Write row by row query resoult to file
	for rows.Next() {
		//Scan row to structure
		err = rows.Scan(&qa.question, &qa.rightanswer)
		if err != nil {
			return fmt.Errorf("Dow scanning error: %w", err)
		}

		//Format and write data
		data := fmt.Sprintf("Question: %s\nRightanswer: %s\n\n", qa.question, qa.rightanswer)
		fmt.Println(data)
		_, err = file.WriteString(data)
		if err != nil {
			return fmt.Errorf("Cant't write string to file: %w", err)
		}
	}
	return nil
}
