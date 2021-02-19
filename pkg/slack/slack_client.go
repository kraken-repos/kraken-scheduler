package slack

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

const DefaultSlackTimeout = 5 * time.Second

type Client struct {
	WebHookURL string
	Timeout    time.Duration
}

type Message struct {
	Text string `json:"text,omitempty"`
}

func (sc *Client) SendMessage(text string) error {
	sm := Message{
		Text: text,
	}
	return sc.sendHttpRequest(&sm)
}

func (sc *Client) sendHttpRequest(sm *Message) error {
	slackBody, err := json.Marshal(sm)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, sc.WebHookURL, bytes.NewBuffer(slackBody))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	if sc.Timeout == 0 {
		sc.Timeout = DefaultSlackTimeout
	}

	client := &http.Client{Timeout: sc.Timeout}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return err
	}

	if buf.String() != "ok" {
		return errors.New("Non-ok response returned from Slack")
	}

	return nil
}