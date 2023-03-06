package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
)

const (
	// API for get available chatgpy models
	API_MODEL_LIST = "https://api.openai.com/v1/models"
	// API for chat completion
	API_CHAT_COMPLETION = "https://api.openai.com/v1/chat/completions"
)

// ChatGPT represents a ChatGPT object
type ChatGPT struct {
	// The apikey of your openapi account
	apiKey string
	// The model chat using.
	model string
	// The userid of a chat
	// will be applied to ChatBody
	userid string

	// The client we use
	c *http.Client
}

// NewChatGPT create a new chatgpt client
func NewChatGPT(apiKey, model, username string) (*ChatGPT, error) {
	// Create a resty client for getting informations
	// and to be attached to ChatGPT object for further use.
	proxyUrl, err := url.Parse(PROXY)
	var c *http.Client
	// Set proxy depending on url parsing result.
	if err == nil && PROXY != "" {
		c = &http.Client{
			Transport: &http.Transport{
				Proxy:               http.ProxyURL(proxyUrl),
				DisableKeepAlives:   false,
				MaxIdleConnsPerHost: 10,
			},
		}
		log.Println("Using Proxy:", proxyUrl.String())
	} else {
		c = &http.Client{}
	}

	log.Println("Get available model list")
	model_available := false
	for _, m := range GetModels(c) {
		if m.ID == model {
			model_available = true
			break
		}
	}

	if !model_available {
		return nil, fmt.Errorf("the model you chosen is not available")
	}

	return &ChatGPT{
		apiKey: apiKey,
		model:  "gpt-3.5-turbo",
		// Generate userid by username
		userid: GenUUID(username),
		c:      c,
	}, nil
}

// NewChat create new chat
func (c *ChatGPT) NewChat() *ChatContext {
	log.Println("ChatGPT client created with apikey:", c.apiKey)

	newChat := c.MakeChatContext()
	newChat.LoadChatMessageHistory() // load history data stored in filesystem if needed.

	return &newChat
}
