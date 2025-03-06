package crocgodyl

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
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

func (a *Client) newRequest(method, path string, body io.Reader) *http.Request {
	req, _ := http.NewRequest(method, fmt.Sprintf("%s/api/client%s", a.PanelURL, path), body)

	req.Header.Set("User-Agent", "Crocgodyl v"+Version)
	req.Header.Set("Authorization", "Bearer "+a.ApiKey)
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
