package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	BOT_TOKEN := os.Getenv("BOT_TOKEN")

	d, err := discordgo.New("Bot " + BOT_TOKEN)
	if err != nil {
		return
	}

	d.AddHandler(handleMessage)
	d.Identify.Intents = discordgo.IntentsGuildMessages
	err = d.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	fmt.Println("ready")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill, syscall.SIGKILL)
	<-sc
	fmt.Println("closing discord")
	d.Close()
}

func handleMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	CHANNEL_ID := os.Getenv("CHANNEL_ID")
	if m.ChannelID != CHANNEL_ID || m.Author.ID == s.State.User.ID {
		return
	}
	if m.Content == "random poop" {
		s.ChannelMessageSend(CHANNEL_ID, "Fucking poop lover :man_facepalming:")
	}

	fmt.Println("alors là")
	// s.ChannelMessageSend(m.ChannelID, "coucou les poop lovers")
}
