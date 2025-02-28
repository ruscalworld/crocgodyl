package crocgodyl

import (
	"fmt"
	"strings"
)

type Error struct {
	Code   string      `json:"code"`
	Status string      `json:"status"`
	Detail string      `json:"detail"`
	Meta   interface{} `json:"meta,omitempty"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s (%s): %s", e.Status, e.Code, e.Detail)
}

type ApiError struct {
	Errors []*Error `json:"errors"`
}

func (e *ApiError) Error() string {
	sb := &strings.Builder{}

	for _, err := range e.Errors {
		sb.WriteString(fmt.Sprintf("\t - %s\n", err.Error()))
	}

	return fmt.Sprintf("API returned %d errors:\n%s", len(e.Errors), sb.String())
}
