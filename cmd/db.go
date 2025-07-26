package cmd

import (
	"fmt"

	dao "github.com/bootdotdev/bootdev/db"
	"github.com/spf13/cobra"
)

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Database management commands",
	Long:  `Commands for managing the local SQLite database including initialization and statistics.`,
}

var dbInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the database",
	RunE: func(cmd *cobra.Command, args []string) error {
		initDatabase()
		return nil
	},
}

var dbStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show database statistics",
	RunE: func(cmd *cobra.Command, args []string) error {
		showDatabaseStats()
		return nil
	},
}

var dbResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset database (purge and recreate all tables)",
	RunE: func(cmd *cobra.Command, args []string) error {
		resetDatabase()
		return nil
	},
}

func initDatabase() {
	config := dao.DBConfig{
		InitFile: "db/init.sql",
		DBPath:   "data/bootdev.db",
	}
	_, err := dao.InitializeDatabase(config)
	if err != nil {
		fmt.Printf("Error initting the DB %v", err)
		return
	}
	fmt.Println("Showing stats ... !")
}

func showDatabaseStats() {
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

	stats, err := dao.GetQuizStats(db)
	if err != nil {
		fmt.Printf("Error retrieving database stats: %v\n", err)
		return
	}

	fmt.Println("=== Database Statistics by Quiz ===")
	for _, stat := range stats {
		fmt.Printf("\nCourse UUID: %s\n", stat.CourseUUID)
		fmt.Printf("  Questions: %d\n", stat.QuestionCount)
		fmt.Printf("  Total Answers: %d\n", stat.TotalAnswers)
		fmt.Printf("  Correct Answers: %d\n", stat.CorrectAnswers)
		if stat.TotalAnswers > 0 {
			fmt.Printf("  Correctness Rate: %.1f%%\n", stat.CorrectnessRate)
		} else {
			fmt.Printf("  Correctness Rate: No answers yet\n")
		}
	}
}

func resetDatabase() {
	config := dao.DBConfig{
		InitFile: "db/init.sql",
		DBPath:   "data/bootdev.db",
	}
	err := dao.ResetDatabase(config)
	if err != nil {
		fmt.Printf("Error resetting database: %v\n", err)
		return
	}
	fmt.Println("Database reset successfully!")
}

func init() {
	rootCmd.AddCommand(dbCmd)
	dbCmd.AddCommand(dbInitCmd)
	dbCmd.AddCommand(dbStatsCmd)
	dbCmd.AddCommand(dbResetCmd)
}
