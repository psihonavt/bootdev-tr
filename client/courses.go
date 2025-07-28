package api

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
)

type CourseLesson struct {
	UUID                     string `json:"UUID"`
	Slug                     string `json:"Slug"`
	Title                    string `json:"Title"`
	LessonDataCodeCompletion struct {
		Readme string `json:"Readme"`
	} `json:"LessonDataCodeCompletion,omitempty"`
	LessonDataMultipleChoice struct {
		Readme string `json:"Readme"`
	} `json:"LessonDataMultipleChoice,omitempty"`
	LessonDataCodeTests struct {
		Readme string `json:"Readme"`
	} `json:"LessonDataCodeTests,omitempty"`
	LessonDataCLI struct {
		Readme string `json:"Readme"`
	} `json:"LessonDataCLI,omitempty"`
	LessonDataTextInput struct {
		Readme string `json:"Readme"`
	} `json:"LessonDataTextInput,omitempty"`
	LessonDataManual struct {
		Readme string `json:"Readme"`
	} `json:"LessonDataManual,omitempty"`
	Content string `json:"-"`
}

type LessonResp struct {
	Lesson CourseLesson `json:"Lesson"`
}

func (l *CourseLesson) setContent(l2 *CourseLesson) {
	if l2.LessonDataCodeCompletion.Readme != "" {
		l.Content = l2.LessonDataCodeCompletion.Readme
	} else if l2.LessonDataMultipleChoice.Readme != "" {
		l.Content = l2.LessonDataMultipleChoice.Readme
	} else if l2.LessonDataCodeTests.Readme != "" {
		l.Content = l2.LessonDataCodeTests.Readme
	} else if l2.LessonDataCLI.Readme != "" {
		l.Content = l2.LessonDataCLI.Readme
	} else if l2.LessonDataTextInput.Readme != "" {
		l.Content = l2.LessonDataTextInput.Readme
	} else if l2.LessonDataManual.Readme != "" {
		l.Content = l2.LessonDataManual.Readme
	}
}

type CourseChapter struct {
	Lessons []CourseLesson `json:"Lessons"`
}

type Course struct {
	UUID         string `json:"UUID"`
	Slug         string `json:"Slug"`
	Descriptiopn string `json:"ShortDescriptiopn"`
	Title        string `json:"Title"`
	Chapters     []struct {
		Lessons []CourseLesson `json:"Lessons"`
	} `json:"Chapters"`
}

func (c *Course) GetLessons() []*CourseLesson {
	var lessons []*CourseLesson
	for chapter := range c.Chapters {
		for lesson := range c.Chapters[chapter].Lessons {
			lessons = append(lessons, &c.Chapters[chapter].Lessons[lesson])
		}
	}
	return lessons
}

func downloadLessonsContent(c *Course) {
	var wg sync.WaitGroup
	for _, lesson := range c.GetLessons() {
		wg.Add(1)
		go func(l *CourseLesson) {
			defer wg.Done()
			resp, err := fetchWithAuth("GET", fmt.Sprintf("/v1/static/lessons/%s", l.UUID))
			if err != nil {
				log.Printf("Error fetching lessons %s: %s", l.UUID, err)
				return
			}
			var lr LessonResp
			// log.Printf("%s\n", resp)
			err = json.Unmarshal(resp, &lr)
			if err != nil {
				log.Printf("Error parsing lesson %s: %s", l.UUID, err)
				return
			}
			l.setContent(&lr.Lesson)
			if l.Content == "" {
				log.Printf("Couldn't parse content from %s\n", resp)
			}
		}(lesson)
	}
	wg.Wait()
}

func FetchCourseAndLessons(courseUUID string) (*Course, error) {
	resp, err := fetchWithAuth("GET", fmt.Sprintf("/v1/courses/%s", courseUUID))
	if err != nil {
		return nil, err
	}

	var c Course
	err = json.Unmarshal(resp, &c)
	if err != nil {
		return nil, err
	}
	downloadLessonsContent(&c)
	return &c, nil
}
