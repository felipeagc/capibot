package main

import (
	"time"
)

// Server model
type Server struct {
	ID            string `gorm:"primary_key"`
	Users         []User `gorm:"many2many:server_users;"`
	PlaylistItems []PlaylistItem
	Messages      []Message
}

// User model
type User struct {
	ID               string   `gorm:"primary_key"`
	Servers          []Server `gorm:"many2many:server_users;"`
	PlaylistRequests []PlaylistItem
	Messages         []Message
}

// Message model
type Message struct {
	ID       uint `sql:"AUTO_INCREMENT" gorm:"primary_key"`
	UserID   string
	Content  string
	ServerID string
	Date     time.Time
}

// PlaylistItem model
type PlaylistItem struct {
	ID       uint `sql:"AUTO_INCREMENT" gorm:"primary_key"`
	Title    string
	URL      string
	ServerID string
	UserID   string
	Date     time.Time
	Played   bool
}

type playlistItemSlice []PlaylistItem

func (p playlistItemSlice) Len() int {
	return len(p)
}

func (p playlistItemSlice) Less(i, j int) bool {
	return p[i].Date.Before(p[j].Date)
}

func (p playlistItemSlice) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
