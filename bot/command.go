package main

import (
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	// CommandPrefix is the prefix for commands
	CommandPrefix = "."
	// CommandCooldown in seconds
	CommandCooldown = 1
)

var (
	// Commands is a slice of the registered commands
	Commands []Command
)

// CommandCallback is the type of a command callback
type CommandCallback func(s *discordgo.Session, event *discordgo.MessageCreate, args []string)

// Command stores information related to a command
type Command struct {
	name        string
	description string
	subcommands []Command
	aliases     []string
	adminOnly   bool
	callback    CommandCallback
}

// Call will check for subcommands and call their callbacks or call this command's callback
func (c *Command) Call(i *Instance, s *discordgo.Session, event *discordgo.MessageCreate, args []string) {
	if i.CommandsOnCooldown {
		Reply(s, event.Message, "On cooldown.")
		return
	}

	if c.adminOnly {
		channel, err := s.Channel(event.ChannelID)
		if err != nil {
			log.Println("Couldn't find message channel.")
			return
		}

		member, err := s.GuildMember(channel.GuildID, event.Message.Author.ID)
		if err != nil {
			log.Println("Couldn't find guild member.")
			return
		}

		isAdmin := false

		for _, memberRole := range member.Roles {
			role, err := s.State.Role(channel.GuildID, memberRole)
			if err != nil {
				log.Println(err)
			}

			if role.Permissions&discordgo.PermissionAdministrator != 0 {
				if memberRole == role.ID {
					isAdmin = true
					break
				}
			}
		}

		if !isAdmin {
			Reply(s, event.Message, "You don't have permission to use this command.")
			return
		}
	}

	for _, subcommand := range c.subcommands {
		if len(args) > 0 {
			if args[0] == subcommand.name {
				subcommand.Call(i, s, event, args[1:])
				return
			}

			for _, alias := range subcommand.aliases {
				if alias == args[0] {
					subcommand.Call(i, s, event, args[1:])
					return
				}
			}
		}
	}

	i.CommandsOnCooldown = true

	go func() {
		time.Sleep(CommandCooldown * time.Second)
		i.CommandsOnCooldown = false
	}()

	c.callback(s, event, args)
}

// NewCommand creates a command
func NewCommand(name string,
	description string,
	adminOnly bool,
	aliases []string,
	subcommands []Command,
	callback CommandCallback) Command {
	return Command{
		name:        name,
		description: description,
		adminOnly:   adminOnly,
		aliases:     aliases,
		subcommands: subcommands,
		callback:    callback,
	}
}

// RegisterCommand registers a command at the top level
func RegisterCommand(command Command) {
	Commands = append(Commands, command)
}
