package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	openai "github.com/sashabaranov/go-openai"
	"github.com/slack-go/slack"
	"github.com/joho/godotenv"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
	"log"
)

const PROMPT_PREFIX = "Then, list 10 meal dishes one could make using any of the following ingredients, which is a list separated by a comma. Each entry represents a full ingredient or product to be considered. Prioritize combinations of ingredients from highly-rated existing recipes. Highlight specific ingredients needed or not included using parenthesis. The list of ingredients starts now "

type Month int

const (
	January Month = 1 + iota
	February
	March
	April
	May
	June
	July
	August
	September
	October
	November
	December
)

var year = time.Now().Year()
var month = time.Now().Month()
var day = time.Now().Day()
var weekday = time.Now().Weekday()

type Theme struct {
		Name   string `json:"name"`
		Prompt struct {
			Season   []string `json:"season"`
			Keywords []string `json:"keywords"`
			OpenAI   string   `json:"OpenAI"`
		} `json:"prompt"`
}

func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

func sendSlackmessage() {
	RESPONSE_FILE := os.Getenv("RESPONSE_FILE")
	content, err := ioutil.ReadFile(RESPONSE_FILE)
	if err != nil {
		fmt.Println("Err")
	}

	var slackMessage = string(content)
	formattedSlackMessage := fmt.Sprintf(strings.Replace(slackMessage, "\"", "", 3))
	SLACK_API_KEY := os.Getenv("SLACK_API_KEY")
	api := slack.New(SLACK_API_KEY)
	CHANNEL_ID := os.Getenv("CHANNEL_ID")
	channelID, timestamp, err := api.PostMessage(
		CHANNEL_ID,
		slack.MsgOptionText(formattedSlackMessage, false),
		slack.MsgOptionAsUser(false), // Add this if you want that the bot would post message as a user, otherwise it will send response using the default slackbot
	)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	fmt.Printf("Message successfully sent to channel %s at %s", channelID, timestamp)
}

func openai_dinner() {
	RESPONSE_FILE := os.Getenv("RESPONSE_FILE")
	f, err := os.Create(RESPONSE_FILE) // create &/or clear the local file
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close() // close the file with defer

	headline := fmt.Sprintf("*Tanglesprings Specials for %s %s %d, %d*\n", weekday, month, day, year)
	f.Write([]byte(headline))
	INVENTORY_FILE := os.Getenv("INVENTORY_FILE")
	content, err := ioutil.ReadFile(INVENTORY_FILE) // opens the inventory file just created
	if err != nil {
		fmt.Println("Err")
	}
	var openaiPrompt = string(content)
	OPENAI_API_KEY := os.Getenv("OPENAI_API_KEY")
	c := openai.NewClient(OPENAI_API_KEY)
	ctx := context.Background()
	req := openai.ChatCompletionRequest{
		Model:     openai.GPT3Dot5Turbo,
		MaxTokens: 500,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: openaiPrompt,
			},
		},
		Stream: true,
	}
	stream, err := c.CreateChatCompletionStream(ctx, req)
	if err != nil {
		fmt.Printf("ChatCompletionStream error: %v\n", err)
		return
	}
	defer stream.Close()

	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return
		}

		if err != nil {
			fmt.Printf("\nStream error: %v\n", err)
			return
		}
		var responseContent = response.Choices[0].Delta.Content
		f.Write([]byte(responseContent)) //write directly into file
	}
}

func main() {
	err := godotenv.Load()
	  if err != nil {
		log.Fatal("Error loading .env file")
	  }
	GROCY_URL := os.Getenv("GROCY_URL")
	url := GROCY_URL
	method := "GET"
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return
	}
	GROCY_API_KEY := os.Getenv("GROCY_API_KEY")
	req.Header.Add("GROCY-API-KEY", GROCY_API_KEY)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	var products []map[string]interface{} //defining the array of products
	
	jsonData := body //json array to be decoded to an array
	//decoding JSON array to products array
	err = json.Unmarshal([]byte(jsonData), &products)
	if err != nil {
		fmt.Println("Error while decoding the data", err.Error())
	}
	INVENTORY_FILE := os.Getenv("INVENTORY_FILE")
	f, err := os.Create(INVENTORY_FILE) // create the inventory file
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close() // close the file with defer

	PERSONALITIES := os.Getenv("PERSONALITIES")
	jsonContent, err := ioutil.ReadFile(PERSONALITIES) // opens the inventory file just created
	if err != nil {
		fmt.Println("Err")
	}
	var jsonInput = string(jsonContent)
	var Themes []Theme

	err = json.Unmarshal([]byte(jsonInput), &Themes)

	if err != nil {
		fmt.Println("JSON decode error!")
		return
	}
	// randomize the personality
	arr := []string{"valley_bro", "bill_and_ted", "your_excellency", "hh_holmes", "flying_spaghetti_monster", "norman_bates", "x_files", "jack_the_ripper"} // Shorthand declaration of array
	rand.Seed(time.Now().UTC().UnixNano())
	var randomInt int = randInt(0, 8)
	var random_theme string = arr[randomInt] 

	for index, theme := range Themes {
		fmt.Print(index + 1)
		if strings.TrimRight(random_theme, "\n") == theme.Name {
			f.Write([]byte(theme.Prompt.OpenAI)) //write directly into file
		}
	}

	f.Write([]byte(PROMPT_PREFIX)) //write directly into file
	for index, product := range products {
		var totals int = index + 1
		fmt.Print(totals)
		var active = int64(product["active"].(float64)) // in-stock units have positive active numbers
		if active > 0 { // it's in stock, adding to the list
			var name = product["name"].(string)
			f.Write([]byte(name)) //write directly into file
			f.WriteString(", ")
		}
	}
	openai_dinner()
	sendSlackmessage()
}
