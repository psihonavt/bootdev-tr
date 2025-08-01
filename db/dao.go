package dao

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// DBConfig holds database configuration
type DBConfig struct {
	DBPath   string
	InitFile string
}

func getDefaultDB() *sql.DB {
	config := DBConfig{
		DBPath:   "data/bootdev.db",
		InitFile: "init.sql",
	}
	db, err := InitializeDatabase(config)
	if err != nil {
		panic(err)
	}
	return db
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
	}
	return db, nil
}

// ResetDatabase drops all tables and recreates them with fresh data
func ResetDatabase(config DBConfig) error {
	// Open database connection
	db, err := sql.Open("sqlite3", config.DBPath+"?_foreign_keys=on")
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Test the connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Drop all tables
	dropQueries := []string{
		"DROP TABLE IF EXISTS user_answers",
		"DROP TABLE IF EXISTS questions",
		"DROP TABLE IF EXISTS quizzes",
		"DROP TABLE IF EXISTS question_types",
	}

	for _, query := range dropQueries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to drop table: %w", err)
		}
	}

	// Recreate schema
	if err := initializeSchema(db, config.InitFile); err != nil {
		return fmt.Errorf("failed to recreate schema: %w", err)
	}

	return nil
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

// CreateQuiz creates a quiz for a course
func CreateQuiz(db *sql.DB, quiz *Quiz) error {
	if db == nil {
		db = getDefaultDB()
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	insertQuiz := `INSERT INTO quizzes (course_uuid) VALUES (?) ON CONFLICT (course_uuid) DO NOTHING`
	insertQuestion := `
	INSERT INTO questions (quiz_id, question_type_id, question_text, explanation, answer_choices, correct_answer) 
	VALUES ((select id from quizzes where course_uuid = ?), (select id from question_types where name = ?),  ?, ?, ?, ?)
	`

	_, err = tx.Exec(insertQuiz, quiz.CourseUUID)
	if err != nil {
		return err
	}

	for _, q := range quiz.Questions {
		_, err := tx.Exec(insertQuestion, quiz.CourseUUID, q.QuestionType, q.QuestionText, q.Explanation, q.AnswerChoices, q.CorrectAnswer)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

// QuizStats holds statistics for a specific quiz/course
type QuizStats struct {
	CourseUUID      string
	QuestionCount   int
	TotalAnswers    int
	CorrectAnswers  int
	CorrectnessRate float64
}

// GetQuizzesStats retrieves statistics for all quizzes
func GetQuizzesStats(db *sql.DB) ([]QuizStats, error) {
	query := `
		SELECT 
			qz.course_uuid,
			COUNT(DISTINCT q.id) as question_count,
			COUNT(ua.id) as total_answers,
			SUM(CASE WHEN ua.is_correct = 1 THEN 1 ELSE 0 END) as correct_answers
		FROM quizzes qz
		LEFT JOIN questions q ON qz.id = q.quiz_id
		LEFT JOIN user_answers ua ON q.id = ua.question_id
		GROUP BY qz.course_uuid
		ORDER BY qz.course_uuid
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying quiz stats: %w", err)
	}
	defer rows.Close()

	var stats []QuizStats
	for rows.Next() {
		var stat QuizStats
		err := rows.Scan(&stat.CourseUUID, &stat.QuestionCount, &stat.TotalAnswers, &stat.CorrectAnswers)
		if err != nil {
			return nil, fmt.Errorf("error scanning quiz stats: %w", err)
		}

		if stat.TotalAnswers > 0 {
			stat.CorrectnessRate = float64(stat.CorrectAnswers) / float64(stat.TotalAnswers) * 100
		}

		stats = append(stats, stat)
	}

	return stats, nil
}

func GetQuizStats(db *sql.DB, courseUUID string) (*QuizStats, error) {
	if db == nil {
		db = getDefaultDB()
	}
	defer db.Close()

	query := `
		SELECT 
			qz.course_uuid,
			COUNT(DISTINCT q.id) as question_count,
			COUNT(ua.id) as total_answers,
			SUM(CASE WHEN ua.is_correct = 1 THEN 1 ELSE 0 END) as correct_answers
		FROM quizzes qz
		LEFT JOIN questions q ON qz.id = q.quiz_id
		LEFT JOIN user_answers ua ON q.id = ua.question_id
		WHERE qz.course_uuid = ?
	`

	var stats QuizStats

	err := db.QueryRow(query, courseUUID).Scan(&stats.CourseUUID, &stats.QuestionCount, &stats.TotalAnswers, &stats.CorrectAnswers)
	if err != nil {
		return nil, fmt.Errorf("error querying quiz stats: %w", err)
	}

	return &stats, nil
}

// Question represents a quiz question
type Question struct {
	ID            int64
	QuizID        int64
	QuestionType  string
	QuestionText  string
	Explanation   string
	AnswerChoices string
	CorrectAnswer string
}

func shuffle[T any](target []T) {
	rand.Shuffle(len(target), func(i, j int) {
		target[i], target[j] = target[j], target[i]
	})
}

func (q *Question) GetAnswerChoices() []string {
	choices := []string{}
	if q.QuestionType != "multiple_choice" {
		return choices
	} else {
		json.Unmarshal([]byte(q.AnswerChoices), &choices)
	}
	return choices
}

func (q *Question) IsCorrectAnswer(answer string) bool {
	return strings.TrimSpace(answer) == strings.TrimSpace(q.CorrectAnswer)
}

type Quiz struct {
	CourseUUID string
	Questions  []Question
}

func (q *Quiz) ShuffleQuestions() {
	shuffle(q.Questions)
	
	// Also shuffle answers within each question
	for i := range q.Questions {
		question := &q.Questions[i]
		if question.QuestionType == "multiple_choice" {
			choices := []string{}
			json.Unmarshal([]byte(question.AnswerChoices), &choices)
			if len(choices) > 1 {
				shuffle(choices)
				choicesJSON, _ := json.Marshal(choices)
				question.AnswerChoices = string(choicesJSON)
			}
		}
	}
}

func GetQuiz(db *sql.DB, courseUUID string) (*Quiz, error) {
	if db == nil {
		db = getDefaultDB()
	}
	defer db.Close()

	query := `
		SELECT q.id, q.quiz_id, qt.name, q.question_text, q.explanation, q.answer_choices, q.correct_answer
		FROM questions q
		JOIN quizzes qz ON q.quiz_id = qz.id
		JOIN question_types qt ON q.question_type_id = qt.id
		WHERE qz.course_uuid = ? 
		  AND qt.name = 'multiple_choice'
		  AND q.id NOT IN (
			  SELECT question_id FROM user_answers WHERE question_id = q.id AND is_correct = true
		  )
		ORDER BY q.id ASC
	`

	questions := []Question{}

	rows, err := db.Query(query, courseUUID)
	if err != nil {
		return nil, fmt.Errorf("error getting quiz: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var question Question
		err := rows.Scan(
			&question.ID,
			&question.QuizID,
			&question.QuestionType,
			&question.QuestionText,
			&question.Explanation,
			&question.AnswerChoices,
			&question.CorrectAnswer,
		)
		if err != nil {
			return nil, fmt.Errorf("error getting a quiz question: %w", err)
		}
		questions = append(questions, question)
	}

	if len(questions) > 0 {
		return &Quiz{CourseUUID: courseUUID, Questions: questions}, nil
	} else {
		return nil, nil
	}
}

func MarkQuestionAnswered(db *sql.DB, questionId int, answer string, isCorrect bool) error {
	stmt := `
	insert into user_answers(question_id, user_answer, is_correct) values (?, ?, ?) 
	`

	if db == nil {
		db = getDefaultDB()
	}
	defer db.Close()

	_, err := db.Exec(stmt, questionId, answer, isCorrect)
	if err != nil {
		panic(fmt.Sprintf("%w", questionId, err))
	}
	return err
}
