package main

import (
	"errors"

	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
)

var (
	// Instances is an array of registered instances
	Instances = make([]Instance, 0)
)

// Instance represents a discord server
type Instance struct {
	GuildID          string
	AutoPlay         bool
	VoiceConnection  *discordgo.VoiceConnection
	StreamingSession *dca.StreamingSession
	EncodingSession  *dca.EncodeSession
	// CommandsOnCooldown is true if command execution is temporarily blocked to avoid spam
	CommandsOnCooldown bool
}

// GetInstance gets an instance from a guild ID
func GetInstance(guildID string) (*Instance, error) {
	for i := 0; i < len(Instances); i++ {
		instance := &Instances[i]
		if instance.GuildID == guildID {
			return instance, nil
		}
	}

	return nil, errors.New("Couldn't find instance with guild ID: " + guildID)
}

// GetInstanceFromMessage gets an instance associated with the message's guild
func GetInstanceFromMessage(s *discordgo.Session, msg *discordgo.Message) (*Instance, error) {
	channel, err := s.State.Channel(msg.ChannelID)
	if err != nil {
		// Could not find channel
		return nil, err
	}

	instance, err := GetInstance(channel.GuildID)

	if err != nil {
		return nil, err
	}

	return instance, nil
}

// RegisterInstance will create and register an instance if it doesn't exist yet
func RegisterInstance(s *discordgo.Session, guildID string) (*Instance, error) {
	if DB == nil {
		panic("Tried to create instance, but DB doesn't exist.")
	}

	_, err := GetInstance(guildID)
	if err == nil {
		// Already registered
		return nil, err
	}

	server := Server{
		ID: guildID,
	}

	DB.FirstOrCreate(&server)

	guild, err := s.Guild(guildID)

	if err != nil {
		return nil, err
	}

	for _, member := range guild.Members {
		user := User{
			ID: member.User.ID,
		}

		DB.FirstOrCreate(&user)

		user.Servers = append(user.Servers, server)

		DB.Save(&user)
	}

	Instances = append(Instances, Instance{
		GuildID:            guildID,
		AutoPlay:           true,
		VoiceConnection:    nil,
		StreamingSession:   nil,
		CommandsOnCooldown: false,
	})

	i, err := GetInstance(guildID)
	if err != nil {
		return nil, err
	}

	return i, nil
}

// SetAutoPlay sets whether the playlist should auto play or not
func (i *Instance) SetAutoPlay(autoPlay bool) {
	i.AutoPlay = autoPlay
}
