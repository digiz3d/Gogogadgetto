package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var (
	ADIBOU_ID     string
	bad_ajectives []string
	bad_nouns     []string
	BOT_TOKEN     string
	CHANNEL_ID    string
	MASTER_ID     string
	PORT          string
)

func initEnv() {
	ADIBOU_ID = os.Getenv("ADIBOU_ID")
	BOT_TOKEN = os.Getenv("BOT_TOKEN")
	CHANNEL_ID = os.Getenv("CHANNEL_ID")
	MASTER_ID = os.Getenv("MASTER_ID")
	PORT = os.Getenv("PORT")
}

func readJsonFileAsStrings(path string) []string {
	data, err := os.ReadFile(path)

	if err != nil {
		panic(err)
	}

	var strings []string
	json.Unmarshal([]byte(data), &strings)
	return strings
}

func main() {
	words := []string{"wesh", "bien", "sisi"}

	bad_ajectives = readJsonFileAsStrings("./bad-ajdectives.json")
	bad_nouns = readJsonFileAsStrings("./bad-nouns.json")

	godotenv.Load()
	initEnv()

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

	ticker := time.NewTicker(2 * time.Minute)

	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				d.UpdateGameStatus(0, words[rand.Intn(len(words))])
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "OK")
		currentTime := time.Now()
		fmt.Printf("Received request at %v!!\n", currentTime.Format("2006-01-02 15:04:05"))
	})

	go http.ListenAndServe(":"+PORT, nil)

	fmt.Println("ready")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	fmt.Println("closing discord")
	close(quit)
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
func makeUserRef(userId string) string {
	return fmt.Sprintf("<@%s>", userId)
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
func makeInsult() string {
	return fmt.Sprintf("%s %s", bad_ajectives[rand.Intn(len(bad_ajectives))], bad_nouns[rand.Intn(len(bad_nouns))])
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
	fmt.Printf("Message: %v\n", m.Type)

	if m.ChannelID != CHANNEL_ID || m.Author.ID == s.State.User.ID {
		return
	}

	reg := regexp.MustCompile(`(?i)^w(?P<e>e+)sh.*`)
	matches := reg.FindStringSubmatch(m.Content)

	if matches != nil {
		indexMatch := reg.SubexpIndex("e")
		eeeee := matches[indexMatch]

		if m.Author.ID != ADIBOU_ID && len(matches) > 0 {
			s.ChannelMessageSendReply(CHANNEL_ID, "w"+eeeee+"sh alors", m.Reference())
			return
		}
	}

	hellowords := []string{"hey", "bonjour", "hi", "salut", "wesh"}

	if m.Content == "!insult" {
		s.ChannelMessageSendReply(CHANNEL_ID, fmt.Sprintf("%s %s !", makeUserRef(m.Author.ID), makeInsult()), m.Reference())
	}

	if m.Author.ID == MASTER_ID && contains(hellowords, strings.ToLower(m.Content)) {
		s.ChannelMessageSendReply(CHANNEL_ID, "Hello, master", m.Reference())
		return
	}

	if m.Author.ID == ADIBOU_ID {
		s.MessageReactionAdd(CHANNEL_ID, m.ID, ":hugging:")
	}

	if m.Content == "random poop" {
		s.ChannelMessageSendReply(CHANNEL_ID, "Fucking poop lover :man_facepalming:", m.Reference())
	}
}
