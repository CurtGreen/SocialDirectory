package main

import (
	"log"
	"time"

	"github.com/jinzhu/gorm"
)

// User type defines model for database records that represent our user's properties
type User struct {
	gorm.Model
	Name           string
	Emails         []Email `gorm:"ForeignKey:UserID"`
	Password       string
	About          Profile        `gorm:"ForeignKey:UserID"`
	Organizations  []Organization `gorm:"ForeignKey:UserID"`
	Memberships    []Membership   `gorm:"ForeignKey:UserID"`
	Roles          []Role         `gorm:"ForeignKey:UserID"`
	Subscription   bool
	SocialAccounts []SocialUser `gorm:"ForeignKey:UserID"`
}

// Save method for User Model
func (user *User) Save() error {
	Db, err := gorm.Open(DATABASE, CREDENTIALS)
	defer Db.Close()
	if err != nil {
		return err
	}

	Db.Save(user)
	return err
}

// Create method for User Model
func (user *User) Create() error {
	Db, err := gorm.Open(DATABASE, CREDENTIALS)
	defer Db.Close()
	if err != nil {
		return err
	}

	Db.Create(user)
	return err
}

// Delete method for User Model
func (user *User) Delete() error {
	Db, err := gorm.Open(DATABASE, CREDENTIALS)
	defer Db.Close()
	if err != nil {
		return err
	}

	Db.Delete(user)
	return err
}

//GetByID Query for Entire User struct from DATABASE
func (user *User) GetByID(uid uint) error {
	Db, err := gorm.Open(DATABASE, CREDENTIALS)
	defer Db.Close()
	if err != nil {
		return err
	}

	Db.Debug().Preload("Emails").Preload("About").Preload("Organizations").Preload("Organizations.Location").Preload("Memberships").Preload("SocialAccounts").Preload("Roles").Where("id = ?", uid).Find(user)
	return err
}

// Profile type is model representation of User profiles, User 1 to 1 relationship
type Profile struct {
	gorm.Model
	DOB        time.Time `gorm:"type:date;"`
	Location   Address   `gorm:"ForeignKey:ProfileID"`
	Techniques []Speciality
	UserID     int
}

// Email type stores Users emails in 1 to Many Relationship
type Email struct {
	gorm.Model
	Email  string
	UserID int
}

// UID method for returning userID from email lookup
func (email *Email) UID(emailAddress string) uint {
	Db, err := gorm.Open(DATABASE, CREDENTIALS)
	defer Db.Close()
	if err != nil {
		log.Panic(err.Error())
	}
	// Define Result struct for Database Query
	type Result struct {
		UserID uint
	}

	result := Result{}
	Db.Table("emails").Select("user_id").Where("email = ?", emailAddress).Scan(&result)

	return result.UserID
}

// Address type stores Addresses
type Address struct {
	gorm.Model
	Street         string
	Ext            string
	Lat            float64 `gorm:"type:double precision"`
	Lng            float64 `gorm:"type:double precision"`
	OrganizationID int     `sql:"unique"`
	ProfileID      int
}

//Role type holds Application wide Roles for User
type Role struct {
	gorm.Model
	Name   string
	UserID int
}

// Speciality type holds User's practicing specialities
type Speciality struct {
	gorm.Model
	Name string
}

// SocialUser model holds information about User's Social Media Provider Account
type SocialUser struct {
	gorm.Model
	Provider       string
	ProviderUserID string
	UserID         int
}
