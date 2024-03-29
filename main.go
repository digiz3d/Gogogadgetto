package main

import (
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

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var (
	BOT_TOKEN       string
	CHANNEL_ID      string
	GUILD_ID        string
	MASTER_ID       string
	PORT            string
	ADIBOU_ID       string
	GPT3_SECRET_KEY string
	GPT2_SECRET_KEY string
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
	ADIBOU_ID = os.Getenv("ADIBOU_ID")
	GPT3_SECRET_KEY = os.Getenv("GPT3_SECRET_KEY")
	GPT2_SECRET_KEY = os.Getenv("GPT2_SECRET_KEY")
}

func main() {
	words := []string{"A dit bouh", "For night", "Counter Offensive: Global Strike"}

	godotenv.Load()
	initEnv()
	init4chan()
	initGpt3()

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

func getChannelName(s *discordgo.Session, channelId string) string {
	channel, err := s.Channel(channelId)
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

func findUserVoiceState(session *discordgo.Session, userid string) (*discordgo.VoiceState, error) {
	guild, err := session.State.Guild(GUILD_ID)
	if err != nil {
		return nil, errors.New("guild not found")
	}
	for _, vs := range guild.VoiceStates {
		if vs.UserID == userid {
			return vs, nil
		}

	}
	return nil, errors.New("could not find the user in any voice channel")
}

var cachedMessages = make(map[string]*discordgo.Message)

func getPreviousMessage(s *discordgo.Session, prevChannelId string, prevMessageId string) (st *discordgo.Message, err error) {
	cachedPreviousMessage := cachedMessages[prevChannelId+prevMessageId]
	if cachedPreviousMessage != nil {
		fmt.Printf("Got cached message: %v %v : %v\n", cachedPreviousMessage.Type, cachedPreviousMessage.ChannelID, cachedPreviousMessage.Content)
		return cachedPreviousMessage, nil
	}
	prev, err := s.ChannelMessage(prevChannelId, prevMessageId)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Not cached message: %v %v : %v\n", prev.Type, prev.ChannelID, prev.Content)
	cachedMessages[prevChannelId+prevMessageId] = prev
	return prev, nil
}

func onMessageEvent(s *discordgo.Session, m *discordgo.MessageCreate) {
	fmt.Printf("Message: %v %v : %v\n", m.Type, m.ChannelID, m.Content)

	if m.Author.ID == s.State.User.ID || m.Author.ID == ADIBOU_ID {
		return
	}

	if m.MessageReference != nil {
		previousMessages := []string{}

		prevChannelId := m.MessageReference.ChannelID
		prevMessageId := m.MessageReference.MessageID

		shouldAnswer := false

		for {
			prev, err := getPreviousMessage(s, prevChannelId, prevMessageId)
			if err != nil {
				break
			}
			previousMessages = append(previousMessages, prev.Content)
			if prev.MessageReference == nil {
				break
			}
			if prev.Author.ID == s.State.User.ID {
				shouldAnswer = true
			}

			prevChannelId = prev.MessageReference.ChannelID
			prevMessageId = prev.MessageReference.MessageID
		}

		if shouldAnswer {
			answer := answerGpt2(m.Content, previousMessages)
			s.ChannelMessageSendReply(m.ChannelID, answer, m.Reference())
			return
		}
	}

	if strings.HasPrefix(m.Content, "play ") {
		if isPlaying {
			return
		}

		youtubeLink := strings.Replace(m.Content, "play ", "", -1)

		if youtubeLink == "" {
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

		playYoutube(vc, youtubeLink, stopPlay)

		return
	}

	if strings.HasPrefix(m.Content, "sayfr ") {
		if isPlaying {
			return
		}
		textToSay := strings.Replace(m.Content, "sayfr ", "", -1)
		if textToSay == "" {
			return
		}
		userVoiceState, err := findUserVoiceState(s, m.Author.ID)
		if err != nil {
			fmt.Println(err)
			return
		}
		vc, err := s.ChannelVoiceJoin(GUILD_ID, userVoiceState.ChannelID, false, false)
		if err != nil {
			fmt.Println("fml:,", err.Error())
			return
		}
		defer vc.Disconnect()

		fmt.Println("the channel id is,", vc.ChannelID)
		textToSpeech(vc, textToSay)
		return
	}

	if m.Content == "stop" {
		for len(stopPlay) > 0 {
			return
		}
		stopPlay <- true
	}

	if m.ChannelID != CHANNEL_ID {
		//stop processing messages except from particular channel
		return
	}

	if (strings.HasPrefix(m.Content, "Qu") || strings.HasPrefix(m.Content, "Tu ") || strings.HasPrefix(m.Content, "Vous ")) && strings.HasSuffix(m.Content, "?") {
		answer := completeGpt3(m.Content)
		s.ChannelMessageSendReply(m.ChannelID, answer, m.Reference())
		return
	}

	hellowords := []string{"hey", "bonjour", "hi", "salut", "wesh", "yop", "yo", "coucou"}
	if m.Author.ID == MASTER_ID && contains(hellowords, strings.ToLower(m.Content)) {
		s.ChannelMessageSendReply(CHANNEL_ID, "Hello, master", m.Reference())
		return
	}

	weshMatches := weshRegex.FindStringSubmatch(m.Content)
	if weshMatches != nil {
		indexMatch := weshRegex.SubexpIndex("e")
		eeeee := weshMatches[indexMatch]

		if len(weshMatches) > 0 {
			s.ChannelMessageSendReply(CHANNEL_ID, "w"+eeeee+"sh alors", m.Reference())
			return
		}
	}

	if contains([]string{"poop", "caca"}, strings.ToLower(m.Content)) {
		s.ChannelMessageSendReply(CHANNEL_ID, "Fucking poop lover :man_facepalming:", m.Reference())
		return
	}

	if strings.HasPrefix(m.Content, "random ") {
		boardName := strings.Replace(m.Content, "random ", "", -1)
		// get random pic from 4chan
		pic, err := getRandomPicture(boardName)
		if err != nil {
			s.ChannelMessageSendReply(m.ChannelID, "Error: "+err.Error(), m.Reference())
			return
		}
		s.ChannelMessageSendReply(m.ChannelID, pic, m.Reference())
		return
	}

	if m.Content == "reboot" {
		os.Exit(0)
	}
}
