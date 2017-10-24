package main

import (
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
	VoiceConnection  *discordgo.VoiceConnection
	StreamingSession *dca.StreamingSession
}

// GetInstance gets an instance from a guild ID
func GetInstance(guildID string) *Instance {
	for i := 0; i < len(Instances); i++ {
		instance := &Instances[i]
		if instance.GuildID == guildID {
			return instance
		}
	}

	return nil
}

// RegisterInstance will create and register an instance if it doesn't exist yet
func RegisterInstance(guildID string) {
	if GetInstance(guildID) != nil {
		return
	}

	instance := Instance{
		GuildID:          guildID,
		VoiceConnection:  nil,
		StreamingSession: nil,
	}

	Instances = append(Instances, instance)
}
