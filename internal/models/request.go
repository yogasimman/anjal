package models

import "time"

type Auth struct {
	Type   string
	Params map[string]string
}

type APIRequest struct {
	ID          string
	Title       string
	Method      string
	URL         string
	QueryParams map[string]string
	Headers     map[string]string
	Auth        *Auth
	Body        string
}

type APIResponse struct {
	StatusCode  int
	Status      string
	Body        string
	Latency     time.Duration
	Headers     map[string][]string
	ContentType string // json, xml, html, text, javascript, css, form, raw
}
