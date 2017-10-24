package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"syscall"

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

func ready(s *discordgo.Session, event *discordgo.Ready) {
	fmt.Println("Bot is ready.")
	for _, guild := range event.Guilds {
		RegisterInstance(guild.ID)
	}
}

func messageCreate(s *discordgo.Session, event *discordgo.MessageCreate) {
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

var (
	joinCommand = NewCommand("join",
		[]string{"j"},
		[]Command{},
		func(s *discordgo.Session, event *discordgo.MessageCreate, args []string) {
			Join(s, event.Message)
		})

	leaveCommand = NewCommand("leave",
		[]string{"l"},
		[]Command{},
		func(s *discordgo.Session, event *discordgo.MessageCreate, args []string) {
			Leave(s, event.Message)
		})

	ytCommand = NewCommand("yt",
		[]string{},
		[]Command{},
		func(s *discordgo.Session, event *discordgo.MessageCreate, args []string) {
			if len(args) == 0 {
				Reply(s, event.Message, "I need a search query/url")
				return
			}

			if len(args) == 1 {
				_, err := url.ParseRequestURI(args[0])
				if err == nil {
					result, err := YoutubeGetInfo(args[0])
					if err != nil {
						return
					}
					url := "https://www.youtube.com/watch?v=" + result.VideoID
					go PlayVideo(s, event.Message, url)
					return
				}
			}

			results, err := YoutubeSearch(strings.Join(args, " "))

			if err != nil {
				return
			}

			if len(results) == 0 {
				return
			}

			url := "https://www.youtube.com/watch?v=" + results[0].VideoID

			go PlayVideo(s, event.Message, url)
		})
)

func main() {
	dbInfo := fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable password=%s",
		"db",
		dbUser,
		dbName,
		dbPassword)
	db, err := gorm.Open("postgres", dbInfo)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
		return
	}

	db.AutoMigrate(
		&Server{},
		&User{},
		&Message{},
		&PlaylistItem{},
	)
	defer db.Close()

	RegisterCommand(joinCommand)
	RegisterCommand(leaveCommand)
	RegisterCommand(ytCommand)

	dg, err := discordgo.New("Bot " + discordToken)
	if err != nil {
		fmt.Println("Error creating Discord session,", err)
		return
	}

	dg.AddHandler(ready)
	dg.AddHandler(messageCreate)

	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening Discord session,", err)
	}
	defer dg.Close()

	fmt.Println("Bot now running. Press CTRL-C to close.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}
