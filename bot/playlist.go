package main

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

// AddToPlaylist adds a youtube video to the playlist
func AddToPlaylist(s *discordgo.Session, msg *discordgo.Message, youtubeResult YoutubeResult) error {
	url := "https://www.youtube.com/watch?v=" + youtubeResult.VideoID

	channel, err := s.Channel(msg.ChannelID)
	if err != nil {
		log.Printf("Failed to get channel with ID: %s", msg.ChannelID)
		return err
	}

	date, err := msg.Timestamp.Parse()

	if err != nil {
		log.Fatalln("Couldn't parse time of message.")
		return err
	}

	// Create PlaylistItem and add it to the DB
	playlistItem := PlaylistItem{
		Title:    youtubeResult.Title,
		URL:      url,
		ServerID: channel.GuildID,
		UserID:   msg.Author.ID,
		Date:     date,
		Played:   false,
	}

	DB.Create(&playlistItem)

	return nil
}
