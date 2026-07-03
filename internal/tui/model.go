package tui

import (
	"fmt"
	"strings"

	"io"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yogasimman/anjal/internal/env"
	"github.com/yogasimman/anjal/internal/models"
)

// Focus states.
type FocusState int

const (
	FocusSplash FocusState = iota
	FocusSidebar
	FocusRequest
	FocusResponse
	FocusAuth
	FocusCollections
	FocusMultiRun
	FocusEdit
	FocusNewCollection
)

type ReqTab int

const (
	ReqTabBody ReqTab = iota
	ReqTabHeaders
	ReqTabAuth
	ReqTabParams
)

type ResTab int

const (
	ResTabBody ResTab = iota
	ResTabHeaders
)

// AppModel is the single source of truth for the entire TUI.
type AppModel struct {
	// Data
	collections    []models.Collection
	activeColIndex int
	requests       []models.APIRequest
	activeRequest  *models.APIRequest
	response       *models.APIResponse
	err            error

	envVars map[string]string

	// UI components
	sidebar  list.Model
	viewport viewport.Model
	spinner  spinner.Model

	// Auth form
	showAuth     bool
	authType     textinput.Model
	authToken    textinput.Model
	authUsername textinput.Model
	authPassword textinput.Model
	authActive   int // which input is focused: 0=type, 1=token/username, 2=password

	// New Collection
	newColInput textinput.Model

	// Edit form
	editMethod  textinput.Model
	editTitle   textinput.Model
	editURL     textinput.Model
	editBody    textarea.Model
	editHeaders textarea.Model
	editParams  textarea.Model
	editActive  int
	reqSubFocus int // 0 = TopBar (Method/URL), 1 = TabContent (Body/Headers/Params)

	// Tabs
	reqTab ReqTab
	resTab ResTab

	// Layout
	width  int
	height int
	focus  FocusState

	isLoading bool

	// Splash
	splashTicks int

	// Multi-run
	selectedRequests map[string]bool
	multiRunResults  []models.APIResponse
	multiRunErrors   []error
	isMultiRunning   bool
	multiRunIndex    int

	startWithCollections bool
}

// requestItem wraps an APIRequest for the bubble list.
type requestItem struct {
	request  models.APIRequest
	index    int
	selected bool
}

func (r requestItem) Title() string {
	methodBadge := badgeGET
	switch r.request.Method {
	case "POST":
		methodBadge = badgePOST
	case "PUT":
		methodBadge = badgePUT
	case "PATCH":
		methodBadge = badgePATCH
	case "DELETE":
		methodBadge = badgeDELETE
	}
	
	methodStr := lipgloss.NewStyle().Width(10).Render(methodBadge.Render(r.request.Method))
	
	checkbox := "[ ]"
	if r.selected {
		checkbox = "[x]"
	}
	return fmt.Sprintf("%s %s %s", checkbox, methodStr, truncate(r.request.URL, 40))
}

func (r requestItem) Description() string {
	parts := []string{}
	if r.request.Title != "" {
		parts = append(parts, r.request.Title)
	}
	if r.request.Auth != nil {
		parts = append(parts, "🛡️ "+r.request.Auth.Type)
	}
	return strings.Join(parts, "  ")
}

func (r requestItem) FilterValue() string {
	return r.request.Method + " " + r.request.URL + " " + r.request.Title
}

// ---- Custom Delegate ----
type customDelegate struct{}

func (d customDelegate) Height() int                             { return 2 }
func (d customDelegate) Spacing() int                            { return 0 }
func (d customDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d customDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(requestItem)
	if !ok {
		return
	}

	title := i.Title()
	desc := i.Description()
	width := m.Width()

	if index == m.Index() {
		// Selected
		style := lipgloss.NewStyle().
			Background(lipgloss.Color("#313244")). // cSurface
			Foreground(lipgloss.Color("#CDD6F4")). // cText
			Width(width).
			Padding(0, 1)

		descStr := lipgloss.NewStyle().Foreground(lipgloss.Color("#CDD6F4")).Render(desc)
		fmt.Fprintf(w, style.Render(title+"\n"+descStr))
	} else {
		// Normal
		style := lipgloss.NewStyle().Padding(0, 1)
		fmt.Fprintf(w, style.Render(title+"\n"+lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086")).Render(desc)))
	}
}

// InitialModel builds the AppModel from parsed collections.
func InitialModel(collections []models.Collection, envVars map[string]string, isWorkspace bool) AppModel {
	// Auth inputs
	at := textinput.New()
	at.Placeholder = "bearer / basic / apikey / cookie"
	at.CharLimit = 20
	at.Width = 30

	tok := textinput.New()
	tok.Placeholder = "token value..."
	tok.CharLimit = 8192
	tok.Width = 50
	tok.EchoMode = textinput.EchoPassword

	un := textinput.New()
	un.Placeholder = "username"
	un.CharLimit = 50
	un.Width = 30

	pw := textinput.New()
	pw.Placeholder = "password"
	pw.CharLimit = 50
	pw.Width = 30
	pw.EchoMode = textinput.EchoPassword

	// Edit inputs
	em := textinput.New()
	em.Placeholder = "GET"
	em.Width = 10

	et := textinput.New()
	et.Placeholder = "Request Title"
	et.Width = 40

	eu := textinput.New()
	eu.Placeholder = "URL"
	eu.Width = 60

	eb := textarea.New()
	eb.Placeholder = "Request Body"
	eb.SetHeight(10)
	eb.SetWidth(60)

	eh := textarea.New()
	eh.Placeholder = "Key: Value\nAuthorization: Bearer token"
	eh.SetHeight(10)
	eh.SetWidth(60)

	ep := textarea.New()
	ep.Placeholder = "key=value\npage=1"
	ep.SetHeight(10)
	ep.SetWidth(60)

	// New Col Input
	nci := textinput.New()
	nci.Placeholder = "my_collection"
	nci.CharLimit = 50
	nci.Width = 40

	// Spinner
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#89B4FA")) // cBlue

	// Viewport
	vp := viewport.New(0, 0)
	vp.Style = lipgloss.NewStyle().Padding(0, 1)

	if envVars == nil {
		envVars = make(map[string]string)
	}

	m := AppModel{
		collections:  collections,
		envVars:      envVars,
		sidebar:      list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0),
		viewport:     vp,
		spinner:      sp,
		authType:     at,
		authToken:    tok,
		authUsername: un,
		authPassword: pw,
		newColInput:  nci,
		editMethod:   em,
		editTitle:    et,
		editURL:      eu,
		editBody:     eb,
		editHeaders:  eh,
		editParams:   ep,
		reqTab:       ReqTabBody,
		resTab:       ResTabBody,
		focus:        FocusSplash,
		selectedRequests: make(map[string]bool),
		splashTicks:  0,
		startWithCollections: isWorkspace,
	}

	m.rebuildSidebar()

	return m
}

func (m *AppModel) rebuildSidebar() {
	var items []list.Item
	var allReqs []models.APIRequest

	if m.activeColIndex >= 0 && m.activeColIndex < len(m.collections) {
		for i, req := range m.collections[m.activeColIndex].Requests {
			allReqs = append(allReqs, req)
			items = append(items, requestItem{
				request:  req,
				index:    i,
				selected: m.selectedRequests[req.ID], // USE ID NOW!
			})
		}
	}
	m.requests = allReqs

	// FIX: Use SetItems instead of list.New() to preserve scroll position!
	// We only initialize it if it's completely empty.
	if len(m.sidebar.Items()) == 0 && m.sidebar.Title == "" {
		m.sidebar = list.New(items, customDelegate{}, 0, 0)
		m.sidebar.SetShowStatusBar(false)
		m.sidebar.SetShowHelp(false)
		m.sidebar.SetFilteringEnabled(false)
		m.sidebar.Styles.Title = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#89B4FA")). // cBlue
			Padding(0, 1)
	} else {
		m.sidebar.SetItems(items)
	}

	// Safely update the activeRequest pointer
	if len(allReqs) > 0 {
		idx := m.sidebar.Index()
		if idx >= len(allReqs) {
			idx = len(allReqs) - 1
			m.sidebar.Select(idx)
		}
		m.activeRequest = &allReqs[idx]
	} else {
		m.activeRequest = nil
	}

	// Important: apply current window dimensions to the new list!
	if m.width > 0 && m.height > 0 {
		m.updateSizes()
	}
}

func (m *AppModel) loadEnvForActive() {
	if m.activeColIndex >= 0 && m.activeColIndex < len(m.collections) {
		vars, _ := env.LoadForCollection(m.collections[m.activeColIndex].Name)
		if vars != nil {
			for k, v := range vars {
				m.envVars[k] = v
			}
		}
	}
}

// resolveRequest substitutes {{.VAR}} placeholders and applies env auth fallback.
func resolveRequest(req *models.APIRequest, vars map[string]string) {
	req.URL = env.Resolve(req.URL, vars)
	req.Body = env.Resolve(req.Body, vars)
	for k, v := range req.Headers {
		req.Headers[k] = env.Resolve(v, vars)
	}
	for k, v := range req.QueryParams {
		req.QueryParams[k] = env.Resolve(v, vars)
	}
	
	// Force override Auth if workspace configured it
	authType := vars["WORKSPACE_AUTH_TYPE"]
	if authType != "" && authType != "none" {
		params := make(map[string]string)
		for k, v := range vars {
			if strings.HasPrefix(k, "WORKSPACE_AUTH_") && k != "WORKSPACE_AUTH_TYPE" {
				key := strings.ToLower(strings.TrimPrefix(k, "WORKSPACE_AUTH_"))
				params[key] = v
			}
		}
		req.Auth = &models.Auth{Type: authType, Params: params}
	} else if req.Auth != nil {
		for k, v := range req.Auth.Params {
			req.Auth.Params[k] = env.Resolve(v, vars)
		}
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
