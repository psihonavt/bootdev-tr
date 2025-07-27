package render

import (
	"fmt"
	"io"
	"os"
	"strings"

	dao "github.com/bootdotdev/bootdev/db"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const listHeight = 11

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type item string

func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type quizState int

const (
	questionState quizState = iota
	answerState
	completedState
)

type model struct {
	list           list.Model
	quiz           *dao.Quiz
	currentQIndex  int
	selectedAnswer int
	state          quizState
	quitting       bool
	lastAnswerOk   bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			switch m.state {
			case questionState:
				// Answer selected - add checkmarks to items
				m.selectedAnswer = m.list.Index()

				// Update list items with checkmarks
				currentQ := m.quiz.Questions[m.currentQIndex]
				choices := currentQ.GetAnswerChoices()
				newItems := []list.Item{}

				for i, choice := range choices {
					itemText := choice

					if i == m.selectedAnswer {
						// User's selected answer
						if currentQ.IsCorrectAnswer(choice) {
							itemText += " ✅"
							m.state = answerState
							m.lastAnswerOk = true
							dao.MarkQuestionAnswered(nil, int(currentQ.ID), choice, true)

						} else {
							itemText += " ❌"
							m.state = answerState
							m.lastAnswerOk = false
							dao.MarkQuestionAnswered(nil, int(currentQ.ID), choice, false)
						}
					}
					newItems = append(newItems, item(itemText))
				}

				m.list.SetItems(newItems)
				return m, nil
			default:
				return m, nil
			}

		case "n":
			if m.state == answerState {
				// Hotkey for next question
				return m.nextQuestion(), nil
			}
		}
	}

	// Only update list when in question state
	if m.state == questionState {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) nextQuestion() tea.Model {
	if m.currentQIndex >= len(m.quiz.Questions)-1 {
		m.state = completedState
		return m
	}

	m.currentQIndex++
	m.selectedAnswer = -1
	m.state = questionState

	// Update list with new question
	question := m.quiz.Questions[m.currentQIndex]
	items := []list.Item{}
	for _, choice := range question.GetAnswerChoices() {
		items = append(items, item(choice))
	}

	m.list.SetItems(items)
	m.list.Title = question.QuestionText
	m.list.Select(0) // Reset selection to first item

	return m
}

func (m model) View() string {
	if m.quitting {
		return quitTextStyle.Render("Quiz ended.")
	}

	if m.state == completedState {
		baseMsg := "🎉 Quiz completed! Great job!"
		stats, err := dao.GetQuizStats(nil, m.quiz.CourseUUID)
		var finalMsg string
		if err != nil {
			finalMsg = fmt.Sprintf("%s\n\nError querying stats: %v", baseMsg, err)
		} else {
			finalMsg = fmt.Sprintf("%s\n\n Course: %s\n QuestionsCount: %d\n Total Answers: %d\n Correct Answers %d\n", baseMsg, stats.CourseUUID, stats.QuestionCount, stats.TotalAnswers, stats.CorrectAnswers)
		}
		return quitTextStyle.Render(finalMsg)

	}

	view := "\n" + m.list.View()

	// Add custom status line when in answer state
	if m.state == answerState {
		statusStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Margin(0, 0, 1, 4)
		view += "\n" + statusStyle.Render("Press enter or n for next question")
	}

	return view
}

func RenderQuiz(quiz *dao.Quiz) {
	if len(quiz.Questions) == 0 {
		fmt.Println("No questions available!")
		return
	}
	quiz.ShuffleQuestions()

	q := quiz.Questions[0]

	items := []list.Item{}
	for _, choice := range q.GetAnswerChoices() {
		items = append(items, item(choice))
	}

	const defaultWidth = 20

	m := model{
		quiz:           quiz,
		currentQIndex:  0,
		selectedAnswer: -1,
		state:          questionState,
		quitting:       false,
	}

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = q.QuestionText
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	m.list = l

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
