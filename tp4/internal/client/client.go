package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"mira/internal/core"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) Create(ctx context.Context, input core.CreateNoteInput) (core.Note, error) {
	var note core.Note
	if err := c.doJSON(ctx, http.MethodPost, "/notes", input, &note); err != nil {
		return core.Note{}, err
	}
	return note, nil
}

func (c *Client) List(ctx context.Context, params core.ListParams) ([]core.Note, error) {
	var envelope struct {
		Data []core.Note `json:"data"`
	}
	query := url.Values{}
	if params.Limit > 0 {
		query.Set("limit", fmt.Sprint(params.Limit))
	}
	if params.Offset > 0 {
		query.Set("offset", fmt.Sprint(params.Offset))
	}
	path := "/notes"
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &envelope); err != nil {
		return nil, err
	}
	return envelope.Data, nil
}

func (c *Client) Search(ctx context.Context, query string, params core.ListParams) ([]core.Note, error) {
	var envelope struct {
		Data []core.Note `json:"data"`
	}
	values := url.Values{}
	values.Set("q", query)
	if params.Limit > 0 {
		values.Set("limit", fmt.Sprint(params.Limit))
	}
	if params.Offset > 0 {
		values.Set("offset", fmt.Sprint(params.Offset))
	}
	if err := c.doJSON(ctx, http.MethodGet, "/search?"+values.Encode(), nil, &envelope); err != nil {
		return nil, err
	}
	return envelope.Data, nil
}

func (c *Client) Update(ctx context.Context, id string, input core.UpdateNoteInput) (core.Note, error) {
	var note core.Note
	if err := c.doJSON(ctx, http.MethodPatch, "/notes/"+id, input, &note); err != nil {
		return core.Note{}, err
	}
	return note, nil
}

func (c *Client) Delete(ctx context.Context, id string) error {
	return c.doJSON(ctx, http.MethodDelete, "/notes/"+id, nil, nil)
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
		return fmt.Errorf("http %d", resp.StatusCode)
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
