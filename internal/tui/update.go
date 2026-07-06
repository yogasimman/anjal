// Copyright (c) 2026 Yogasimman Ravisagar
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package tui

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yogasimman/anjal/internal/env"
	"github.com/yogasimman/anjal/internal/httpclient"
	"github.com/yogasimman/anjal/internal/models"
	"github.com/yogasimman/anjal/internal/parser"
	"time"
)

// ---- Messages ----

type responseMsg struct {
	response models.APIResponse
	err      error
}

type authSavedMsg struct {
	collection string
	authType   string
	token      string
}

type multiRunNextMsg struct{}
type multiRunDoneMsg struct{}


// ---- Init ----

func (m AppModel) Init() tea.Cmd {
	return m.spinner.Tick
}

// ---- Update ----

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	// FIX 1: Pass generic background messages (like BlinkMsg) to ALL inputs
	// so the UI doesn't freeze! We ignore KeyMsgs so they don't all type at once.
	if _, isKeyMsg := msg.(tea.KeyMsg); !isKeyMsg {
		m.editMethod, cmd = m.editMethod.Update(msg); cmds = append(cmds, cmd)
		m.editTitle, cmd = m.editTitle.Update(msg); cmds = append(cmds, cmd)
		m.editURL, cmd = m.editURL.Update(msg); cmds = append(cmds, cmd)
		m.editBody, cmd = m.editBody.Update(msg); cmds = append(cmds, cmd)
		m.editHeaders, cmd = m.editHeaders.Update(msg); cmds = append(cmds, cmd)
		m.editParams, cmd = m.editParams.Update(msg); cmds = append(cmds, cmd)
		
		m.authType, cmd = m.authType.Update(msg); cmds = append(cmds, cmd)
		m.authToken, cmd = m.authToken.Update(msg); cmds = append(cmds, cmd)
		m.authUsername, cmd = m.authUsername.Update(msg); cmds = append(cmds, cmd)
		m.authPassword, cmd = m.authPassword.Update(msg); cmds = append(cmds, cmd)
		m.newColInput, cmd = m.newColInput.Update(msg); cmds = append(cmds, cmd)
	}

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateSizes()

	case spinner.TickMsg:
		if m.focus == FocusSplash {
			m.splashTicks++
			if m.splashTicks > 15 {
				if m.startWithCollections {
					m.focus = FocusCollections
				} else {
					m.focus = FocusSidebar
				}
			}
			var spCmd tea.Cmd
			m.spinner, spCmd = m.spinner.Update(msg)
			return m, spCmd
		}
		if m.isLoading {
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case tea.KeyMsg:
		// 1. Check if we need to exit entirely.
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		// 2. If Auth form is open, it steals almost all key presses.
		if m.showAuth {
			if msg.String() == "esc" {
				m.showAuth = false
				m.focus = FocusSidebar
				return m, nil
			}
			return m.handleAuthKeys(msg)
		}
		
		if m.focus == FocusNewCollection {
			return m.handleNewCollectionKeys(msg)
		}

		// 3. Mode-specific overrides (must come before global letter hotkeys!)
		if m.focus == FocusCollections {
			return m.handleCollectionKeys(msg)
		}
		
		if m.focus == FocusEdit {
			return m.handleEditKeys(msg)
		}

		// 4. Global hotkeys (when NOT in auth/text input mode)
		switch msg.String() {
		case "q":
			return m, tea.Quit

		case "c":
			m.focus = FocusCollections
			return m, nil

		case "a", "f3":
			m.showAuth = true
			m.focus = FocusAuth
			m.authActive = 0
			
			if v, ok := m.envVars["WORKSPACE_AUTH_TYPE"]; ok {
				m.authType.SetValue(v)
			} else {
				m.authType.SetValue("")
			}
			
			if v, ok := m.envVars["WORKSPACE_AUTH_TOKEN"]; ok {
				m.authToken.SetValue(v)
			} else {
				m.authToken.SetValue("")
			}
			
			if v, ok := m.envVars["WORKSPACE_AUTH_USERNAME"]; ok {
				m.authUsername.SetValue(v)
			} else {
				m.authUsername.SetValue("")
			}
			
			if v, ok := m.envVars["WORKSPACE_AUTH_PASSWORD"]; ok {
				m.authPassword.SetValue(v)
			} else {
				m.authPassword.SetValue("")
			}

			m.authType.Focus()
			return m, nil
		}

		switch msg.String() {
		case "e":
			if m.activeRequest != nil && m.focus == FocusRequest {
				m.focus = FocusEdit
				if m.reqSubFocus == 0 {
					m.editActive = 0
					m.editMethod.SetValue(m.activeRequest.Method)
					m.editTitle.SetValue(m.activeRequest.Title)
					m.editURL.SetValue(m.activeRequest.URL)
					m.editMethod.Focus()
					m.editTitle.Blur()
					m.editURL.Blur()
				} else {
					switch m.reqTab {
					case ReqTabBody:
						m.editActive = 3
						m.editBody.SetValue(m.activeRequest.Body)
						m.editBody.Focus()
					case ReqTabHeaders:
						m.editActive = 4
						m.editHeaders.SetValue(mapToString(m.activeRequest.Headers, ":"))
						m.editHeaders.Focus()
					case ReqTabParams:
						m.editActive = 5
						m.editParams.SetValue(mapToString(m.activeRequest.QueryParams, "="))
						m.editParams.Focus()
					case ReqTabAuth:
						m.showAuth = true
						m.focus = FocusAuth
						m.authActive = 0
						m.authType.SetValue(m.envVars["WORKSPACE_AUTH_TYPE"])
						m.authToken.SetValue(m.envVars["WORKSPACE_AUTH_TOKEN"])
						m.authUsername.SetValue(m.envVars["WORKSPACE_AUTH_USERNAME"])
						m.authPassword.SetValue(m.envVars["WORKSPACE_AUTH_PASSWORD"])
						m.authType.Focus()
					}
				}
			}
			return m, nil
			
		case "n": // New request
			if m.activeColIndex < len(m.collections) {
				col := m.collections[m.activeColIndex]
				newReq := models.APIRequest{
					Title:  fmt.Sprintf("New Request %d", time.Now().Unix()),
					Method: "GET",
					URL:    "https://api.example.com",
				}
				err := parser.AddRequest(col.FilePath, newReq)
				if err != nil {
					m.err = fmt.Errorf("Failed to create request: %v", err)
				} else {
					cols, _ := parser.LoadWorkspace()
					if len(cols) > 0 {
						m.collections = cols
						m.rebuildSidebar()
						lastIdx := len(m.requests) - 1
						if lastIdx >= 0 {
							m.sidebar.Select(lastIdx)
							m.onSidebarChange()
						}
						m.focus = FocusEdit
						m.reqSubFocus = 0
						m.editActive = 2
						if m.activeRequest != nil {
							m.editMethod.SetValue(m.activeRequest.Method)
							m.editTitle.SetValue(m.activeRequest.Title)
							m.editURL.SetValue(m.activeRequest.URL)
						}
						m.editMethod.Blur()
						m.editTitle.Blur()
						m.editURL.Focus()
					}
				}
			}
			return m, nil
			
			return m, nil

		case "d", "delete":
			if m.focus == FocusSidebar && m.activeRequest != nil {
				col := m.collections[m.activeColIndex]
				err := parser.DeleteRequest(col.FilePath, m.activeRequest.ID)
				if err != nil {
					m.err = fmt.Errorf("Failed to delete request: %v", err)
				} else {
					cols, _ := parser.LoadWorkspace()
					if len(cols) > 0 {
						m.collections = cols
						m.rebuildSidebar()
					}
				}
			}
			return m, nil
		case "D":
			if m.focus == FocusSidebar && m.singleFilePath == "" && len(m.collections) > 0 {
				col := m.collections[m.activeColIndex]
				err := os.Remove(col.FilePath)
				if err != nil {
					m.err = fmt.Errorf("Failed to delete collection: %v", err)
				} else {
					cols, _ := parser.LoadWorkspace()
					m.collections = cols
					if m.activeColIndex >= len(m.collections) {
						m.activeColIndex = len(m.collections) - 1
					}
					if m.activeColIndex < 0 {
						m.activeColIndex = 0
					}
					m.activeRequest = nil
					m.response = nil
					m.rebuildSidebar()
				}
			}
			return m, nil
		case "y":
			if m.response != nil {
				clipboard.WriteAll(m.response.Body)
			}
			return m, nil

		case "Y":
			if m.activeRequest != nil {
				clipboard.WriteAll(m.activeRequest.URL)
			}
			return m, nil

		case "tab":
			if m.focus == FocusSidebar {
				m.focus = FocusRequest
			} else if m.focus == FocusRequest {
				m.focus = FocusResponse
			} else {
				m.focus = FocusSidebar
			}
			return m, nil

		case "shift+tab":
			if m.focus == FocusSidebar {
				m.focus = FocusResponse
			} else if m.focus == FocusRequest {
				m.focus = FocusSidebar
			} else {
				m.focus = FocusRequest
			}
			return m, nil

		case "left", "h":
			if m.focus == FocusRequest {
				m.reqTab = (m.reqTab - 1 + 4) % 4
			} else if m.focus == FocusSidebar {
				m.reqTab = (m.reqTab - 1 + 4) % 4
			} else if m.focus == FocusResponse {
				m.resTab = (m.resTab - 1 + 2) % 2
				content := formatResponse(&m, m.response)
				m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(content))
			}
			return m, nil

		case "right", "l":
			if m.focus == FocusRequest {
				m.reqTab = (m.reqTab + 1) % 4
			} else if m.focus == FocusSidebar {
				m.reqTab = (m.reqTab + 1) % 4
			} else if m.focus == FocusResponse {
				m.resTab = (m.resTab + 1) % 2
				content := formatResponse(&m, m.response)
				m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(content))
			}
			return m, nil

		case "up", "k":
			if m.focus == FocusSidebar {
				m.sidebar, cmd = m.sidebar.Update(msg)
				cmds = append(cmds, cmd)
				m.onSidebarChange()
			} else if m.focus == FocusResponse || m.focus == FocusMultiRun {
				m.viewport, cmd = m.viewport.Update(msg)
				cmds = append(cmds, cmd)
			} else if m.focus == FocusRequest {
				m.reqSubFocus = 0
			}

		case "down", "j":
			if m.focus == FocusSidebar {
				m.sidebar, cmd = m.sidebar.Update(msg)
				cmds = append(cmds, cmd)
				m.onSidebarChange()
			} else if m.focus == FocusResponse || m.focus == FocusMultiRun {
				m.viewport, cmd = m.viewport.Update(msg)
				cmds = append(cmds, cmd)
			} else if m.focus == FocusRequest {
				m.reqSubFocus = 1
			}

		case " ":
			if m.focus == FocusSidebar && m.activeRequest != nil {
				idx := m.sidebar.Index()
				reqID := m.requests[idx].ID // Get the ID!
				
				if m.selectedRequests == nil {
					m.selectedRequests = make(map[string]bool)
				}
				m.selectedRequests[reqID] = !m.selectedRequests[reqID]
				
				m.rebuildSidebar()
			}
			return m, nil

		case "R", "r", "f5":
			// Start multi-run
			hasSelection := false
			for _, selected := range m.selectedRequests {
				if selected {
					hasSelection = true
					break
				}
			}
			
			// If none are selected, run ALL of them
			if !hasSelection {
				for i := range m.requests {
					m.selectedRequests[m.requests[i].ID] = true
				}
				m.rebuildSidebar()
			}
			
			if len(m.requests) > 0 {
				m.focus = FocusMultiRun
				m.updateSizes()
				m.isMultiRunning = true
				m.multiRunResults = nil
				m.multiRunErrors = nil
				m.multiRunIndex = 0
				
				// Reset viewport content
				m.viewport.SetContent(lipgloss.NewStyle().Foreground(cMauve).Render(m.spinner.View() + " Starting runner..."))
				
				return m, func() tea.Msg { return multiRunNextMsg{} }
			}

		case "enter":
			if m.focus == FocusSidebar && m.activeRequest != nil && !m.isLoading {
				m.isLoading = true
				m.response = nil
				m.err = nil
				return m, m.executeRequest(*m.activeRequest)
			}

		case "esc":
			if m.focus == FocusMultiRun && !m.isMultiRunning {
				m.focus = FocusSidebar
				m.updateSizes()
			} else if m.focus == FocusResponse {
				m.focus = FocusSidebar
			}

		default:
			// Forward unhandled keys to focused component
			if m.focus == FocusSidebar {
				m.sidebar, cmd = m.sidebar.Update(msg)
				cmds = append(cmds, cmd)
				m.onSidebarChange()
			} else if m.focus == FocusResponse || m.focus == FocusMultiRun {
				m.viewport, cmd = m.viewport.Update(msg)
				cmds = append(cmds, cmd)
			}
		}

	case tea.MouseMsg:
		if msg.Action == tea.MouseActionRelease && msg.Button == tea.MouseButtonLeft {
			sidebarWidth := m.width / 3
			if sidebarWidth < 30 {
				sidebarWidth = 30
			}
			if msg.X < sidebarWidth {
				m.focus = FocusSidebar
			} else {
				detailH := 10
				if msg.Y < detailH+1 {
					m.focus = FocusRequest
				} else {
					m.focus = FocusResponse
				}
			}
		}
		
		var cmd tea.Cmd
		if m.focus == FocusSidebar {
			m.sidebar, cmd = m.sidebar.Update(msg)
			cmds = append(cmds, cmd)
			m.onSidebarChange()
		} else if m.focus == FocusResponse || m.focus == FocusMultiRun {
			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
		}

	case authSavedMsg:
		m.showAuth = false
		m.focus = FocusSidebar
		m.envVars["WORKSPACE_AUTH_TYPE"] = msg.authType
		m.envVars["WORKSPACE_AUTH_TOKEN"] = msg.token
		env.Save(msg.collection, "WORKSPACE_AUTH_TYPE", msg.authType)
		env.Save(msg.collection, "WORKSPACE_AUTH_TOKEN", msg.token)
		return m, nil

	case multiRunNextMsg:
		// Find the next selected request to run
		for i := m.multiRunIndex; i < len(m.requests); i++ {
			if m.selectedRequests[m.requests[i].ID] {
				m.multiRunIndex = i
				req := m.requests[i]
				return m, m.executeMultiRequest(req)
			}
		}
		// If we finish loop, we are done
		return m, func() tea.Msg { return multiRunDoneMsg{} }

	case responseMsg:
		if m.isMultiRunning {
			if msg.err != nil {
				m.multiRunErrors = append(m.multiRunErrors, msg.err)
				// Create a dummy response for the error
				m.multiRunResults = append(m.multiRunResults, models.APIResponse{Status: "Error"})
			} else {
				m.multiRunErrors = append(m.multiRunErrors, nil)
				m.multiRunResults = append(m.multiRunResults, msg.response)
			}
			m.multiRunIndex++
			return m, func() tea.Msg { return multiRunNextMsg{} }
		}

		m.isLoading = false
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.response = &msg.response
			m.err = nil
			if m.activeRequest != nil {
				m.responseCache[m.activeRequest.ID] = m.response
			}
			content := formatResponse(&m, m.response)
			
			// Wrap content to fit viewport
			wrappedContent := lipgloss.NewStyle().Width(m.viewport.Width).Render(content)
			m.viewport.SetContent(wrappedContent)
			
			m.viewport.GotoTop()
			m.focus = FocusResponse
		}

	case multiRunDoneMsg:
		m.isMultiRunning = false
		m.isLoading = false
		// Render results
		var b strings.Builder
		b.WriteString(bold.Render("🏃 Multi-Run Completed\n\n"))
		var executedReqs []models.APIRequest
		for i := 0; i < len(m.requests); i++ {
			if m.selectedRequests[m.requests[i].ID] {
				executedReqs = append(executedReqs, m.requests[i])
			}
		}

		for i, res := range m.multiRunResults {
			if i >= len(executedReqs) {
				break
			}
			req := executedReqs[i]
			
			// Print Request Header
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#CBA6F7")).Bold(true).Render(fmt.Sprintf("[%d] %s %s", i+1, req.Method, req.URL)) + "\n")
			
			if m.multiRunErrors[i] != nil {
				b.WriteString(fmt.Sprintf("❌ Error: %v\n\n", m.multiRunErrors[i]))
			} else {
				statusColor := green
				statusIcon := "🟢"
				if res.StatusCode >= 400 && res.StatusCode < 500 {
					statusColor = lipgloss.NewStyle().Foreground(lipgloss.Color("#F9E2AF")) // cYellow
					statusIcon = "🟡"
				} else if res.StatusCode >= 500 {
					statusColor = red
					statusIcon = "🔴"
				}
				b.WriteString(statusColor.Render(fmt.Sprintf("%s %d %s  ⏱ %v", statusIcon, res.StatusCode, res.Status, res.Latency)) + "\n\n")
				
				if res.Body != "" {
					var prettyJSON bytes.Buffer
					if err := json.Indent(&prettyJSON, []byte(res.Body), "", "  "); err == nil {
						b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#A6ADC8")).Render(prettyJSON.String()) + "\n")
					} else {
						b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#A6ADC8")).Render(res.Body) + "\n")
					}
				} else {
					b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086")).Render("Empty body") + "\n")
				}
			}
			
			b.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#45475A")).Render(strings.Repeat("─", m.viewport.Width)) + "\n\n")
		}
		m.viewport.SetContent(b.String())
		m.viewport.GotoTop()
		return m, nil
	}

	return m, tea.Batch(cmds...)
}

// ---- New Collection form keys ----

func (m AppModel) handleNewCollectionKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "esc":
		m.focus = FocusSidebar
		return m, nil

	case "enter":
		name := strings.TrimSpace(m.newColInput.Value())
		if name == "" {
			return m, nil
		}
		
		if !strings.HasSuffix(strings.ToLower(name), ".md") {
			name += ".md"
		}

		var workspacePath string
		if len(m.collections) > 0 {
			workspacePath = filepath.Dir(m.collections[0].FilePath)
		} else {
			workspacePath, _ = env.ResolveWriteDir()
		}

		if workspacePath != "" {
			newPath := filepath.Join(workspacePath, name)
			
			// Create the file with a dummy request
			content := []byte(fmt.Sprintf("# %s\n\n```http\nGET https://api.example.com\n```\n", name))
			os.WriteFile(newPath, content, 0644)
			
			cols, _ := parser.LoadWorkspace()
			if len(cols) > 0 {
				m.collections = cols
				
				// Find the index of the newly created collection
				for i, c := range m.collections {
					if c.FilePath == newPath {
						m.activeColIndex = i
						break
					}
				}
				
				m.rebuildSidebar()
			}
		}
		
		m.focus = FocusSidebar
		return m, nil
	}

	m.newColInput, cmd = m.newColInput.Update(msg)
	return m, cmd
}

// ---- Auth form keys ----

func (m AppModel) handleAuthKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "esc":
		m.showAuth = false
		m.focus = FocusSidebar
		return m, nil

	case "tab":
		visible := []int{0}
		authType := strings.ToLower(strings.TrimSpace(m.authType.Value()))
		if authType == "bearer" || authType == "apikey" || authType == "cookie" {
			visible = append(visible, 1)
		} else if authType == "basic" {
			visible = append(visible, 2, 3)
		}
		
		currIdx := 0
		for i, v := range visible {
			if v == m.authActive {
				currIdx = i
				break
			}
		}
		nextIdx := (currIdx + 1) % len(visible)
		m.authActive = visible[nextIdx]

		m.authType.Blur()
		m.authToken.Blur()
		m.authUsername.Blur()
		m.authPassword.Blur()
		switch m.authActive {
		case 0:
			m.authType.Focus()
		case 1:
			m.authToken.Focus()
		case 2:
			m.authUsername.Focus()
		case 3:
			m.authPassword.Focus()
		}
		return m, nil

	case "enter":
		// Save auth
		colName := ""
		if m.activeColIndex < len(m.collections) {
			colName = m.collections[m.activeColIndex].Name
		}
		authType := m.authType.Value()
		token := m.authToken.Value()
		username := m.authUsername.Value()
		password := m.authPassword.Value()

		// Save all relevant fields
		if authType != "" {
			env.Save(colName, "WORKSPACE_AUTH_TYPE", authType)
			switch authType {
			case "bearer", "apikey":
				if token != "" {
					env.Save(colName, "WORKSPACE_AUTH_TOKEN", token)
				}
			case "basic":
				if username != "" {
					env.Save(colName, "WORKSPACE_AUTH_USERNAME", username)
				}
				if password != "" {
					env.Save(colName, "WORKSPACE_AUTH_PASSWORD", password)
				}
			case "cookie":
				if token != "" {
					env.Save(colName, "WORKSPACE_AUTH_TOKEN", token)
				}
			}
			m.loadEnvForActive()
		}

		return m, func() tea.Msg { return authSavedMsg{collection: colName, authType: authType, token: token} }
	}

	// Route to active input
	switch m.authActive {
	case 0:
		m.authType, cmd = m.authType.Update(msg)
	case 1:
		m.authToken, cmd = m.authToken.Update(msg)
	case 2:
		m.authPassword, cmd = m.authPassword.Update(msg)
	}
	return m, cmd
}

// ---- Collection switching ----

func (m AppModel) handleCollectionKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.focus = FocusSidebar
		return m, nil

	case "up", "k":
		if m.activeColIndex > 0 {
			m.activeColIndex--
			m.rebuildSidebar()
			m.loadEnvForActive()
		}

	case "down", "j":
		if m.activeColIndex < len(m.collections)-1 {
			m.activeColIndex++
			m.rebuildSidebar()
			m.loadEnvForActive()
		}

	case "n":
		m.focus = FocusNewCollection
		m.newColInput.SetValue("")
		m.newColInput.Focus()
		return m, nil

	case "d":
		if m.activeColIndex >= 0 && m.activeColIndex < len(m.collections) {
			os.Remove(m.collections[m.activeColIndex].FilePath)
			cols, _ := parser.LoadWorkspace()
			if len(cols) > 0 {
				m.collections = cols
				m.activeColIndex = 0
				m.rebuildSidebar()
				m.loadEnvForActive()
			} else {
				return m, tea.Quit
			}
		}
		return m, nil

	case "enter":
		m.focus = FocusSidebar
		return m, nil
	}
	
	return m, nil
}

// ---- Edit form keys ----

func (m AppModel) handleEditKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "esc":
		m.focus = FocusRequest
		return m, nil

	case "tab":
		if m.reqSubFocus == 0 {
			m.editActive = (m.editActive + 1) % 3
			m.editMethod.Blur()
			m.editTitle.Blur()
			m.editURL.Blur()
			switch m.editActive {
			case 0:
				m.editMethod.Focus()
			case 1:
				m.editTitle.Focus()
			case 2:
				m.editURL.Focus()
			}
		}
		return m, nil

	case "ctrl+s", "enter":
		// Only save on enter if it's the top bar (textinputs)
		// For textareas, enter should insert a newline, so we only save on ctrl+s
		if msg.String() == "enter" && m.reqSubFocus == 1 {
			break
		}

		if m.activeRequest != nil && m.activeColIndex < len(m.collections) {
			if m.reqSubFocus == 0 {
				m.activeRequest.Method = m.editMethod.Value()
				m.activeRequest.Title = m.editTitle.Value()
				m.activeRequest.URL = m.editURL.Value()
			} else {
				switch m.editActive {
				case 3:
					m.activeRequest.Body = m.editBody.Value()
				case 4:
					m.activeRequest.Headers = stringToMap(m.editHeaders.Value(), ":")
				case 5:
					m.activeRequest.QueryParams = stringToMap(m.editParams.Value(), "=")
				}
			}
			
			col := m.collections[m.activeColIndex]
			err := parser.UpdateRequest(col.FilePath, m.activeRequest.ID, *m.activeRequest)
			if err != nil {
				m.err = fmt.Errorf("Failed to update request: %v", err)
			} else {
				cols, _ := parser.LoadWorkspace()
				if len(cols) > 0 {
					m.collections = cols
					m.rebuildSidebar()
				}
			}
		}
		m.focus = FocusRequest
		return m, nil
	}

	// Route to active input
	switch m.editActive {
	case 0:
		m.editMethod, cmd = m.editMethod.Update(msg)
	case 1:
		m.editTitle, cmd = m.editTitle.Update(msg)
	case 2:
		m.editURL, cmd = m.editURL.Update(msg)
	case 3:
		m.editBody, cmd = m.editBody.Update(msg)
	case 4:
		m.editHeaders, cmd = m.editHeaders.Update(msg)
	case 5:
		m.editParams, cmd = m.editParams.Update(msg)
	}
	return m, cmd
}

// ---- Internal helpers ----

func (m *AppModel) onSidebarChange() {
	idx := m.sidebar.Index()
	if idx >= 0 && idx < len(m.requests) {
		req := &m.requests[idx]
		m.activeRequest = req
		
		m.response = m.responseCache[req.ID]

		// Keep textareas synchronized for passive viewing
		m.editMethod.SetValue(req.Method)
		m.editTitle.SetValue(req.Title)
		m.editURL.SetValue(req.URL)
		m.editBody.SetValue(req.Body)
		m.editHeaders.SetValue(mapToString(req.Headers, ":"))
		m.editParams.SetValue(mapToString(req.QueryParams, "="))
		
		// Make sure they are not focused
		m.editMethod.Blur()
		m.editTitle.Blur()
		m.editURL.Blur()
		m.editBody.Blur()
		m.editHeaders.Blur()
		m.editParams.Blur()
	}
}

func (m *AppModel) updateSizes() {
	sidebarWidth := m.width / 3
	if sidebarWidth < 30 {
		sidebarWidth = 30
	}
	bodyH := m.height - 3
	if bodyH < 0 { bodyH = 0 }
	
	m.sidebar.SetSize(sidebarWidth-4, bodyH-4)
	
	rightW := m.width - sidebarWidth
	detailH := bodyH / 2
	if detailH < 14 { detailH = 14 }
	
	responseH := bodyH - detailH
	if responseH < 0 { responseH = 0 }
	
	m.viewport.Width = rightW - 6
	if m.focus == FocusMultiRun {
		m.viewport.Height = bodyH - 4
	} else {
		m.viewport.Height = responseH - 4
	}
	
	// Update textarea sizes so they form rigid, scrollable text boxes
	taWidth := rightW - 8
	if taWidth < 0 { taWidth = 0 }
	
	// detailH is the full panel height.
	// borders = 2, Request Details title = 2. Inside renderDetail = detailH - 4.
	// Inside renderDetail: TopBar(2) + Tabs(2) + Divider(2) + TabContent
	// So TabContent height = (detailH - 4) - 6 = detailH - 10
	taHeight := detailH - 10
	if taHeight < 1 { taHeight = 1 }
	
	m.editBody.SetWidth(taWidth)
	m.editBody.SetHeight(taHeight)
	
	m.editHeaders.SetWidth(taWidth)
	m.editHeaders.SetHeight(taHeight)
	
	m.editParams.SetWidth(taWidth)
	m.editParams.SetHeight(taHeight)
}

func (m AppModel) executeRequest(req models.APIRequest) tea.Cmd {
	return func() tea.Msg {
		resolveRequest(&req, m.envVars)
		ctx := context.Background()
		resp, err := httpclient.Execute(ctx, req)
		if err != nil {
			return responseMsg{err: err}
		}
		return responseMsg{response: resp}
	}
}

func (m AppModel) executeMultiRequest(req models.APIRequest) tea.Cmd {
	return func() tea.Msg {
		resolveRequest(&req, m.envVars)
		ctx := context.Background()
		resp, err := httpclient.Execute(ctx, req)
		if err != nil {
			return responseMsg{err: err}
		}
		return responseMsg{response: resp}
	}
}


var (
	accent = lipgloss.NewStyle().Foreground(lipgloss.Color("#CBA6F7"))
	green  = lipgloss.NewStyle().Foreground(lipgloss.Color("#A6E3A1"))
	red    = lipgloss.NewStyle().Foreground(lipgloss.Color("#F38BA8"))
	blue   = lipgloss.NewStyle().Foreground(lipgloss.Color("#89B4FA"))
	purple = lipgloss.NewStyle().Foreground(lipgloss.Color("#CBA6F7"))
	gray   = lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086"))
	bold   = lipgloss.NewStyle().Bold(true)
)

func dim(s string) string {
	return gray.Render(s)
}

func mapToString(m map[string]string, sep string) string {
	var b strings.Builder
	for k, v := range m {
		b.WriteString(fmt.Sprintf("%s%s %s\n", k, sep, v))
	}
	return strings.TrimSuffix(b.String(), "\n")
}

func stringToMap(s string, sep string) map[string]string {
	m := make(map[string]string)
	lines := strings.Split(s, "\n")
	for _, line := range lines {
		if idx := strings.Index(line, sep); idx != -1 {
			k := strings.TrimSpace(line[:idx])
			v := strings.TrimSpace(line[idx+len(sep):])
			if k != "" {
				m[k] = v
			}
		}
	}
	return m
}
