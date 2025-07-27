package cmd

import (
	"fmt"
	"os"

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

	quiz, err := dao.GetQuiz(db, courseUUID)
	if err != nil {
		fmt.Printf("Error getting quiz: %w", err)
		os.Exit(1)
	}
	render.RenderQuiz(quiz)
}

func init() {
	rootCmd.AddCommand(quizCmd)
	quizCmd.AddCommand(myCompletedCourses)
	quizCmd.AddCommand(downloadCourseContent)
	quizCmd.AddCommand(startQuizCmd)
}
