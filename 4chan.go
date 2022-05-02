package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
)

// a 4chan board as found in the boards.json file
type Board struct {
	Board string `json:"board"`
}

var boards []Board
var boardNames []string

func readJsonFileAsBoards(path string) []Board {
	data, err := os.ReadFile(path)

	if err != nil {
		panic(err)
	}

	var boa []Board
	json.Unmarshal([]byte(data), &boa)
	return boa
}

func init4chan() {
	// reads 4chans board json file into a string of the board short names
	boards = readJsonFileAsBoards("./boards.json")
	for _, board := range boards {
		boardNames = append(boardNames, board.Board)
	}
	fmt.Print(boardNames)
}

func getBoardsList() string {
	list := []string{}

	for _, board := range boards {
		list = append(list, board.Board)
	}

	return strings.Join(list, ", ")
}

func sendRequest(url string, jsonResponse chan string) {
	resp, err := http.Get(url)
	if err != nil {
		jsonResponse <- "error"
	}
	defer resp.Body.Close()

	resp2, err2 := io.ReadAll(resp.Body)

	if err2 != nil {
		jsonResponse <- "error"
	}

	jsonResponse <- string(resp2)
}

func getRandomThread(boardName string) (string, error) {
	if boardName == "" {
		boardName = "b"
	}

	if !contains(boardNames, boardName) {
		return "", errors.New("That board doesn't exist. Try one of these: " + getBoardsList())
	}

	url := "https://a.4cdn.org/" + boardName + "/catalog.json"
	resp := make(chan string)
	go sendRequest(url, resp)
	result := <-resp
	if result == "error" {
		return "", errors.New("Error getting catalog.json")
	}

	var pages []interface{}
	var threads []string

	json.Unmarshal([]byte(result), &pages)

	// for each page, add the threads to the threads array
	for _, page := range pages {
		for _, thread := range page.(map[string]interface{})["threads"].([]interface{}) {

			noInt := int(thread.(map[string]interface{})["no"].(float64))
			noStr := fmt.Sprint(noInt)
			threads = append(threads, noStr)
		}
	}

	selectedThread := threads[rand.Intn(len(threads))]

	return selectedThread, nil
}

// function to get a random picture from a thread
func getRandomPicture(boardName string) (string, error) {
	thread, err := getRandomThread(boardName)
	if err != nil {
		return "", err
	}
	url := "https://a.4cdn.org/" + boardName + "/thread/" + thread + ".json"
	resp := make(chan string)
	go sendRequest(url, resp)
	result := <-resp
	if result == "error" {
		return "", errors.New("Error getting thread")
	}

	var posts interface{}
	json.Unmarshal([]byte(result), &posts)

	images := []string{}

	// for each post, add the images to the images array
	for _, post := range posts.(map[string]interface{})["posts"].([]interface{}) {
		if post.(map[string]interface{})["ext"] == ".jpg" {
			timestamp := int(post.(map[string]interface{})["tim"].(float64))
			ts := fmt.Sprint(timestamp)

			images = append(images, "https://i.4cdn.org/"+boardName+"/"+ts+post.(map[string]interface{})["ext"].(string))
		}
	}

	randomImage := images[rand.Intn(len(images))]

	return randomImage, nil
}
