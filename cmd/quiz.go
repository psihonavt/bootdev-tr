package cmd

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	api "github.com/bootdotdev/bootdev/client"
)

func getLessonContent() {
	apiURL := viper.GetString("api_url")
	client := &http.Client{}
	// Best effort - logout should never fail
	r, _ := http.NewRequest("POST", apiURL+"/v1/auth/logout2", bytes.NewBuffer([]byte{}))
	r.Header.Add("X-Refresh-Token", viper.GetString("refresh_token"))
	if resp, err := client.Do(r); err != nil || resp.StatusCode != 200 {
		panic(fmt.Sprintf("Got non-200; response %v; err: %v", resp, err))
	}

	// viper.Set("access_token", "")
	// viper.Set("refresh_token", "")
	// viper.Set("last_refresh", time.Now().Unix())
	// viper.WriteConfig()
	fmt.Println("Logged out successfully.")
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

var completedCoursesCmd = &cobra.Command{
	Use:          "completed_courses",
	Short:        "Displays completed courses",
	PreRun:       requireAuth,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		getCompletedCourses()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(completedCoursesCmd)
}
