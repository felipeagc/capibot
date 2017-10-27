package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"syscall"
	"time"

	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var (
	discordToken = os.Getenv("DISCORD_TOKEN")
	dbUser       = os.Getenv("POSTGRES_USER")
	dbPassword   = os.Getenv("POSTGRES_PASSWORD")
	dbName       = os.Getenv("POSTGRES_DB")
)

// DB is the database
var DB *gorm.DB

var (
	joinCommand = NewCommand("join",
		"Joins your current voice channel",
		true,
		[]string{"j"},
		[]Command{},
		func(s *discordgo.Session, event *discordgo.MessageCreate, args []string) {
			channel, err := s.State.Channel(event.ChannelID)
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
				if vs.UserID == event.Message.Author.ID {
					instance, err := GetInstance(vs.GuildID)
					if err != nil {
						log.Println(err)
					}
					instance.JoinVoice(s, vs.ChannelID)
				}
			}
		})

	leaveCommand = NewCommand("leave",
		"Leaves your current voice channel",
		true,
		[]string{"l"},
		[]Command{},
		func(s *discordgo.Session, event *discordgo.MessageCreate, args []string) {
			instance, err := GetInstanceFromMessage(s, event.Message)
			if err != nil {
				log.Println(err)
				return
			}
			instance.LeaveVoice()
		})

	pauseCommand = NewCommand("pause",
		"Pauses the current playlist item",
		false,
		[]string{"p"},
		[]Command{},
		func(s *discordgo.Session, event *discordgo.MessageCreate, args []string) {
			instance, err := GetInstanceFromMessage(s, event.Message)
			if err != nil {
				log.Println(err)
				return
			}

			// TODO: check if user is in the same voice channel as the bot

			if instance.IsCurrentlyPlaying() {
				instance.StreamingSession.SetPaused(true)
			}
		})

	resumeCommand = NewCommand("resume",
		"Resumes the current playlist item",
		false,
		[]string{"r"},
		[]Command{},
		func(s *discordgo.Session, event *discordgo.MessageCreate, args []string) {
			instance, err := GetInstanceFromMessage(s, event.Message)
			if err != nil {
				log.Println(err)
				return
			}

			// TODO: check if user is in the same voice channel as the bot

			if instance.IsCurrentlyPlaying() {
				instance.StreamingSession.SetPaused(false)
			}
		})

	skipCommand = NewCommand("skip",
		"Votes to skip the current playlist item",
		false,
		[]string{"s"},
		[]Command{},
		func(s *discordgo.Session, event *discordgo.MessageCreate, args []string) {
		})

	clearCommand = NewCommand("clear",
		"Clears the playlist",
		true,
		[]string{"c"},
		[]Command{},
		func(s *discordgo.Session, event *discordgo.MessageCreate, args []string) {
			DB.Model(&PlaylistItem{}).Updates(map[string]interface{}{"played": true})
		})

	nextCommand = NewCommand("next",
		"Forcefully skips to the next playlist item",
		true,
		[]string{"n"},
		[]Command{},
		func(s *discordgo.Session, event *discordgo.MessageCreate, args []string) {
			instance, err := GetInstanceFromMessage(s, event.Message)
			if err != nil {
				log.Println(err)
				return
			}

			if instance.VoiceConnection == nil {
				Reply(s, event.Message, "Bot is not in a voice channel.")
				return
			}

			// TODO: check if bot and user are in the same voice channel

			instance.SetAutoPlay(true)

			instance.StopCurrentItem()

			playlistItem, err := instance.TryToPlayNext()
			if err != nil {
				log.Println(err)
				Reply(s, event.Message, "Couldn't find a playlist item to play next.")
				return
			}

			Reply(s, event.Message, "Playing next item: "+playlistItem.Title)
		})

	stopCommand = NewCommand("stop",
		"Stops the current playlist item without playing the next one up.",
		true,
		[]string{},
		[]Command{},
		func(s *discordgo.Session, event *discordgo.MessageCreate, args []string) {
			instance, err := GetInstanceFromMessage(s, event.Message)
			if err != nil {
				log.Println(err)
				return
			}

			instance.SetAutoPlay(false)
			instance.StopCurrentItem()
			// TODO: feedback message
		})

	playCommand = NewCommand("music",
		"Music related commands",
		false,
		[]string{"m"},
		[]Command{
			pauseCommand,
			resumeCommand,
			skipCommand,
			clearCommand,
			nextCommand,
			stopCommand,
		},
		func(s *discordgo.Session, event *discordgo.MessageCreate, args []string) {
			instance, err := GetInstanceFromMessage(s, event.Message)
			if err != nil {
				log.Println(err)
				return
			}

			playlistItems := make(playlistItemSlice, 0)
			DB.Where(map[string]interface{}{"played": false}).Find(&playlistItems)
			sort.Sort(playlistItems)

			if len(args) == 0 {
				replyText := "Playlist: \n"
				for i, item := range playlistItems {
					replyText += strconv.Itoa(i+1) + ". " + item.Title + "\n"
				}
				Reply(s, event.Message, replyText)
				return
			}

			if len(args) == 1 {
				_, err := url.ParseRequestURI(args[0])
				if err == nil {
					youtubeResult, err := YoutubeGetInfo(args[0])
					if err != nil {
						log.Printf("Failed to get youtube info for: %s", args[0])
						return
					}

					err = AddToPlaylist(s, event.Message, *youtubeResult)

					if err != nil {
						Reply(s, event.Message, "Error adding playlist item.")
						return
					}

					Reply(s, event.Message, "Added \""+youtubeResult.Title+"\" to the playlist.")
					if len(playlistItems) <= 0 {
						instance.TryToPlayNext()
					}
					return
				}
			}

			youtubeResults, err := YoutubeSearch(strings.Join(args, " "))

			if err != nil {
				Reply(s, event.Message, "Youtube search failed.")
				return
			}

			if len(youtubeResults) == 0 {
				Reply(s, event.Message, "Youtube search yielded no results.")
				return
			}

			youtubeResult := youtubeResults[0]
			if err != nil {
				log.Printf("Failed to get youtube info for: %s", args[0])
				return
			}

			err = AddToPlaylist(s, event.Message, youtubeResult)

			if err != nil {
				Reply(s, event.Message, "Error adding playlist item.")
				return
			}

			Reply(s, event.Message, "Added \""+youtubeResult.Title+"\" to the playlist.")

			if len(playlistItems) <= 0 {
				instance.TryToPlayNext()
			}
		})
)

func ready(s *discordgo.Session, event *discordgo.Ready) {
	log.Println("Bot is ready.")
}

func guildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {
	instance, err := RegisterInstance(s, event.Guild.ID)
	if err != nil {
		log.Println("Failed to create guild")
		return
	}

	log.Printf("Created guild %s", event.Guild.ID)

	for _, channel := range event.Guild.Channels {
		if channel.Type == discordgo.ChannelTypeGuildVoice {
			lower := strings.ToLower(channel.Name)
			words := []string{
				"music",
				"dj",
				"song",
				"bot",
				"sound",
			}
			for _, word := range words {
				if strings.Contains(lower, word) {
					instance.JoinVoice(s, channel.ID)
					return
				}
			}
		}

	}
}

func guildMemberAdd(s *discordgo.Session, event *discordgo.GuildMemberAdd) {
	user := User{
		ID: event.Member.User.ID,
	}

	DB.FirstOrCreate(&user)

	server := Server{
		ID: event.GuildID,
	}

	DB.FirstOrCreate(&server)

	user.Servers = append(user.Servers, server)

	DB.Save(&user)
}

func messageCreate(s *discordgo.Session, event *discordgo.MessageCreate) {
	go func() {
		date, err := event.Message.Timestamp.Parse()

		if err != nil {
			log.Fatalln("Couldn't parse time of message.")
			return
		}

		channel, err := s.Channel(event.ChannelID)
		if err != nil {
			log.Fatalln("Couldn't find channel of message.")
		}

		// Log message
		message := Message{
			UserID:   event.Message.Author.ID,
			ServerID: channel.GuildID,
			Content:  event.Message.Content,
			Date:     date,
		}

		DB.Create(&message)
	}()

	if !strings.HasPrefix(event.Message.Content, CommandPrefix) {
		return
	}

	args := strings.Fields(event.Message.Content)

	if len(args) == 0 {
		return
	}

	commandName := strings.TrimPrefix(args[0], CommandPrefix)
	args = args[1:]

	for _, command := range Commands {
		if command.name == commandName {
			command.Call(s, event, args)
		} else {
			for _, alias := range command.aliases {
				if alias == commandName {
					command.Call(s, event, args)
				}
			}
		}
	}
}

func main() {
	time.Sleep(6 * time.Second)
	dbInfo := fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable password=%s",
		"db",
		dbUser,
		dbName,
		dbPassword)
	db, err := gorm.Open("postgres", dbInfo)
	DB = db
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
		return
	}

	DB.AutoMigrate(
		&Server{},
		&User{},
		&Message{},
		&PlaylistItem{},
	)
	defer DB.Close()

	RegisterCommand(joinCommand)
	RegisterCommand(leaveCommand)
	RegisterCommand(playCommand)

	dg, err := discordgo.New("Bot " + discordToken)
	if err != nil {
		log.Fatalln("Error creating Discord session,", err)
		return
	}

	dg.AddHandler(ready)
	dg.AddHandler(messageCreate)
	dg.AddHandler(guildCreate)
	dg.AddHandler(guildMemberAdd)

	err = dg.Open()
	if err != nil {
		log.Fatalln("Error opening Discord session,", err)
	}
	defer dg.Close()

	log.Println("Bot now running. Press CTRL-C to close.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}
