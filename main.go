package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"github.com/shomali11/slacker"
	"github.com/slack-go/slack/slackevents"
	"google.golang.org/api/option"
)

// func printCommandEvents(analyticsChannel <-chan *slacker.CommandEvent) {
// 	for event := range analyticsChannel {
// 		fmt.Println("Command Events")
// 		fmt.Println(event.Timestamp)
// 		fmt.Println(event.Command)
// 		fmt.Println(event.Parameters)
// 		fmt.Println(event.Event)
// 		fmt.Println()
// 	}
// }

func connectGemini() *genai.Client {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}
	return client
}

func main() {
	godotenv.Load()
	botToken := os.Getenv("SLACK_BOT_TOKEN")
	appToken := os.Getenv("SLACK_APP_TOKEN")

	// Initiate slack client
	bot := slacker.NewClient(botToken, appToken)

	// Initiate Gemini client
	client := connectGemini()
	defer client.Close()
	model := client.GenerativeModel("gemini-pro")
	// // This might work in paid versions.
	// model.SystemInstruction = &genai.Content{
	// 	Parts: []genai.Part{genai.Text("Act as an expert in IT, Software Engineering & all related field who is capable to provide solution to any kind of problems of any kind of IT company. Apart from the general greeting, if the question is out of these areas deny to answer it smartly & in a formal way.")},
	// }

	// // Printig command details
	// go printCommandEvents(bot.CommandEvents())

	bot.Command("<text>", &slacker.CommandDefinition{
		Description: "AI Bot",
		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
			event := botCtx.Event()
			if _, ok := event.Data.(*slackevents.AppMentionEvent); ok {
				auth, _ := botCtx.APIClient().AuthTest()
				// isMentioned := strings.Contains(event.Text, "<@"+auth.UserID+">")
				userId := event.UserID
				text := request.Param("text")
				text = strings.ReplaceAll(text, "<@"+auth.UserID+">", "")
				if text == "" {
					text = "Hello"
				}

				// Hit Gemini API
				resp, err := model.GenerateContent(context.Background(), genai.Text(text))
				if err != nil {
					response.Reply(fmt.Sprintf("<@%s>, %v", userId, err))
					log.Println("Error: ", err)
					return
				}

				// Return data to slack
				response.Reply(fmt.Sprintf("<@%s>, %v", userId, resp.Candidates[0].Content.Parts[0]))
			}
		},
	})

	// Que: How this affects, if we don't do this?
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := bot.Listen(ctx)
	if err != nil {
		log.Fatal(err)
	}
}
