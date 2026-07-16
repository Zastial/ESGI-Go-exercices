package mira

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) Search(ctx context.Context, query string, limit int) ([]Note, error) {
	var envelope struct {
		Data []Note `json:"data"`
	}
	values := url.Values{}
	values.Set("q", query)
	if limit > 0 {
		values.Set("limit", fmt.Sprint(limit))
	}
	path := "/search?" + values.Encode()
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &envelope); err != nil {
		return nil, err
	}
	return envelope.Data, nil
}

func (c *Client) Get(ctx context.Context, id string) (Note, error) {
	var note Note
	if err := c.doJSON(ctx, http.MethodGet, "/notes/"+id, nil, &note); err != nil {
		return Note{}, err
	}
	return note, nil
}

func (c *Client) Create(ctx context.Context, input CreateNoteInput) (Note, error) {
	var note Note
	if err := c.doJSON(ctx, http.MethodPost, "/notes", input, &note); err != nil {
		return Note{}, err
	}
	return note, nil
}

func (c *Client) List(ctx context.Context, limit int) ([]Note, error) {
	var envelope struct {
		Data []Note `json:"data"`
	}
	values := url.Values{}
	if limit > 0 {
		values.Set("limit", fmt.Sprint(limit))
	}
	path := "/notes"
	if encoded := values.Encode(); encoded != "" {
		path += "?" + encoded
	}
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &envelope); err != nil {
		return nil, err
	}
	return envelope.Data, nil
}

func (c *Client) doJSON(ctx context.Context, method, path string, body any, out any) error {
	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		buf, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(buf)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errBody struct {
			Error   string `json:"error"`
			Message string `json:"message"`
		}
		json.NewDecoder(resp.Body).Decode(&errBody)
		if errBody.Message != "" {
			return fmt.Errorf("http %d: %s", resp.StatusCode, errBody.Message)
		}
		return fmt.Errorf("http %d", resp.StatusCode)
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
