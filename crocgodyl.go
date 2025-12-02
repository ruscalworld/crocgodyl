package crocgodyl

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"iter"
	"net/http"
	"slices"
)

const Version = "1.0.0"

type Application struct {
	PanelURL string
	ApiKey   string
	Http     *http.Client
}

type Client struct {
	PanelURL string
	ApiKey   string
	Http     *http.Client
}

func NewApp(url, key string) (*Application, error) {
	if url == "" {
		return nil, errors.New("a valid panel url is required")
	}
	if key == "" {
		return nil, errors.New("a valid application api key is required")
	}

	app := &Application{
		PanelURL: url,
		ApiKey:   key,
		Http:     &http.Client{},
	}

	return app, nil
}

func (a *Application) newRequest(method, path string, body io.Reader) *http.Request {
	req, _ := http.NewRequest(method, fmt.Sprintf("%s/api/application%s", a.PanelURL, path), body)

	req.Header.Set("User-Agent", "Crocgodyl v"+Version)
	req.Header.Set("Authorization", "Bearer "+a.ApiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	return req
}

func NewClient(url, key string) (*Client, error) {
	if url == "" {
		return nil, errors.New("a valid panel url is required")
	}
	if key == "" {
		return nil, errors.New("a valid client api key is required")
	}

	client := &Client{
		PanelURL: url,
		ApiKey:   key,
		Http:     &http.Client{},
	}

	return client, nil
}

func (c *Client) newRequest(method, path string, body io.Reader) *http.Request {
	req, _ := http.NewRequest(method, fmt.Sprintf("%s/api/client%s", c.PanelURL, path), body)

	req.Header.Set("User-Agent", "Crocgodyl v"+Version)
	req.Header.Set("Authorization", "Bearer "+c.ApiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	return req
}

func validate(res *http.Response) ([]byte, error) {
	switch res.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusAccepted:
		buf, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}

		return buf, nil

	case http.StatusNoContent:
		return nil, nil

	default:
		buf, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}

		var errs *ApiError
		if err := json.Unmarshal(buf, &errs); err != nil {
			return nil, err
		}

		return nil, errs
	}
}

type Object[T any] struct {
	Object     string `json:"object"`
	Attributes T      `json:"attributes"`
}

type ObjectList[T any] struct {
	Object string      `json:"object"`
	Data   []Object[T] `json:"data"`
}

func (l *ObjectList[T]) IterObjects() iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, obj := range l.Data {
			if !yield(obj.Attributes) {
				return
			}
		}
	}
}

func (l *ObjectList[T]) Objects() []T {
	return slices.Collect(l.IterObjects())
}
