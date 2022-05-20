package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type Conversation struct {
	GeneratedResponses []string `json:"generated_responses"`
	PastUserInputs     []string `json:"past_user_inputs"`
}

type Response struct {
	GeneratedText string       `json:"generated_text"`
	Conversation  Conversation `json:"conversation"`
}

func postRequest(query string, previousMessages []string, jsonResponse chan string) {
	inputs := make(map[string]interface{})
	inputs["past_user_inputs"] = previousMessages
	inputs["generated_responses"] = []string{"Hi."}
	inputs["text"] = query

	bytesToSend, _ := (json.Marshal(map[string]map[string]interface{}{"inputs": inputs}))
	bufferToSend := bytes.NewBuffer(bytesToSend)

	req, err := http.NewRequest("POST", "https://api-inference.huggingface.co/models/facebook/blenderbot-400M-distill", bufferToSend)
	if err != nil {
		jsonResponse <- "rr"
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+GPT2_SECRET_KEY)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		jsonResponse <- "error"
	}
	defer resp.Body.Close()

	resp2, err2 := io.ReadAll(resp.Body)

	if err2 != nil {
		jsonResponse <- "error"
	}

	var response Response
	err = json.Unmarshal(resp2, &response)
	if err != nil {
		jsonResponse <- "error"
	}

	jsonResponse <- string(response.GeneratedText)
}

func answerGpt2(query string, previousMessages []string) string {
	strChan := make(chan string)
	go postRequest(query, previousMessages, strChan)
	return <-strChan
}
