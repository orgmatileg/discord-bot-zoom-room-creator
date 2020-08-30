package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/donvito/zoom-go/zoomAPI"
	"github.com/donvito/zoom-go/zoomAPI/constants/meeting"
	"github.com/google/uuid"
)

var (
	zoomUserID      = ""
	zoomJwtToken    = ""
	discordBotToken = ""
)

func validationConfigEnv() {
	if env := os.Getenv("ZOOM_USER_ID"); env == "" {
		log.Fatal("ERROR: ENV ZOOM_USER_ID is empty")
	} else {
		zoomUserID = env
	}
	if env := os.Getenv("ZOOM_TOKEN"); env == "" {
		log.Fatal("ERROR: ENV ZOOM_TOKEN is empty")
	} else {
		zoomJwtToken = env
	}
	if env := os.Getenv("DISCORD_BOT_TOKEN"); env == "" {
		log.Fatal("ERROR: ENV DISCORD_BOT_TOKEN is empty")
	} else {
		discordBotToken = env
	}
}

func main() {

	validationConfigEnv()

	discord, err := discordgo.New("Bot " + discordBotToken)
	if err != nil {
		log.Printf("ERROR: error creating Discord session: %s \n", err.Error())
		return
	}

	discord.AddHandler(messageCreate)

	discord.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages)

	err = discord.Open()
	if err != nil {
		log.Printf("ERROR: error opening connection: %s \n", err.Error())
		return
	}

	log.Println("INFO: Bot is now running.  Press CTRL-C to exit.")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	discord.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	if m.Message.Content == "!zoom" {
		joinURL, err := createZoomRoom()
		if err != nil {
			log.Printf("ERROR: %s \n", err.Error())
			return
		}

		_, err = s.ChannelMessageSend(m.Message.ChannelID, joinURL)
		if err != nil {
			log.Printf("ERROR: %s \n", err.Error())
			return
		}

	}
}

func createZoomRoom() (string, error) {

	var (
		zoomURL    = "https://api.zoom.us/v2"
		zoomToken  = zoomJwtToken
		userID     = zoomUserID
		randomUUID = uuid.New().String()
		// duration in minute
		duration = 30
	)

	apiClient := zoomAPI.NewClient(zoomURL, zoomToken)

	resp, err := apiClient.CreateMeeting(userID,
		userID,
		meeting.MeetingTypeInstant,
		"",
		duration,
		"",
		"Asia/Jakarta",
		"", // max 8 chars
		fmt.Sprintf("Discord Bot Zoom Creator - %s", randomUUID),
		nil,
		&zoomAPI.Settings{
			JoinBeforeHost: true,
			ApprovalType:   0,
		})

	if err != nil {
		log.Printf("ERROR: %s \n", err.Error())
		return "", err
	}

	return resp.JoinUrl, nil
}
