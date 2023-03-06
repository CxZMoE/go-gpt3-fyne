package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"fyne.io/fyne/v2/widget"
)

/*
	Conventions:
		Chat Completion => Chat
*/

// Chat Roles
const (
	// The system message helps set the behavior of the assistant.
	CHAT_ROLE_SYSTEM = "system"
	// The assistant messages help store prior responses.
	CHAT_ROLE_ASSIST = "assistant"
	// The user messages help instruct the assistant.
	CHAT_ROLE_USER = "user"
)

// Where we store chat contexts
var ChatPool = make(map[string]ChatContext)

// ChatMessage represents a piece of chat message
type ChatMessage struct {
	id      string
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatOptions sets options for a chat
type ChatOptions struct {
	// ID of the model we using. eg.gpt-3.5-turbo
	Model string `json:"model"`

	// What sampling temperature to use, between 0 and 2.
	// Higher values like 0.8 will make the output more random,
	// while lower values like 0.2 will make it more focused and deterministic.
	// We generally recommend altering this or top_p but not both.
	Temperature int `json:"temperature"`

	// An alternative to sampling with temperature,
	// called nucleus sampling,
	// where the model considers the results of the tokens with top_p probability mass.
	// So 0.1 means only the tokens comprising the top 10% probability mass are considered.
	TopP int `json:"top_p"`

	// How many chat completion choices to generate for each input message.
	Choices int `json:"n"`

	// If set, partial message deltas will be sent,
	// like in ChatGPT.
	// Tokens will be sent as data-only server-sent events as they become available,
	// with the stream terminated by a data: [DONE] message.
	IsStreamming bool `json:"stream"`

	// Up to 4 sequences where the API will stop generating further tokens.
	Stops []string `json:"stop"`

	// The maximum number of tokens allowed for the generated answer. By default,
	// the number of tokens the model can return will be (4096 - prompt tokens).
	// MaxTokens int `json:"max_tokens"`

	// Number between -2.0 and 2.0.
	// Positive values penalize new tokens based on whether they appear in the text so far,
	// increasing the model's likelihood to talk about new topics.
	PresencePenalty int `json:"presence_penalty"`

	// Number between -2.0 and 2.0.
	// Positive values penalize new tokens based on their existing frequency in the text so far,
	// decreasing the model's likelihood to repeat the same line verbatim.
	FrequencyPenalty int `json:"frequency_penalty"`

	// A unique identifier representing your end-user,
	// which can help OpenAI to monitor and detect abuse.
	User string `json:"user"`
}

func DefaultChatOptions() ChatOptions {
	return ChatOptions{
		Model:        DEFAULT_MODEL,
		Temperature:  1,
		TopP:         1,
		Choices:      1,
		IsStreamming: false,
		Stops:        nil,
		// MaxTokens:        4096,
		PresencePenalty:  0,
		FrequencyPenalty: 0,
		User:             "",
	}
}

// ChatBody the request body of a chat completion reqeust.
type ChatBody struct {
	// Messages stores prior responses
	Messages []ChatMessage `json:"messages"`
	// Options
	ChatOptions
}

// ChatContext contains chat id and chat message history
type ChatContext struct {
	// Conversation ID
	ID string
	// Chat history
	History []ChatMessage
	// Chat options
	Options ChatOptions
	// Chat http client
	Client *http.Client

	// full content
	content string
}

// MakeChatContext makes a new chat context
// contains a chat conversation id
func (gpt *ChatGPT) MakeChatContext() ChatContext {
	userId := gpt.userid
	// Try delete the user of same id,
	// This should be done more elegently in the fure.
	delete(ChatPool, userId)

	chatContext := ChatContext{
		// Generate UUID by userid and timestamp
		ID:      GenUUID(fmt.Sprintf("xgpt%s-%d", userId, time.Now().Unix())),
		History: make([]ChatMessage, 0),
		Options: DefaultChatOptions(),
		Client:  gpt.c,
	}

	chatContext.Options.IsStreamming = true
	// Set user
	chatContext.Options.User = userId
	ChatPool[userId] = chatContext

	return chatContext
}

func (c *ChatContext) LoadChatMessageHistory() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	fd := wd + "/session.dat"

	_, err = os.Stat(fd)
	if os.IsNotExist(err) {
		return
	}

	log.Println("Load last session data.")
	f, err := os.Open(fd)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	data, _ := io.ReadAll(f)
	json.Unmarshal(data, &c.History)
	log.Println("Loaded length:", len(c.History))
}

func (c *ChatContext) CleanChatMessageHistory() {
	// Only left last history
	c.History = c.History[len(c.History)-1:]
	c.SyncChatMessageHistory()
}

func (c *ChatContext) SyncChatMessageHistory() {
	wd, _ := os.Getwd()
	f, err := os.OpenFile(wd+"/session.dat", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	dat, err := json.Marshal(c.History)
	if err != nil {
		panic(err)
	}
	f.Write(dat)
	f.Sync()
}

// AddChatMessageHistory add chat message to history of context
func (c *ChatContext) AddChatMessageHistory(appendMsgs ...ChatMessage) {
	for _, cm := range appendMsgs {
		log.Println("Add message of role:", cm.Role, "content:", cm.Content)
	}
	// remove history index 0 when history length == 3
	c.History = append(c.History, appendMsgs...)
	if len(c.History) > CAPACITY {
		c.History = c.History[len(c.History)-CAPACITY:]
	}
	c.SyncChatMessageHistory()
}

// Send send a new message to server
func (c *ChatContext) Send(msg string, rt *widget.Entry) (*ChatMessage, error) {

	// Add new msg to history firstly
	c.AddChatMessageHistory(ChatMessage{
		Role:    CHAT_ROLE_USER,
		Content: msg,
	})
	c.content += "==> " + msg + "\n"

	// Create a reqeust body
	chatBody := ChatBody{
		Messages:    c.History,
		ChatOptions: c.Options,
	}

	// Start a new request
	reqBodyData, err := json.Marshal(chatBody)
	if err != nil {
		panic(err)
	}
	req, err := http.NewRequest(http.MethodPost, API_CHAT_COMPLETION, bytes.NewReader(reqBodyData))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Authorization", "Bearer "+API_KEY)
	req.Header.Set("Content-Type", "application/json")
	if c.Options.IsStreamming {
		req.Header.Set("Connection", "keep-alive")
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		// pop the last msg
		c.History = c.History[:len(c.History)-1]
		return nil, fmt.Errorf("send chat body failed: %s", err.Error())
	}
	defer resp.Body.Close()

	// err status code
	if resp.StatusCode != http.StatusOK {
		buf, _ := io.ReadAll(resp.Body)
		log.Println("resp.Body() =>\n", string(buf))
		log.Println("resp.Body() =>END")
		resp.Body.Close()
		if resp.StatusCode == 400 {
			log.Println("Token's too long, clear token.")
			c.CleanChatMessageHistory()
		}
		return nil, fmt.Errorf("chat failed with status code: %d\nbody: %s", resp.StatusCode, string(buf))
	}

	if c.Options.IsStreamming {
		sc := bufio.NewReader(resp.Body)
		index := 0
		var message ChatMessage
		for {
			// if resp.Header.Get("Connection") != "keep-alive" {
			// 	log.Println(resp.Header.Get("Connection"))
			// }
			buf, _, err := sc.ReadLine()
			if err != nil {
				if err == io.EOF {
					return c.Send(msg, rt)
				}
				return nil, err
			}
			if len(buf) == 0 {
				continue
			}

			tokens := strings.Split(string(buf), "data: ")[1]

			// Exit when received [DONE]
			if strings.Contains(tokens, "[DONE]") {
				c.AddChatMessageHistory(message)
				fmt.Println()
				c.content += "\n"
				break
			}

			// Parse Data
			var data = make(map[string]interface{})
			err = json.Unmarshal([]byte(tokens), &data)
			if err != nil {
				panic(fmt.Sprintf("err: %s\n data: %s", err.Error(), string(tokens)))
			}
			delta_data := data["choices"].([]interface{})[0].(map[string]interface{})["delta"].(map[string]interface{})

			// first data shold be role=
			if index == 0 {
				id := data["id"]
				role := delta_data["role"]
				if id != nil && role != nil {
					message.id = id.(string)
					message.Role = role.(string)
				} else {
					return nil, fmt.Errorf("role is nil")
				}
			} else {
				content := delta_data["content"]
				if content != nil {
					delta_content := content.(string)
					message.Content += delta_content
					// show content
					// fmt.Print(delta_content)
					c.content += delta_content
					rt.CursorRow = 65536
					rt.SetText(c.content)

				} else {
					continue
				}
			}
			index += 1
		}
		return &message, nil
	}

	log.Println("selecting no stream mode")
	// [200] OK
	var respBody map[string]interface{}
	respBodyData, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(respBodyData, &respBody)
	if err != nil {
		panic(err)
	}

	message := c.ParseData(respBody)
	return &message, nil
}

func (c *ChatContext) ParseData(respBody map[string]interface{}) ChatMessage {
	choices := respBody["choices"].([]interface{})

	// The message of choice 0
	var message ChatMessage

	if len(choices) > 0 {
		choice := choices[0].(map[string]interface{})
		message.Role = choice["message"].(map[string]interface{})["role"].(string)
		message.Content = choice["message"].(map[string]interface{})["content"].(string)

		c.AddChatMessageHistory(message)
	} else {
		panic("choices of length 0")
	}
	message.id = respBody["id"].(string)

	return message
}
