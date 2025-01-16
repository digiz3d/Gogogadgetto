package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os/exec"
	"strconv"

	"github.com/bwmarrin/dgvoice"
	"github.com/bwmarrin/discordgo"
)

const (
	channels  int = 2
	frameRate int = 48000
	frameSize int = 960
)

func playYoutube(v *discordgo.VoiceConnection, link string, stopPlay chan bool) {
	ytdl := exec.Command("./venv/bin/yt-dlp", "-f", "bestaudio", link, "-o", "-")
	ffmpeg := exec.Command("ffmpeg", "-i", "-", "-f", "s16le", "-ar", strconv.Itoa(frameRate), "-ac", strconv.Itoa(channels), "-")
	ytdlOut, err := ytdl.StdoutPipe()
	if err != nil {
		fmt.Println("ytdlOut err", err)
	}
	ffmpeg.Stdin = ytdlOut
	ffmpegOut, err := ffmpeg.StdoutPipe()
	if err != nil {
		fmt.Println("ffmpegOut err", err)
	}
	ffmpegbuf := bufio.NewReaderSize(ffmpegOut, 16384)

	err = ytdl.Start()
	if err != nil {
		fmt.Println("ytdl.Start err", err)
	}
	defer ytdl.Process.Kill()

	err = ffmpeg.Start()
	if err != nil {
		fmt.Println("ffmpeg.Start err", err)
	}
	defer ffmpeg.Process.Kill()

	send := make(chan []int16, 2)
	defer close(send)

	v.Speaking(true)
	defer v.Speaking(false)

	go func() {
		dgvoice.SendPCM(v, send)
	}()

	for {
		// read data from ffmpeg stdout
		audiobuf := make([]int16, frameSize*channels)

		err = binary.Read(ffmpegbuf, binary.LittleEndian, &audiobuf)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			fmt.Print("End of stream. GGWP")
			return
		}
		if err != nil {
			fmt.Print("error reading from ffmpeg stdout", err)
			return
		}

		// Send received PCM to the sendPCM channel
		select {
		case send <- audiobuf:
		case <-stopPlay:
			return
		}
	}

}
