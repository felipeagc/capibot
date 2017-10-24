package main

import (
	"time"
)

// Server model
type Server struct {
	ID            string    `gorm:"primary_key"`
	Users         []User    `gorm:"many2many:server_users;"`
	Messages      []Message `gorm:"many2many:server_messages;"`
	PlaylistItems []PlaylistItem
}

// User model
type User struct {
	ID               string    `gorm:"primary_key"`
	Messages         []Message `gorm:"many2many:user_messages;"`
	Servers          []Server  `gorm:"many2many:server_users;"`
	PlaylistRequests []PlaylistItem
}

// Message model
type Message struct {
	ID      uint `sql:"AUTO_INCREMENT" gorm:"primary_key"`
	User    User `gorm:"many2many:user_messages;"`
	Content string
	Server  Server `gorm:"many2many:server_messages;"`
	Date    time.Time
}

// PlaylistItem model
type PlaylistItem struct {
	ID     uint `sql:"AUTO_INCREMENT" gorm:"primary_key"`
	Title  string
	URL    string
	Server Server
	User   User
	Date   time.Time
	Played bool
}
