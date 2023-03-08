package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

var API_KEY string
var DEFAULT_MODEL string
var DEFAULT_USERNAME string
var CAPACITY int
var PROXY string

var logFile *os.File

func init() {
	wd, _ := os.Getwd()
	fd, err := os.OpenFile(fmt.Sprintf("%s/log-%d.log", wd, time.Now().Unix()), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	logFile = fd
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))

	// Check ttf file
	ttf_path := wd + "/FONT.TTF"
	_, err = os.Stat(ttf_path)
	if !os.IsNotExist(err) {
		err := os.Setenv("FYNE_FONT", ttf_path)
		if err != nil {
			panic(err)
		}
	}

	// Read config
	config_path := wd + "/auth.json"
	_, err = os.Stat(config_path)
	if os.IsNotExist(err) {
		log.Fatalln("Config file", config_path, "does not exist.")
	}
	cofigFile, err := os.Open(config_path)
	if err != nil {
		panic(err)
	}
	defer cofigFile.Close()
	data, err := io.ReadAll(cofigFile)
	if err != nil {
		panic(err)
	}
	configJson := make(map[string]interface{})
	json.Unmarshal(data, &configJson)

	// OpenAI API Key
	apiKey := configJson["apiKey"]
	if apiKey == nil {
		log.Fatalln("apiKey is not configured.")
	} else {
		API_KEY = apiKey.(string)
	}

	// GPT Model Name
	model := configJson["model"]
	if model == nil {
		log.Fatalln("gpt model name is not configured.")
	} else {
		DEFAULT_MODEL = model.(string)
	}

	// Chat Username
	username := configJson["username"]
	if username == nil {
		DEFAULT_USERNAME = GenUUID(fmt.Sprintf("guest-%d", time.Now().Unix()))
	} else {
		DEFAULT_USERNAME = username.(string)
	}

	// Message Capacity
	capacity := configJson["capacity"]
	if capacity == nil {
		CAPACITY = 20
	} else {
		CAPACITY = int(capacity.(float64))
	}

	// Proxy Address:Port
	proxy := configJson["proxy"]
	if proxy == nil {
		PROXY = os.Getenv("HTTP_PROXY")
	} else {
		PROXY = proxy.(string)
	}

	log.Println("Start with config:")
	log.Println("apiKey:", API_KEY)
	log.Println("model:", DEFAULT_MODEL)
	log.Println("username:", DEFAULT_USERNAME)
	log.Println("capacity:", CAPACITY)
}

// CLI start a cli version
func CLI() {
	// Create a ChatGPT client
	gpt, err := NewChatGPT(API_KEY, DEFAULT_MODEL, DEFAULT_USERNAME)
	if err != nil {
		panic(err)
	}
	log.Println("ChatGPT client created with apikey:", gpt.apiKey)

	newChat := gpt.MakeChatContext()
	newChat.LoadChatMessageHistory() // load history data stored in filesystem if needed.

	sc := bufio.NewScanner(os.Stdin)
	sc.Split(bufio.ScanLines)
	for {
		fmt.Print("==> ")
		if sc.Scan() {
			fmt.Printf("<== [%s]\n", time.Now().Format(time.RFC1123))
			recvMsg, err := newChat.Send(sc.Text())
			if err != nil {
				panic(err)
			}
			if !newChat.Options.IsStreamming {
				fmt.Printf("[%s]\n%s\n\n%s\n\n", recvMsg.id, time.Now().Format(time.RFC1123), recvMsg.Content)
			}
		}
	}
}

func main() {
	CLI()
}
