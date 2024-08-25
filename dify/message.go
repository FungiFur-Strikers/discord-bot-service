package dify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type RequestBody struct {
	Inputs         map[string]interface{} `json:"inputs"`
	ConversationID string                 `json:"conversation_id"`
	Query          string                 `json:"query"`
	ResponseMode   string                 `json:"response_mode"`
	User           string                 `json:"user"`
}

type ResponseBody struct {
	Event          string `json:"event"`
	TaskID         string `json:"task_id"`
	ID             string `json:"id"`
	MessageID      string `json:"message_id"`
	ConversationID string `json:"conversation_id"`
	Mode           string `json:"mode"`
	Answer         string `json:"answer"`
}

func GenerateMessage(baseUrl string, token string, conversationId, query string) (*ResponseBody, error) {
	url := baseUrl + "/v1/chat-messages"
	method := "POST"

	body := RequestBody{
		Inputs:         map[string]interface{}{},
		ConversationID: conversationId,
		Query:          query,
		ResponseMode:   "blocking",
		User:           "user",
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status: %s, body: %s", resp.Status, string(respBody))
	}

	var response ResponseBody
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return &response, nil
}
