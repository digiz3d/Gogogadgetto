package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/dgvoice"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var (
	bad_ajectives []string
	bad_nouns     []string
	BOT_TOKEN     string
	CHANNEL_ID    string
	GUILD_ID      string
	MASTER_ID     string
	PORT          string
)

var stopPlay = make(chan bool)
var isPlaying = false
var weshRegex = regexp.MustCompile(`(?i)^w(?P<e>e+)sh.*`)

func initEnv() {
	BOT_TOKEN = os.Getenv("BOT_TOKEN")
	CHANNEL_ID = os.Getenv("CHANNEL_ID")
	GUILD_ID = os.Getenv("GUILD_ID")
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
	words := []string{"Adibou", "Fortnite", "Couter strike: Global warming"}

	bad_ajectives = readJsonFileAsStrings("./bad-ajdectives.json")
	bad_nouns = readJsonFileAsStrings("./bad-nouns.json")

	godotenv.Load()
	initEnv()

	discord, err := discordgo.New("Bot " + BOT_TOKEN)
	if err != nil {
		fmt.Printf("Soooo%v\n", err)
		return
	}

	discord.AddHandler(onPresenceEvent)
	discord.AddHandler(onMessageEvent)
	discord.AddHandler(onVoiceEvent)
	discord.AddHandler(onChannelEvent)

	discord.Identify.Intents = discordgo.IntentsAll

	err = discord.Open()
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
				discord.UpdateGameStatus(0, words[rand.Intn(len(words))])
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
	discord.Close()
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

func findUserVoiceState(session *discordgo.Session, userid string) (*discordgo.VoiceState, error) {
	guild, err := session.State.Guild(GUILD_ID)
	if err != nil {
		return nil, errors.New("Guild not found. Should never happen.")
	}
	for _, vs := range guild.VoiceStates {
		if vs.UserID == userid {
			return vs, nil
		}

	}
	return nil, errors.New("Could not find the user in any voice channel.")
}

func onMessageEvent(s *discordgo.Session, m *discordgo.MessageCreate) {
	fmt.Printf("Message: %v\n", m.Type)

	if m.Author.ID == s.State.User.ID {
		// ignore own messages
		return
	}

	if strings.HasPrefix(m.Content, "play ") {
		if isPlaying {
			return
		}
		for len(stopPlay) > 0 {
			<-stopPlay
		}

		track := strings.Replace(m.Content, "play ", "", -1)
		filename := fmt.Sprintf("%v.mp3", track)
		if _, err := os.Stat(filename); err != nil {
			fmt.Printf("The file %v does not exists.", filename)
			return
		}

		userVoiceState, err := findUserVoiceState(s, m.Author.ID)
		if err != nil {
			fmt.Println(err)
			return
		}

		isPlaying = true
		defer func() { isPlaying = false }()

		vc, err := s.ChannelVoiceJoin(GUILD_ID, userVoiceState.ChannelID, false, false)
		if err != nil {
			fmt.Println("fml:,", err.Error())
			return
		}
		defer vc.Disconnect()

		fmt.Println("the channel id is,", vc.ChannelID)

		vc.Speaking(true)
		defer vc.Speaking(false)

		dgvoice.PlayAudioFile(vc, filename, stopPlay)

		return
	}

	if m.Content == "stop" {
		stopPlay <- true
		return
	}

	if m.ChannelID != CHANNEL_ID {
		//stop processing messages except from particular channel
		return
	}

	matches := weshRegex.FindStringSubmatch(m.Content)

	if matches != nil {
		indexMatch := weshRegex.SubexpIndex("e")
		eeeee := matches[indexMatch]

		if len(matches) > 0 {
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

	if contains([]string{"poop", "caca"}, strings.ToLower(m.Content)) {
		s.ChannelMessageSendReply(CHANNEL_ID, "Fucking poop lover :man_facepalming:", m.Reference())
	}
}
