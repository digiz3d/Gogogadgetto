package main

import (
	"context"
	"fmt"

	gogpt "github.com/sashabaranov/go-gpt3"
)

var client *gogpt.Client
var ctx context.Context

func initGpt3() {
	client = gogpt.NewClient(GPT3_SECRET_KEY)
	ctx = context.Background()

}

func completeGpt3(query string) string {
	req := gogpt.CompletionRequest{
		MaxTokens:   50,
		Prompt:      query,
		Temperature: 0.1,
	}
	resp, err := client.CreateCompletion(ctx, "text-curie-001", req)
	if err != nil {
		return "Jsp 1" + err.Error()
	}

	fmt.Println(resp.Choices[0].Text)

	return resp.Choices[0].Text
}
