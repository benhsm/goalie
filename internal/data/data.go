package data

import (
	"log"
	"time"

	"github.com/adrg/xdg"
	"gorm.io/driver/sqlite"
	_ "gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/charmbracelet/lipgloss"
)

// Types

type Why struct {
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

	Whys []*Why `gorm:"many2many:goals_intentions;"`
}

func NewStore() Store {
	dataFilePath, err := xdg.DataFile("why/why.db")
	if err != nil {
		log.Fatalf("Could not find datafile path: %v", err)
	}
	db, err := gorm.Open(sqlite.Open(dataFilePath))
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	db.AutoMigrate(&Why{}, &Intention{})
	return Store{
		db: db,
	}
}

type Store struct {
	db *gorm.DB
}

func (s *Store) GetDailyIntentions(day time.Time) ([]Intention, error) {
	// Day is considered to begin at 4:00AM
	dayStart := time.Date(day.Year(), day.Month(), day.Day(), 4, 0, 0, 0, day.Location())
	dayEnd := dayStart.Add(time.Duration(24) * time.Hour)
	var result []Intention
	err := s.db.Model(&Intention{}).Preload("Whys").Where("created_at BETWEEN ? AND ?", dayStart, dayEnd).Find(&result).Error
	return result, err
}

type WhyStatusEnum int

const (
	Active = iota
	Archived
	All
)

func (s *Store) GetWhys(status WhyStatusEnum) ([]Why, error) {
	var result []Why
	var err error
	switch status {
	case Active:
		err = s.db.Where("archived = 0").Find(&result).Error
	case Archived:
		err = s.db.Where("archived = 1").Find(&result).Error
	case All:
		err = s.db.Find(&result).Error
	}
	return result, err
}

func (s *Store) UpsertItems(items []any) error {
	err := s.db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&items).Error
	return err
}
