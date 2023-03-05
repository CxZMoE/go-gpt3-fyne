package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"time"
)

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
			recvMsg, err := newChat.Send(sc.Text(), nil)
			if err != nil {
				panic(err)
			}
			if !newChat.Options.IsStreamming {

				fmt.Printf("[%s]\n%s\n\n%s\n\n", recvMsg.id, time.Now().Format(time.RFC1123), recvMsg.Content)
			}
		}
	}
}
