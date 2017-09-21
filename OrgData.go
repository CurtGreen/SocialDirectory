package main

import (
	"time"

	"github.com/jinzhu/gorm"
)

// Organization type defines model for User's various Organizations in a 1 to Many relationship
type Organization struct {
	gorm.Model
	UserID   int
	Name     string
	Clients  []Membership `gorm:"ForeignKey:OrganizationID"`
	Events   []Event      `gorm:"ForeignKey:PageID"`
	Messages []Message    `gorm:"ForeignKey:PageID"`
	Location Address      `gorm:"ForeignKey:OrganizationID"`
	Contacts Contact
}

// Membership type defines model to track a User's membership with an Organization
type Membership struct {
	gorm.Model
	UserID         int
	OrganizationID int
	Role           string
	Approved       bool
}

// Contact type holds Contact information for User's Organization
type Contact struct {
	gorm.Model
	Email string
	Phone string
}

// Event type defines Model for Organization's Events in 1 to Many relationship
type Event struct {
	gorm.Model
	Time        time.Time
	Description string
	PageID      int
}

// Message type defines Model for Organization messages
type Message struct {
	gorm.Model
	PageID   int
	OwnerID  int
	Text     string
	Comments []Comment `gorm:"ForeignKey:MessageID"`
}

// Comment type defines model for Message comments in a 1 to Many Relationship
type Comment struct {
	gorm.Model
	MessageID int
	Text      string
}
