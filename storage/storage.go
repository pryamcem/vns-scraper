package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

const (
	tableSchema = ` CREATE TABLE IF NOT EXISTS test_%d (
id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
question TEXT, 
rightanswer TEXT,
UNIQUE(question, rightanswer)
);`
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

// CreateSchemaByNum creates table with specific number.
func (s *Storage) CreateTableByNum(testNum int) error {
	// Exec schema.
	_, err := s.db.Exec(fmt.Sprintf(tableSchema, testNum))
	if err != nil {
		return fmt.Errorf("Can't create tables: %w", err)
	}
	return nil
}

// Put question and rightanswer to table test_<testNum>
func (s *Storage) Put(testNum int, question, rightanswer string) error {
	q := `INSERT OR IGNORE INTO test_%d (question, rightanswer) VALUES ( ?, ?);`
	_, err := s.db.Exec(fmt.Sprintf(q, testNum), question, rightanswer)
	if err != nil {
		return err
	}
	return nil
}

// PickRightanswer by question from storage.
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
	// Query the database for all rows in the "test_11" table.
	rows, err := s.db.Query(fmt.Sprintf("SELECT question, rightanswer FROM test_%d", testNum))
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// Write each row to the output file.
	var question, rightanswer string
	for rows.Next() {
		if err := rows.Scan(&question, &rightanswer); err != nil {
			log.Fatal(err)
		}
		_, err := fmt.Fprintf(file, "Question: %s\nRightanswer: %s\n\n", question, rightanswer)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Check for any errors encountered while iterating over rows.
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	return nil
}
