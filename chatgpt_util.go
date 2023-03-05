package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type Model struct {
	ID         string      `json:"id"`
	Object     string      `json:"object"`
	OwnedBy    string      `json:"owned_by"`
	Permission interface{} `json:"permission"`
}

// GetModels gets the models available to us
func GetModels(c *http.Client) []Model {
	req, err := http.NewRequest(http.MethodGet, API_MODEL_LIST, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Authorization", "Bearer "+API_KEY)

	resp, err := c.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var data map[string]interface{}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(respBody, &data)
	if err != nil {
		panic(err)
	}
	ok := 0
	var modelList []Model
	for _, md := range data["data"].([]interface{}) {
		item := md.(map[string]interface{})
		modelList = append(modelList, Model{
			ID:         item["id"].(string),
			Object:     item["object"].(string),
			OwnedBy:    item["owned_by"].(string),
			Permission: item["permission"],
		})
	}
	ok = 1
	defer func() {
		if ok != 1 {
			log.Println(string(respBody))
		}
	}()
	return modelList
}

// GenUUID generate a unique user id for chat use
func GenUUID(username string) string {
	uid := uuid.NewMD5(uuid.NameSpaceX500, []byte(username))
	return uid.String()
}
