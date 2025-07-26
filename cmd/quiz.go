package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	api "github.com/bootdotdev/bootdev/client"
	dao "github.com/bootdotdev/bootdev/db"
)

func getCourseContent(uuid string) {
	course, err := api.FetchCourseAndLessons(uuid)
	if err != nil {
		fmt.Printf("Error fetching course content %v", err)
		return
	}
	fmt.Printf("Title: %s\n", course.Title)
	fmt.Printf("Lessons: %d\n", len(course.GetLessons()))
	for _, l := range course.GetLessons() {
		fmt.Printf("Lessons %p %s: %d\n", l, l.Slug, len(l.Content))
	}
}

func getCompletedCourses() {
	if courses, err := api.GetUserCourses(); err != nil {
		fmt.Printf("Failed to fetch user's courses %s", err)
	} else {
		for _, course := range courses {
			if course.IsCompleted() {
				fmt.Printf("%s (%s); (Completed At %s)\n", course.Title, course.UUID, course.CompletedAt)
			}
		}
	}
}

var quizCmd = &cobra.Command{
	Use:   "quiz",
	Short: "quiz related commands",
}

var myCompletedCourses = &cobra.Command{
	Use:          "completed_courses",
	Short:        "Displays my completed courses",
	PreRun:       requireAuth,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		getCompletedCourses()
		return nil
	},
}

var downloadCourseContent = &cobra.Command{
	Use:          "download_course",
	Short:        "Downlaods all lessons from a course and stores its content in the DB",
	PreRun:       requireAuth,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			fmt.Println("Course UUID is required. Use `quiz completed_courses` to list your completed courses")
			return nil
		}
		courseUUID := args[0]
		getCourseContent(courseUUID)
		return nil
	},
}

var startQuizCmd = &cobra.Command{
	Use:   "start <COURSE_UUID>",
	Short: "Start an interactive quiz for a course",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		courseUUID := args[0]
		startQuiz(courseUUID)
		return nil
	},
}

func startQuiz(courseUUID string) {
	config := dao.DBConfig{
		InitFile: "db/init.sql",
		DBPath:   "data/bootdev.db",
	}
	db, err := dao.InitializeDatabase(config)
	if err != nil {
		fmt.Printf("Error connecting to database: %v\n", err)
		return
	}
	defer db.Close()

	// Get next unanswered multiple choice question
	question, err := dao.GetNextMultipleChoiceQuestion(db, courseUUID)
	if err != nil {
		fmt.Printf("Error getting question: %v\n", err)
		return
	}
	if question == nil {
		fmt.Println("No more questions available for this course!")
		return
	}

	// Parse answer choices
	var choices []string
	err = json.Unmarshal([]byte(question.AnswerChoices), &choices)
	if err != nil {
		fmt.Printf("Error parsing answer choices: %v\n", err)
		return
	}

	// Display question
	fmt.Printf("\n%s\n\n", question.QuestionText)
	for i, choice := range choices {
		fmt.Printf("%d. %s\n", i+1, choice)
	}

	// Get user input
	fmt.Print("\nSelect your answer (1-4): ")
	var input string
	fmt.Scanln(&input)

	// Validate input
	choiceNum, err := strconv.Atoi(input)
	if err != nil || choiceNum < 1 || choiceNum > len(choices) {
		fmt.Println("Invalid choice. Please enter a number between 1 and", len(choices))
		return
	}

	selectedAnswer := choices[choiceNum-1]
	isCorrect := strings.TrimSpace(selectedAnswer) == strings.TrimSpace(question.CorrectAnswer)

	// Record answer
	err = dao.RecordUserAnswer(db, question.ID, selectedAnswer, isCorrect)
	if err != nil {
		fmt.Printf("Error recording answer: %v\n", err)
		return
	}

	// Show feedback
	fmt.Printf("\n")
	for i, choice := range choices {
		if i == choiceNum-1 {
			if isCorrect {
				fmt.Printf("%d. %s ✅\n", i+1, choice)
			} else {
				fmt.Printf("%d. %s ❌\n", i+1, choice)
			}
		} else if strings.TrimSpace(choice) == strings.TrimSpace(question.CorrectAnswer) {
			fmt.Printf("%d. %s ✅\n", i+1, choice)
		} else {
			fmt.Printf("%d. %s\n", i+1, choice)
		}
	}

	if isCorrect {
		fmt.Printf("\nCorrect! 🎉\n")
	} else {
		fmt.Printf("\nIncorrect. The correct answer was: %s\n", question.CorrectAnswer)
	}

	if question.Explanation != "" {
		fmt.Printf("\nExplanation: %s\n", question.Explanation)
	}
}

func init() {
	rootCmd.AddCommand(quizCmd)
	quizCmd.AddCommand(myCompletedCourses)
	quizCmd.AddCommand(downloadCourseContent)
	quizCmd.AddCommand(startQuizCmd)
}
