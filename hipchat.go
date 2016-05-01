package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

var (
	defaultURL = "https://api.hipchat.com"
	notifyPath = "%s/v2/room/%s/notification?auth_token=%s"
)

// Client represents the HipChat client.
type Client struct {
	URL string
}

// Description represents the HipChat card description
type Description struct {
	Format string `json:"format"`
	Value  string `json:"value"`
}

// Activity represents the HipChat card activity
type Activity struct {
	HTML  string `json:"html"`
	Icon  string `json:"icon,omitempty"`
}

// Card represents the HipChat card
type Card struct {
	ID          string       `json:"id"`
	Style       string       `json:"style"`
	Format      string       `json:"format,omitempty"`
	Title       string       `json:"title"`
	URL         string       `json:"url"`
	Icon        *string      `json:"icon,omitempty"`
	Description *Description `json:"description,omitempty"`
	Activity    Activity     `json:"activity,omitempty"`
}

// Message represents the HipChat notification message.
type Message struct {
	From    string `json:"from"`
	Color   string `json:"color"`
	Notify  bool   `json:"notify"`
	Message string `json:"message"`
	Card    *Card  `json:"card,omitempty"`
}

// NewClient returns a new HipChat Client.
func NewClient(url, room, token string) *Client {
	if url == "" {
		url = defaultURL
	}

	return &Client{
		URL: fmt.Sprintf(notifyPath, url, room, token),
	}
}

// Send takes a HipChat notification message and sends it.
func (c *Client) Send(msg *Message) error {

	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	buf := bytes.NewReader(body)
	_, err = http.NewRequest("POST", c.URL, buf)
	if err != nil {
		return err
	}

	resp, err := http.Post(c.URL, "application/json", buf)
	if err != nil {
		return err
	}

	if resp.StatusCode >= http.StatusBadRequest {
		t, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return NewHipChatError(resp.StatusCode, string(t))
	}

	return nil
}

// HipChatError represents a HipChat error.
type HipChatError struct {
	Code int
	Body string
}

// NewHipChatError takes a code and body and returns a new *HipChatError.
func NewHipChatError(code int, body string) *HipChatError {
	return &HipChatError{Code: code, Body: body}
}

// Error implements the error interface.
func (e *HipChatError) Error() string {
	return fmt.Sprintf("HipChatError: %d %s", e.Code, e.Body)
}
