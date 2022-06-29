package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"strconv"

	"github.com/bwmarrin/dgvoice"
	"github.com/bwmarrin/discordgo"
)

func textToSpeech(v *discordgo.VoiceConnection, text string) {
	url := fmt.Sprintf("http://translate.google.com/translate_tts?ie=UTF-8&total=1&idx=0&textlen=32&client=tw-ob&q=%s&tl=%s", url.QueryEscape(text), "fr")
	response, err := http.Get(url)
	if err != nil {
		return
	}
	defer response.Body.Close()

	ffmpeg := exec.Command("ffmpeg", "-i", "-", "-f", "s16le", "-ar", strconv.Itoa(frameRate), "-ac", strconv.Itoa(channels), "-")
	ffmpeg.Stdin = response.Body
	ffmpegOut, err := ffmpeg.StdoutPipe()
	if err != nil {
		fmt.Println("ffmpegOut err", err)
	}
	ffmpegbuf := bufio.NewReaderSize(ffmpegOut, 16384)

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
