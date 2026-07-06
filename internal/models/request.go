// Copyright (c) 2026 Yogasimman Ravisagar
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

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

// Collection represents a single Markdown file containing multiple API requests.
// Auth is the collection-level authorization — if a request has no Auth of its
// own, it falls back to this one (cascading inheritance).
type Collection struct {
	Name     string
	FilePath string
	Auth     *Auth
	Requests []APIRequest
}
