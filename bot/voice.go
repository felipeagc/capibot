package main

import (
	"errors"
	"io"
	"sort"

	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
	"github.com/otium/ytdl"
)

// JoinVoice the message author's voice channel
func (i *Instance) JoinVoice(s *discordgo.Session, channelID string) error {
	if i.VoiceConnection != nil {
		err := i.VoiceConnection.ChangeChannel(channelID, false, true)
		if err != nil {
			return err
		}
	} else {
		vc, err := s.ChannelVoiceJoin(i.GuildID, channelID, false, true)
		if err != nil {
			// Error joining voice channel
			return err
		}
		i.VoiceConnection = vc
	}

	return nil
}

// LeaveVoice the voice channel in this server
func (i *Instance) LeaveVoice() error {
	vc := i.VoiceConnection
	if vc == nil {
		return nil
	}

	err := vc.Disconnect()
	if err != nil {
		return err
	}
	i.VoiceConnection = nil

	return nil
}

// PlayItem plays a youtube video from a url if the bot is in a voice channel in the specified guild
func (i *Instance) PlayItem(url string) {
	vc := i.VoiceConnection
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

	i.EncodingSession, err = dca.EncodeFile(downloadURL.String(), options)
	if err != nil {
		// Handle the error
	}
	defer i.EncodingSession.Cleanup()

	done := make(chan error)
	session := dca.NewStream(i.EncodingSession, vc, done)
	i.StreamingSession = session
	err = <-done

	i.EncodingSession = nil
	i.StreamingSession = nil

	if i.AutoPlay {
		i.TryToPlayNext()
	}

	if err != nil && err != io.EOF {
		// Handle the error
	}
}

// IsCurrentlyPlaying returns whether there's audio playing in the instance.
func (i *Instance) IsCurrentlyPlaying() bool {
	return i.StreamingSession != nil || i.EncodingSession != nil
}

// StopCurrentItem stops the currently playing playlist item.
func (i *Instance) StopCurrentItem() {
	if i.IsCurrentlyPlaying() {
		i.StreamingSession.SetPaused(true)
		i.EncodingSession.Stop()
		i.StreamingSession = nil
		i.EncodingSession = nil
	}
}

// TryToPlayNext checks if there is any item currently playing
// If not, then play the next one in the list.
func (i *Instance) TryToPlayNext() (*PlaylistItem, error) {
	if i.IsCurrentlyPlaying() {
		return nil, errors.New("Already playing a playlist item")
	}

	playlistItems := make(playlistItemSlice, 0)
	if err := DB.Where(map[string]interface{}{"played": false}).Find(&playlistItems).Error; err != nil {
		return nil, errors.New("Couldn't find a playlist item to play next")
	}

	if len(playlistItems) <= 0 {
		return nil, errors.New("Playlist is empty")
	}

	sort.Sort(playlistItems)
	playlistItem := playlistItems[0]

	playlistItem.Played = true

	DB.Save(&playlistItem)

	go i.PlayItem(playlistItem.URL)

	return &playlistItem, nil
}
