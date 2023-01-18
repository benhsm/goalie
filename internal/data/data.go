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

	Intentions []*Intention `gorm:"many2many:whys_intentions;"`
}

// TableName overrides the gorm table name used by Why to "whys"
func (Why) TableName() string {
	return "whys"
}

type Intention struct {
	ID   uint
	Date time.Time

	Content   string
	Done      bool
	Cancelled bool

	Outcome bool
	// True for outcomes added at the end of the day
	Unintended bool

	// To keep track of where the intention is on the list relative to others
	Position int

	Whys []*Why `gorm:"many2many:whys_intentions;"`
}

func NewStore() Store {
	dataFilePath, err := xdg.DataFile("goalie/goalie.db")
	if err != nil {
		log.Fatalf("Could not find datafile path: %v", err)
	}
	db, err := gorm.Open(sqlite.Open(dataFilePath))
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	db.AutoMigrate(&Why{}, &Intention{}, &Day{})
	return Store{
		db: db,
	}
}

type Store struct {
	db *gorm.DB
}

type WhyStatusEnum int

const (
	Active WhyStatusEnum = iota
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

func (s *Store) UpsertWhys(items []Why) error {
	err := s.db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&items).Error
	return err
}

func (s *Store) DeleteWhys(whys []Why) error {
	err := s.db.Transaction(func(tx *gorm.DB) error {
		for _, why := range whys {
			var err error
			if why.ID > 0 {
				err = s.db.Delete(&why).Error
			}
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func (s *Store) UpsertIntentions(items []Intention) error {
	err := s.db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&items).Error
	return err
}

func (s *Store) GetDaysIntentions(day time.Time) ([]Intention, error) {
	var results []Intention
	err := s.db.Model(&Intention{}).Preload("Whys").Where("date = ?", day).Find(&results).Error
	return results, err
}

// Reviews

type Day struct {
	Date       time.Time
	WhyID      uint
	Why        Why
	Enough     bool
	Reflection string
}

func (s *Store) UpsertDayReview(days []Day) error {
	err := s.db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&days).Error
	return err
}
