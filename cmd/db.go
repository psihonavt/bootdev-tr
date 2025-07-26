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
	fmt.Println("Showing stats ... !")
}

func init() {
	rootCmd.AddCommand(dbCmd)
	dbCmd.AddCommand(dbInitCmd)
	dbCmd.AddCommand(dbStatsCmd)
}
