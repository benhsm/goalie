package data

import (
	"database/sql"
	"time"

	_ "gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/charmbracelet/lipgloss"
)

// Types

type Goal struct {
	ID        uint
	CreatedAt time.Time

	Name        string
	Description string

	Number   int
	Color    lipgloss.Color
	Archived bool

	Intentions []*Intention `gorm:"many2many:goals_intentions;"`
}

type Intention struct {
	ID        uint
	CreatedAt time.Time

	Content   string
	Done      bool
	Cancelled bool

	Goals []*Goal `gorm:"many2many:goals_intentions;"`
}

func GetDailyIntentions(db gorm.DB, day time.Time) []Intention {
	// Day is considered to begin at 4:00AM
	dayStart := time.Date(day.Year(), day.Month(), day.Day(), 4, 0, 0, 0, day.Location())
	dayEnd := dayStart.Add(time.Duration(24) * time.Hour)
	var result []Intention
	db.Where("created_at BETWEEN ? AND ?", dayStart, dayEnd).Find(&result)
	return result
}

type GoalStatusEnum int

const (
	Active = iota
	Archived
	All
)

func GetGoals(db gorm.DB, status GoalStatusEnum) []Goal {
	var result []Goal
	switch status {
	case Active:
		db.Where("archived = 0").Find(&result)
	case Archived:
		db.Where("archived = 1").Find(&result)
	case All:
		db.Find(&result)
	}
	return result
}

func UpsertItems[T Goal | Intention](db *sql.DB, items []T) error {
	if err := db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&items).Error; err != nil {
		return err
	}
	return nil
}
