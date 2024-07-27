package crocgodyl

import (
	"encoding/json"
	"fmt"
	"time"
)

type Egg struct {
	ID           int               `json:"id"`
	UUID         string            `json:"uuid"`
	Name         string            `json:"name"`
	Nest         int               `json:"nest"`
	Author       string            `json:"author"`
	Description  string            `json:"description"`
	DockerImage  string            `json:"docker_image"`
	DockerImages map[string]string `json:"docker_images"`
	Config       struct {
		Files   map[string]any `json:"files"`
		Startup struct {
			Done            string   `json:"done"`
			UserInteraction []string `json:"userInteraction"`
		} `json:"startup"`
		Stop string `json:"stop"`
		Logs struct {
			Custom   bool   `json:"custom"`
			Location string `json:"location"`
		} `json:"logs"`
	} `json:"config"`
	Startup string `json:"startup"`
	Script  struct {
		Privileged bool   `json:"privileged"`
		Install    string `json:"install"`
		Entry      string `json:"entry"`
		Container  string `json:"container"`
	} `json:"script"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

func (a *Application) GetEggs(nest int) ([]*Egg, error) {
	req := a.newRequest("GET", fmt.Sprintf("/nests/%d/eggs", nest), nil)
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
			Attributes *Egg `json:"attributes"`
		} `json:"data"`
	}
	if err = json.Unmarshal(buf, &model); err != nil {
		return nil, err
	}

	eggs := make([]*Egg, 0, len(model.Data))
	for _, n := range model.Data {
		eggs = append(eggs, n.Attributes)
	}

	return eggs, nil
}

func (a *Application) GetEgg(nest, id int) (*Egg, error) {
	req := a.newRequest("GET", fmt.Sprintf("/nests/%d/eggs/%d", nest, id), nil)
	res, err := a.Http.Do(req)
	if err != nil {
		return nil, err
	}

	buf, err := validate(res)
	if err != nil {
		return nil, err
	}

	var model struct {
		Attributes Egg `json:"attributes"`
	}
	if err = json.Unmarshal(buf, &model); err != nil {
		return nil, err
	}

	return &model.Attributes, nil
}

type EggVariable struct {
	ID           int        `json:"id"`
	Egg          int        `json:"egg"`
	Name         string     `json:"name"`
	Description  string     `json:"description"`
	EnvVariable  string     `json:"env_variable"`
	DefaultValue string     `json:"default_value"`
	Rules        string     `json:"rules"`
	UserViewable bool       `json:"user_viewable"`
	UserEditable bool       `json:"user_editable"`
	CreatedAt    *time.Time `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at,omitempty"`
}

func (a *Application) GetEggVariables(nest, id int) ([]*EggVariable, error) {
	req := a.newRequest("GET", fmt.Sprintf("/nests/%d/eggs/%d?include=variables", nest, id), nil)
	res, err := a.Http.Do(req)
	if err != nil {
		return nil, err
	}

	buf, err := validate(res)
	if err != nil {
		return nil, err
	}

	var model struct {
		Attributes struct {
			Relationships struct {
				Variables struct {
					Data []struct {
						Attributes *EggVariable `json:"attributes"`
					} `json:"data"`
				} `json:"variables"`
			} `json:"relationships"`
		} `json:"attributes"`
	}
	if err = json.Unmarshal(buf, &model); err != nil {
		return nil, err
	}

	variables := make([]*EggVariable, 0, len(model.Attributes.Relationships.Variables.Data))
	for _, n := range model.Attributes.Relationships.Variables.Data {
		variables = append(variables, n.Attributes)
	}

	return variables, nil
}
