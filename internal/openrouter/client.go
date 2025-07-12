package openrouter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	apiKey    string
	modelName string
	client    *http.Client
}

func NewClient(apiKey, modelName string) *Client {
	return &Client{
		apiKey:    apiKey,
		modelName: modelName,
		client:    &http.Client{},
	}
}

func (c *Client) Query(messages []Message) (string, error) {
	requestBody := Request{
		Model:    c.modelName,
		Messages: messages,
	}

	requestBody.Provider.AllowFallbacks = true
	requestBody.Provider.RequireParameters = true
	requestBody.Provider.Order = []string{"DeepSeek"}
	requestBody.Provider.MaxPrice.Prompt = 0
	requestBody.Provider.MaxPrice.Completion = 0

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HTTP-Referer", "https://github.com")
	req.Header.Set("X-Title", "Telegram Bot")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status: %s", resp.Status)
	}

	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("error decoding response: %w", err)
	}

	if response.Error.Message != "" {
		return "", fmt.Errorf("API error: %s", response.Error.Message)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return response.Choices[0].Message.Content, nil
}
