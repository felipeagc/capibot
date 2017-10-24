package main

import (
	"io"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
	"github.com/otium/ytdl"
)

// Join the message author's voice channel
func Join(s *discordgo.Session, msg *discordgo.Message) {
	channel, err := s.State.Channel(msg.ChannelID)
	if err != nil {
		// Could not find channel
		return
	}

	g, err := s.State.Guild(channel.GuildID)
	if err != nil {
		// Could not find guild
		return
	}

	for _, vs := range g.VoiceStates {
		if vs.UserID == msg.Author.ID {
			instance := GetInstance(vs.GuildID)
			if instance.VoiceConnection != nil {
				instance.VoiceConnection.ChangeChannel(vs.ChannelID, false, true)
			} else {
				vc, err := s.ChannelVoiceJoin(vs.GuildID, vs.ChannelID, false, true)
				if err != nil {
					// Error joining voice channel
				}
				instance.VoiceConnection = vc
			}
		}
	}
}

// Leave the voice channel in this server
func Leave(s *discordgo.Session, msg *discordgo.Message) {
	channel, err := s.State.Channel(msg.ChannelID)
	if err != nil {
		// Could not find channel
		return
	}

	g, err := s.State.Guild(channel.GuildID)
	if err != nil {
		// Could not find guild
		return
	}

	vc := GetInstance(g.ID).VoiceConnection
	if vc == nil {
		return
	}

	vc.Disconnect()
	GetInstance(g.ID).VoiceConnection = nil
}

// PlayVideo plays a youtube video from a url if the bot is in a voice channel in the specified guild
func PlayVideo(s *discordgo.Session, msg *discordgo.Message, url string) {
	channel, err := s.State.Channel(msg.ChannelID)
	if err != nil {
		// Could not find channel
		log.Fatal(err)
		return
	}

	g, err := s.State.Guild(channel.GuildID)
	if err != nil {
		// Could not find guild
		log.Fatal(err)
		return
	}

	vc := GetInstance(g.ID).VoiceConnection
	if vc == nil {
		return
	}

	// Change these accordingly
	options := dca.StdEncodeOptions
	options.RawOutput = true
	options.Bitrate = 96
	options.Application = "lowdelay"

	videoInfo, err := ytdl.GetVideoInfo(url)
	if err != nil {
		// Handle the error
	}

	format := videoInfo.Formats.Extremes(ytdl.FormatAudioBitrateKey, true)[0]
	downloadURL, err := videoInfo.GetDownloadURL(format)
	if err != nil {
		// Handle the error
	}

	encodingSession, err := dca.EncodeFile(downloadURL.String(), options)
	if err != nil {
		// Handle the error
	}
	defer encodingSession.Cleanup()

	done := make(chan error)
	session := dca.NewStream(encodingSession, vc, done)
	GetInstance(g.ID).StreamingSession = session
	err = <-done

	GetInstance(g.ID).StreamingSession = nil
	log.Println("Finished item")

	if err != nil && err != io.EOF {
		// Handle the error
	}
}
