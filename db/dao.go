package dao

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// DBConfig holds database configuration
type DBConfig struct {
	DBPath   string
	InitFile string
}

// InitializeDatabase creates and initializes the SQLite database if it doesn't exist
func InitializeDatabase(config DBConfig) (*sql.DB, error) {
	// Check if database file already exists
	dbExists := true
	if _, err := os.Stat(config.DBPath); os.IsNotExist(err) {
		dbExists = false
	}

	// Create directory if it doesn't exist
	dbDir := filepath.Dir(config.DBPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database connection
	db, err := sql.Open("sqlite3", config.DBPath+"?_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// If database didn't exist or is empty, initialize it
	if !dbExists || isDatabaseEmpty(db) {
		if err := initializeSchema(db, config.InitFile); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to initialize database schema: %w", err)
		}
		fmt.Printf("Database initialized successfully at %s\n", config.DBPath)
	} else {
		fmt.Printf("Using existing database at %s\n", config.DBPath)
	}

	return db, nil
}

// isDatabaseEmpty checks if the database has any tables
func isDatabaseEmpty(db *sql.DB) bool {
	var count int
	query := `SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'`
	err := db.QueryRow(query).Scan(&count)
	if err != nil {
		return true // Assume empty if we can't check
	}
	return count == 0
}

// initializeSchema reads and executes the SQL initialization file
func initializeSchema(db *sql.DB, initFile string) error {
	// Read the SQL file
	sqlContent, err := os.ReadFile(initFile)
	if err != nil {
		return fmt.Errorf("failed to read init file %s: %w", initFile, err)
	}

	// Execute the SQL content
	if _, err := db.Exec(string(sqlContent)); err != nil {
		return fmt.Errorf("failed to execute initialization SQL: %w", err)
	}

	return nil
}

// Helper functions for basic operations

// CreateCourse creates a new course
func CreateCourse(db *sql.DB, uuid, title, slug string) error {
	query := `INSERT INTO courses (uuid, title, slug) VALUES (?, ?, ?)`
	_, err := db.Exec(query, uuid, title, slug)
	return err
}

// CreateLesson creates a new lesson
func CreateLesson(db *sql.DB, uuid, courseUUID, title, slug, content string) error {
	query := `INSERT INTO lessons (uuid, course_uuid, title, slug, content) VALUES (?, ?, ?, ?, ?)`
	_, err := db.Exec(query, uuid, courseUUID, title, slug, content)
	return err
}

// CreateQuiz creates a quiz for a course
func CreateQuiz(db *sql.DB, courseUUID string) (int64, error) {
	query := `INSERT INTO quizzes (course_uuid) VALUES (?)`
	result, err := db.Exec(query, courseUUID)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// CreateQuestion creates a new question
func CreateQuestion(db *sql.DB, quizID int64, questionTypeID int, questionText, explanation, answerChoices, correctAnswer string) (int64, error) {
	query := `INSERT INTO questions (quiz_id, question_type_id, question_text, explanation, answer_choices, correct_answer) VALUES (?, ?, ?, ?, ?, ?)`
	result, err := db.Exec(query, quizID, questionTypeID, questionText, explanation, answerChoices, correctAnswer)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// RecordUserAnswer records a user's answer to a question
func RecordUserAnswer(db *sql.DB, questionID int64, userAnswer string, isCorrect bool) error {
	query := `INSERT INTO user_answers (question_id, user_answer, is_correct) VALUES (?, ?, ?)`
	_, err := db.Exec(query, questionID, userAnswer, isCorrect)
	return err
}

// Example usage
func main() {
	config := DBConfig{
		DBPath:   "./data/quiz.db",
		InitFile: "./init.sql",
	}

	db, err := InitializeDatabase(config)
	if err != nil {
		fmt.Printf("Error initializing database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Verify initialization
	var courseCount int
	err = db.QueryRow("SELECT COUNT(*) FROM courses").Scan(&courseCount)
	if err != nil {
		fmt.Printf("Error querying courses: %v\n", err)
		return
	}

	fmt.Printf("Database ready! Found %d courses.\n", courseCount)
}
