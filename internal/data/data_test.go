package data

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/charmbracelet/lipgloss"
)

func TestSetGoals(t *testing.T) {
	goal_one := Goal{
		Name:        "Goal one",
		Description: "Goal one description",
		Number:      1,
		Color:       lipgloss.Color("255"),
	}

	goal_two := Goal{
		Name:        "Goal two",
		Description: "Goal two description",
		Number:      1,
		Color:       lipgloss.Color("1"),
	}

	goals := []Goal{goal_one, goal_two}

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", &err)
	}
	defer db.Close()

	GetGoals(db, goals)
}
