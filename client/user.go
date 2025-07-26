package api

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/viper"
)

type HttpTaCResponse struct {
	Status string `json:"status"`
	Data   struct {
		Courses []UserCourse
	} `json:"data"`
}

type UserInfo struct {
	UUID   string `json:"UUID"`
	Handle string `json:"Handle"`
}

type UserCourse struct {
	UUID        string `json:"UUID"`
	Slug        string `json:"Slug"`
	Title       string `json:"Title"`
	CompletedAt string `json:"CompletedAt"`
}

func (uc *UserCourse) IsCompleted() bool {
	return uc.CompletedAt != ""
}

func GetUserInfo() (*UserInfo, error) {
	resp, err := fetchWithAuth("GET", "/v1/users")
	if err != nil {
		return nil, err
	}

	var u UserInfo
	return &u, json.Unmarshal(resp, &u)
}

func GetUserCourses() ([]UserCourse, error) {
	resp, err := fetchWithAuth("GET", fmt.Sprintf("/v1/users/public/%s/tracks_and_courses", viper.Get("user_handle")))
	if err != nil {
		return nil, err
	}

	var tacResp HttpTaCResponse
	err = json.Unmarshal(resp, &tacResp)
	if err != nil {
		return nil, err
	}

	return tacResp.Data.Courses, nil
}
