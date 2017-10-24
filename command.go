package main

import "github.com/bwmarrin/discordgo"
import "time"

const (
	// CommandPrefix is the prefix for commands
	CommandPrefix = "."
	// CommandCooldown in seconds
	CommandCooldown = 1
)

var (
	// Commands is a slice of the registered commands
	Commands []Command
	// OnCooldown is true if command execution is temporarily blocked to avoid spam
	OnCooldown bool
)

// CommandCallback is the type of a command callback
type CommandCallback func(s *discordgo.Session, event *discordgo.MessageCreate, args []string)

// Command stores information related to a command
type Command struct {
	name        string
	subcommands []Command
	aliases     []string
	callback    CommandCallback
}

// Call will check for subcommands and call their callbacks or call this command's callback
func (c *Command) Call(s *discordgo.Session, event *discordgo.MessageCreate, args []string) {
	if OnCooldown {
		// TODO: send some message about spam or something
		s.ChannelMessageSend(event.ChannelID, "On cooldown")
		return
	}

	for _, subcommand := range c.subcommands {
		if len(args) > 0 {
			if args[0] == subcommand.name {
				subcommand.Call(s, event, args[1:])
				return
			}

			for _, alias := range subcommand.aliases {
				if alias == args[0] {
					subcommand.Call(s, event, args[1:])
					return
				}
			}
		}
	}

	OnCooldown = true

	go func() {
		time.Sleep(CommandCooldown * time.Second)
		OnCooldown = false
	}()

	c.callback(s, event, args)
}

// NewCommand creates a command
func NewCommand(name string, aliases []string, subcommands []Command, callback CommandCallback) Command {
	return Command{
		name:        name,
		aliases:     aliases,
		subcommands: subcommands,
		callback:    callback,
	}
}

// RegisterCommand registers a command at the top level
func RegisterCommand(command Command) {
	Commands = append(Commands, command)
}
