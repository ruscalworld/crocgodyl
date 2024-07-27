package crocgodyl

import (
	"encoding/json"
	"fmt"
	"time"
)

type Nest struct {
	ID          int        `json:"id"`
	UUID        string     `json:"uuid"`
	Author      string     `json:"author"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	CreatedAt   *time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

func (a *Application) GetNests() ([]*Nest, error) {
	req := a.newRequest("GET", "/nests", nil)
	res, err := a.Http.Do(req)
	if err != nil {
		return nil, err
	}

	buf, err := validate(res)
	if err != nil {
		return nil, err
	}

	var model struct {
		Data []struct {
			Attributes *Nest `json:"attributes"`
		} `json:"data"`
	}
	if err = json.Unmarshal(buf, &model); err != nil {
		return nil, err
	}

	nests := make([]*Nest, 0, len(model.Data))
	for _, n := range model.Data {
		nests = append(nests, n.Attributes)
	}

	return nests, nil
}

func (a *Application) GetNest(id int) (*Nest, error) {
	req := a.newRequest("GET", fmt.Sprintf("/nests/%d", id), nil)
	res, err := a.Http.Do(req)
	if err != nil {
		return nil, err
	}

	buf, err := validate(res)
	if err != nil {
		return nil, err
	}

	var model struct {
		Attributes Nest `json:"attributes"`
	}
	if err = json.Unmarshal(buf, &model); err != nil {
		return nil, err
	}

	return &model.Attributes, nil
}
