package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	api "github.com/bootdotdev/bootdev/client"
	dao "github.com/bootdotdev/bootdev/db"
	render "github.com/bootdotdev/bootdev/render"
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
	Use:   "quiz-mgmt",
	Short: "quiz related commands",
}

var quizModeCmd = &cobra.Command{
	Use:          "quiz",
	Short:        "quiz mode",
	PreRun:       requireAuth,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("")
		fmt.Println(totalRecallLogo)

		courseUUID, err := quizSelect()
		if err != nil {
			fmt.Println(err)
			return nil
		}
		if courseUUID == "" {
			fmt.Println("Noting to do...")
			return nil
		}
		startQuiz(courseUUID)
		return nil
	},
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

var generateQuizCmd = &cobra.Command{
	Use:   "generate <COURSE_UUID>",
	Short: "Generate quiz questions using AI for a course",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		courseUUID := args[0]
		questionsCount, _ := cmd.Flags().GetInt("questions")
		generateQuiz(courseUUID, questionsCount)
		return nil
	},
}

func clearLines(lineCount int) {
	// Clear all the printed lines before starting bubbletea UI
	for i := 0; i < lineCount; i++ {
		fmt.Print("\033[A\033[K") // Move cursor up one line and clear it
	}
}

const totalRecallLogo = `
████████╗ ██████╗ ████████╗ █████╗ ██╗         ██████╗ ███████╗ ██████╗ █████╗ ██╗     ██╗     
╚══██╔══╝██╔═══██╗╚══██╔══╝██╔══██╗██║         ██╔══██╗██╔════╝██╔════╝██╔══██╗██║     ██║     
   ██║   ██║   ██║   ██║   ███████║██║         ██████╔╝█████╗  ██║     ███████║██║     ██║     
   ██║   ██║   ██║   ██║   ██╔══██║██║         ██╔══██╗██╔══╝  ██║     ██╔══██║██║     ██║     
   ██║   ╚██████╔╝   ██║   ██║  ██║███████╗    ██║  ██║███████╗╚██████╗██║  ██║███████╗███████╗
   ╚═╝    ╚═════╝    ╚═╝   ╚═╝  ╚═╝╚══════╝    ╚═╝  ╚═╝╚══════╝ ╚═════╝╚═╝  ╚═╝╚══════╝╚══════╝
`

func withBlinkingMessage(message string, fn func() error) error {
	done := make(chan error, 1)

	// Start the blinking animation in a goroutine
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		visible := true

		for {
			select {
			case <-done:
				// Clear the message when done
				fmt.Print("\033[K\r")
				return
			case <-ticker.C:
				if visible {
					fmt.Printf("\r%s", message)
				} else {
					fmt.Print("\033[K\r")
				}
				visible = !visible
			}
		}
	}()

	// Execute the function
	err := fn()
	done <- err
	close(done)

	return err
}

func quizSelect() (string, error) {
	myCompletedCourses, err := api.GetUserCourses()
	if err != nil {
		return "", fmt.Errorf("Error retrieving user's courses: %w\n", err)
	}

	if len(myCompletedCourses) == 0 {
		return "", fmt.Errorf("No course found")
	}

	type cwq struct {
		c *api.UserCourse
		q *dao.Quiz
	}
	var coursesWithQuizzes []cwq

	fmt.Println("Select a course:")
	lineCount := 1 // "Select a course:" line

	for idx, userCourse := range myCompletedCourses {
		msg := fmt.Sprintf("\033[1m%d\033[0m. %s", idx, userCourse.Title)
		maybeQuiz, err := dao.GetQuiz(nil, userCourse.UUID)
		if err != nil {
			fmt.Printf("Error retrieving a quiz: %v\n", err)
		}
		if maybeQuiz != nil {
			msg = fmt.Sprintf("%s (Total questions available %d)", msg, len(maybeQuiz.Questions))
		} else {
			msg = fmt.Sprintf("%s (No questions available)", msg)
		}
		coursesWithQuizzes = append(coursesWithQuizzes, cwq{c: &userCourse, q: maybeQuiz})
		fmt.Println(msg)
		lineCount++
	}

	reader := bufio.NewReader(os.Stdin)
	selectedIndex := -1
	for selectedIndex == -1 {
		fmt.Print("Let's GO (use the index (in bold (a number))): ")
		lineCount += 1 // prompt lines

		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			lineCount++
			continue
		}

		input = strings.TrimSpace(input)
		index, err := strconv.Atoi(input)
		if err != nil {
			fmt.Printf("Invalid input: %s\n", input)
			lineCount++
			continue
		}

		if index < 0 || index >= len(myCompletedCourses) {
			fmt.Printf("Index %d is out of range. Please select between 0 and %d.\n", index, len(myCompletedCourses)-1)
			lineCount++
			continue
		}
		selectedIndex = index
	}

	courseWithQuiz := coursesWithQuizzes[selectedIndex]

	if courseWithQuiz.q == nil {
		msg := "No questions found for this course"
		if os.Getenv("ANTHROPIC_API_KEY") == "" {
			fmt.Printf("%s. And I don't see the ANTHROPIC_API_KEY in the environment. Use a mock quiz(y) or quit(q)? [y/q]: ", msg)
			lineCount++
			input, err := reader.ReadString('\n')
			if err != nil {
				return "", fmt.Errorf("Error reading input: %w\n", err)
			}
			input = strings.TrimSpace(input)
			if input == "y" {
				clearLines(lineCount)
				return "total-recall-uuid", nil
			} else {
				return "", nil
			}
		} else {
			fmt.Printf("%s. I see the ANTHROPIC_API_KEY in the environment. Do you want to ask Claude to generate N questions based on the course content? [N/q]: ", msg)
			lineCount++

			input, err := reader.ReadString('\n')
			if err != nil {
				return "", fmt.Errorf("Error reading input: %w\n", err)
			}
			input = strings.TrimSpace(input)
			if input == "q" {
				return "", nil
			}
			questionsNumber, err := strconv.Atoi(input)
			if err != nil {
				return "", fmt.Errorf("Invalid input for questions number: %s\n", input)
			}

			err = withBlinkingMessage("Talking to Claude ...", func() error {
				course, err := api.FetchCourseAndLessons(courseWithQuiz.c.UUID)
				if err != nil {
					return fmt.Errorf("Failed to fetch lessons content for %s: %v", courseWithQuiz.c.Title, err)
				}
				quiz, err := api.GenerateQuiz(course, questionsNumber)
				if err != nil {
					return fmt.Errorf("Failed to generate quiz questions: %s", input)
				}
				err = dao.CreateQuiz(nil, quiz)
				if err != nil {
					return fmt.Errorf("Error saving quiz to the db: %v", err)
				}
				return nil
			})
			if err != nil {
				return "", err
			}
		}
	}

	clearLines(lineCount)
	return courseWithQuiz.c.UUID, nil
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

	quiz, err := dao.GetQuiz(db, courseUUID)
	if err != nil {
		fmt.Printf("Error getting quiz: %s", err)
		os.Exit(1)
	}
	render.RenderQuiz(quiz)
}

func generateQuiz(courseUUID string, questionsCount int) {
	fmt.Printf("Generating %d questions for course %s...\n\n", questionsCount, courseUUID)

	// Fetch course content
	course, err := api.FetchCourseAndLessons(courseUUID)
	if err != nil {
		fmt.Printf("Error fetching course content: %v\n", err)
		return
	}

	fmt.Printf("Course: %s\n", course.Title)
	fmt.Printf("Lessons found: %d\n\n", len(course.GetLessons()))

	// Generate quiz using Claude API
	quiz, err := api.GenerateQuiz(course, questionsCount)
	if err != nil {
		fmt.Printf("Error generating quiz: %v\n", err)
		return
	}

	// Print generated questions
	fmt.Printf("=== Generated Questions ===\n\n")
	for i, question := range quiz.Questions {
		fmt.Printf("Question %d: %s\n", i+1, question.QuestionText)

		choices := question.GetAnswerChoices()
		for j, choice := range choices {
			marker := ""
			if choice == question.CorrectAnswer {
				marker = " ✅"
			}
			fmt.Printf("  %c) %s%s\n", 'A'+j, choice, marker)
		}

		if question.Explanation != "" {
			fmt.Printf("  Explanation: %s\n", question.Explanation)
		}
		fmt.Println()
	}

	err = dao.CreateQuiz(nil, quiz)
	if err != nil {
		fmt.Printf(("Error saving quiz to the DB: %v"), err)
		return
	}

	fmt.Printf("Successfully generated %d questions!\n", len(quiz.Questions))
}

func init() {
	rootCmd.AddCommand(quizCmd)
	rootCmd.AddCommand(quizModeCmd)
	quizCmd.AddCommand(myCompletedCourses)
	quizCmd.AddCommand(downloadCourseContent)
	quizCmd.AddCommand(startQuizCmd)
	quizCmd.AddCommand(generateQuizCmd)

	// Add --questions flag with default value of 10
	generateQuizCmd.Flags().IntP("questions", "q", 10, "Number of questions to generate")
}
