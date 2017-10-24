package main

import (
	"github.com/bwmarrin/discordgo"
)

const (
	replyBefore = ""
	replyAfter  = ""
	sendBefore  = ""
	sendAfter   = ""
)

// Send messages to a channel with a string containing the text.
// The message might be subject to change because of templates.
func Send(s *discordgo.Session, channel string, text string) {
	s.ChannelMessageSend(channel, sendBefore+text+sendAfter)
}

// Reply sends a message prepending the text with a mention of another message's author
func Reply(s *discordgo.Session, msg *discordgo.Message, text string) {
	mention := msg.Author.Mention() + ", "
	s.ChannelMessageSend(msg.ChannelID, replyBefore+mention+text+replyAfter)
}
