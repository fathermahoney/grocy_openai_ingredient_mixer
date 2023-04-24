package main

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    openai "github.com/sashabaranov/go-openai"
    "github.com/slack-go/slack"
    // "github.com/joho/godotenv"
    "io"
    "io/ioutil"
    "math/rand"
    "net/http"
    "os"
    "strings"
    "time"
    "log"
    "strconv"
    "bytes"
    "bufio"
)

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

type Personality struct {
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

func main() {
    // Create necessary buffers
    var openai_buf bytes.Buffer
    openai_writer := bufio.NewWriter(&openai_buf)
    
    // Create request to get all products from Grocy
    GROCY_URL := os.Getenv("GROCY_URL")
    url := GROCY_URL
	log.Print(url)
    method := "GET"
    client := &http.Client{}
    req, err := http.NewRequest(method, url, nil)

    if err != nil {
        log.Fatal(err)
        return
    }
    GROCY_API_KEY := os.Getenv("GROCY_API_KEY")
    req.Header.Add("GROCY-API-KEY", GROCY_API_KEY)
    res, err := client.Do(req)
    if err != nil {
        log.Fatal(err)
        return
    }
    defer res.Body.Close()
    body, err := ioutil.ReadAll(res.Body)
    if err != nil {
        log.Fatal(err)
        return
    }
    
    var products []map[string]interface{} //defining the array of products
    err = json.Unmarshal([]byte(body), &products)
    
    if err != nil {
        log.Fatal("Error while decoding the data", err.Error())
    }

    // Load personality prompt file
    PERSONALITIES := ".personalities.json"
    jsonContent, err := ioutil.ReadFile(PERSONALITIES) // opens the inventory file just created
    
    if err != nil {
        log.Fatal("Err")
    }
    
    var jsonInput = string(jsonContent)
    var Personalities []Personality
    
    err = json.Unmarshal([]byte(jsonInput), &Personalities)

    if err != nil {
        log.Fatal("JSON decode error!")
        return
    }

    // get the length of the personalities object
    var personality_names []string
    for _, Personality := range Personalities {
        personality_names = append(personality_names, Personality.Name)
    }
    
    rand.Seed(time.Now().UTC().UnixNano())
    var randomInt int = randInt(0, len(Personalities))
    var random_theme string = personality_names[randomInt]
    var personality_choice string
    
    RANDOM_PROMPT, err := strconv.ParseBool(os.Getenv("RANDOM_PROMPT"))
    if RANDOM_PROMPT {
        personality_choice = random_theme
    } else {
        personality_choice = os.Getenv("STATIC_PROMPT")
    }

    for _, Personality := range Personalities {
        if strings.TrimRight(personality_choice, "\n") == Personality.Name {
            openai_writer.WriteString(string([]byte(Personality.Prompt.OpenAI)))
            openai_writer.Flush()
        }
    }
    openai_writer.WriteString(os.Getenv("INGREDIENT_LIST_HANDLING"))
    openai_writer.Flush()
    
    for _, product := range products {
        var active = int64(product["active"].(float64)) // in-stock units have positive active numbers
        if active > 0 { // it's in stock, adding to the list
            var name = product["name"].(string)
            openai_writer.WriteString(name+",")
            openai_writer.Flush()
        }
    }
    openai_ingredient_mixer(&openai_buf)
}

func openai_ingredient_mixer(openai_buf *bytes.Buffer) {
    var slack_buf bytes.Buffer
    writer := bufio.NewWriter(&slack_buf)
    headline := fmt.Sprintf("*Tanglesprings Specials for %s %s %d, %d*\n", weekday, month, day, year)
    writer.WriteString(headline)
    writer.Flush()
        
    var openaiPrompt string = openai_buf.String()
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
        log.Printf("ChatCompletionStream error: %v\n", err)
        return
    }
    defer stream.Close()
    
    for {
        response, err := stream.Recv()    
        if errors.Is(err, io.EOF) {
            break
        }

        if err != nil {
            log.Printf("\nStream error: %v\n", err)
            return
        }
        var responseContent = response.Choices[0].Delta.Content
        writer.WriteString(responseContent)
        writer.Flush()
    }
    fmt.Println(slack_buf.String())
    sendSlackmessage(&slack_buf)
}

func sendSlackmessage(slack_buf *bytes.Buffer) {
    fmt.Println(slack_buf.String())
    var slackMessage = slack_buf.String()
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
        log.Printf("%s", err)
        return
    }
    log.Printf("Ingredients mixed and Slack message successfully sent to channel %s at %s", channelID, timestamp)
}