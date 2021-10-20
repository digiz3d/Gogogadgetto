package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

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

	d.AddHandler(onPresenceEvent)
	d.AddHandler(onMessageEvent)
	d.AddHandler(onVoiceEvent)
	d.AddHandler(onChannelEvent)

	d.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsGuildPresences | discordgo.IntentsGuildMessageTyping | discordgo.IntentsGuildVoiceStates

	err = d.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "OK")
		currentTime := time.Now()
		fmt.Printf("Received request at %v!!\n", currentTime.Format("2006-01-02 15:04:05"))
	})

	go http.ListenAndServe(":"+os.Getenv("PORT"), nil)

	fmt.Println("ready")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	fmt.Println("closing discord")
	d.Close()
}

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

func getUserName(s *discordgo.Session, userId string) string {
	user, err := s.User(userId)
	if err != nil {
		return "anonymous blyat"
	}
	return user.Username
}

func getChannelName(s *discordgo.Session, userId string) string {
	channel, err := s.Channel(userId)
	if err != nil {
		return "idk"
	}
	return channel.Name
}

func onPresenceEvent(s *discordgo.Session, m *discordgo.PresenceUpdate) {
	fmt.Printf("Presence: %v %v\n", m.Presence.Status, getUserName(s, m.User.ID))
}

func onVoiceEvent(s *discordgo.Session, m *discordgo.VoiceStateUpdate) {
	fmt.Printf("Voice: user %v, channel: %v, %v\n", getUserName(s, m.UserID), getChannelName(s, m.ChannelID), m.VoiceState.SelfMute)
}

func onChannelEvent(s *discordgo.Session, m *discordgo.VoiceStateUpdate) {
	fmt.Printf("Channel: %v switched to %v \n", getUserName(s, m.UserID), getChannelName(s, m.ChannelID))
}

func onMessageEvent(s *discordgo.Session, m *discordgo.MessageCreate) {
	CHANNEL_ID := os.Getenv("CHANNEL_ID")
	fmt.Printf("Message: %v\n", m.Type)

	if m.ChannelID != CHANNEL_ID || m.Author.ID == s.State.User.ID {
		return
	}

	hellowords := []string{"hey", "bonjour", "hi", "salut", "wesh"}

	if m.Author.ID == os.Getenv("MASTER_ID") && contains(hellowords, strings.ToLower(m.Content)) {
		s.ChannelMessageSendReply(CHANNEL_ID, "Hello, master", m.Reference())
	}

	fmt.Printf("len embeds %v\n", len(m.Embeds))

	if m.Author.ID == os.Getenv("ADIBOU_ID") {
		s.MessageReactionAdd(CHANNEL_ID, m.ID, ":hugging:")
	}

	if m.Content == "random poop" {
		s.ChannelMessageSendReply(CHANNEL_ID, "Fucking poop lover :man_facepalming:", m.Reference())
	}

	fmt.Println("alors là")
}
