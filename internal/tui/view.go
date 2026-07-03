package tui

import (
	"bytes"
	"encoding/json"
	_ "embed"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yogasimman/anjal/internal/models"
)

//go:embed ascii_logo.txt
var splashLogo string

//go:embed ascii_logo_text.txt
var splashText string

// Catppuccin Mocha colors
var (
	cText    = lipgloss.Color("#CDD6F4")
	cSubtext = lipgloss.Color("#A6ADC8")
	cOverlay = lipgloss.Color("#6C7086")
	cSurface = lipgloss.Color("#313244")
	
	cBlue    = lipgloss.Color("#89B4FA")
	cMauve   = lipgloss.Color("#CBA6F7")
	cGreen   = lipgloss.Color("#A6E3A1")
	cRed     = lipgloss.Color("#F38BA8")
	cPeach   = lipgloss.Color("#FAB387")
	cYellow  = lipgloss.Color("#F9E2AF")
	cTeal    = lipgloss.Color("#94E2D5")

	// Core styles
	logoStyle = lipgloss.NewStyle().
			Foreground(cMauve).
			Bold(true)

	versionStyle = lipgloss.NewStyle().
			Foreground(cOverlay)
			
	statusBarStyle = lipgloss.NewStyle().
			Background(cSurface).
			Foreground(cText).
			Padding(0, 1)

	// Method Badges
	badgeGET = lipgloss.NewStyle().
		Background(cTeal).
		Foreground(lipgloss.Color("#11111B")).
		Bold(true).Padding(0, 1)

	badgePOST = lipgloss.NewStyle().
		Background(cBlue).
		Foreground(lipgloss.Color("#11111B")).
		Bold(true).Padding(0, 1)

	badgePUT = lipgloss.NewStyle().
		Background(cPeach).
		Foreground(lipgloss.Color("#11111B")).
		Bold(true).Padding(0, 1)

	badgePATCH = lipgloss.NewStyle().
		Background(cYellow).
		Foreground(lipgloss.Color("#11111B")).
		Bold(true).Padding(0, 1)

	badgeDELETE = lipgloss.NewStyle().
		Background(cRed).
		Foreground(lipgloss.Color("#11111B")).
		Bold(true).Padding(0, 1)

	// Auth modal & dialogs
	modalStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(cBlue).
			Padding(1, 2)
)

// Helper for dynamic borders based on focus
func panelStyle(focused bool) lipgloss.Style {
	borderColor := cOverlay
	if focused {
		borderColor = cMauve
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1)
}

func (m AppModel) View() string {
	if m.width == 0 {
		return "Initializing..."
	}
	if m.focus == FocusSplash {
		return m.renderSplash()
	}

	// ---- Header ----
	header := m.renderHeader()

	// ---- Help bar / Footer ----
	help := m.renderHelp()

	bodyH := m.height - lipgloss.Height(header) - lipgloss.Height(help)
	if bodyH < 0 {
		bodyH = 0
	}

	// ---- Body ----
	var body string
	if m.focus == FocusNewCollection {
		body = m.renderNewCollectionModal(bodyH)
	} else if m.showAuth {
		body = m.renderAuthModal(bodyH)
	} else if m.focus == FocusCollections {
		body = m.renderCollectionBrowser(bodyH)
	} else {
		body = m.renderMain(bodyH)
	}

	return lipgloss.JoinVertical(lipgloss.Top, header, body, help)
}

func (m AppModel) renderHeader() string {
	menus := []string{
		"[e] Edit",
		"[a] Auth",
		"[c] Col",
		"[n] New Req",
		"[C] New Col",
		"[d] Del",
		"[r] Run All",
	}

	var menuStr strings.Builder
	for _, menu := range menus {
		menuStr.WriteString(lipgloss.NewStyle().Foreground(cSubtext).Render(menu) + " ")
	}

	left := lipgloss.JoinHorizontal(lipgloss.Center, logoStyle.Render("✦ Anjal "), versionStyle.Render("v1.0  "), menuStr.String())

	spinnerDisp := ""
	if m.isLoading {
		spinnerDisp = "  " + lipgloss.NewStyle().Foreground(cBlue).Render(m.spinner.View() + " Loading...")
	}

	rightWidth := m.width - lipgloss.Width(left) - 4
	if rightWidth < 0 {
		rightWidth = 0
	}

	return lipgloss.NewStyle().Width(m.width).Padding(0, 1).Render(
		lipgloss.JoinHorizontal(lipgloss.Top,
			left,
			strings.Repeat(" ", rightWidth),
			spinnerDisp,
		),
	)
}

func formatResponse(m *AppModel, res *models.APIResponse) string {
	var b strings.Builder

	statusIcon := "🟢"
	if res.StatusCode >= 400 && res.StatusCode < 500 {
		statusIcon = "🟡"
	} else if res.StatusCode >= 500 {
		statusIcon = "🔴"
	}

	contentType := ""
	if ct, ok := res.Headers["Content-Type"]; ok && len(ct) > 0 {
		contentType = ct[0]
	}

	headerBar := lipgloss.JoinHorizontal(lipgloss.Top,
		statusIcon+" "+lipgloss.NewStyle().Bold(true).Render(res.Status),
		"  •  ⏱  ",
		res.Latency.String(),
		"  •  📄 ",
		contentType,
	)
	b.WriteString(headerBar + "\n\n")

	tabs := []string{"Body", "Headers"}
	var tabStr strings.Builder
	for i, t := range tabs {
		if i == int(m.resTab) {
			tabStr.WriteString(lipgloss.NewStyle().Foreground(cMauve).Bold(true).Render(" [ " + t + " ] "))
		} else {
			tabStr.WriteString(lipgloss.NewStyle().Foreground(cOverlay).Render("   " + t + "   "))
		}
	}
	b.WriteString(tabStr.String() + "\n")
	
	lineWidth := m.viewport.Width
	if lineWidth > 100 {
		lineWidth = 100
	}
	b.WriteString(lipgloss.NewStyle().Foreground(cSurface).Render(strings.Repeat("─", lineWidth)) + "\n\n")

	switch m.resTab {
	case ResTabBody:
		if res.Body != "" {
			var prettyJSON bytes.Buffer
			if err := json.Indent(&prettyJSON, []byte(res.Body), "", "  "); err == nil {
				b.WriteString(prettyJSON.String())
			} else {
				b.WriteString(res.Body)
			}
		} else {
			b.WriteString(lipgloss.NewStyle().Foreground(cOverlay).Render("Empty response body"))
		}
	case ResTabHeaders:
		for k, v := range res.Headers {
			b.WriteString(fmt.Sprintf("%s: %s\n", lipgloss.NewStyle().Foreground(cBlue).Render(k), strings.Join(v, ", ")))
		}
	}

	return b.String()
}

func (m AppModel) renderMain(bodyH int) string {
	sidebarW := m.width / 3
	if sidebarW < 30 {
		sidebarW = 30
	}
	rightW := m.width - sidebarW

	detailH := bodyH / 2
	if detailH < 14 {
		detailH = 14
	}
	responseH := bodyH - detailH
	if responseH < 0 {
		responseH = 0
	}

	// Sidebar Panel
	sidebarFocus := m.focus == FocusSidebar
	sidebarStyle := panelStyle(sidebarFocus).Width(sidebarW - 2).Height(bodyH - 2)
	// We'll just rely on the list's own title for now
	sidebar := sidebarStyle.Render(m.sidebar.View())

	// Right Pane content
	var right string
	
	if m.focus == FocusMultiRun {
		multiStyleBox := panelStyle(true).Width(rightW - 2).Height(bodyH - 2)
		right = multiStyleBox.Render(lipgloss.NewStyle().Foreground(cMauve).Bold(true).Render(" Multi-Run Results") + "\n\n" + m.viewport.View())
	} else {
		// Request Detail Panel
		reqFocus := m.focus == FocusRequest || m.focus == FocusEdit
		detailStyleBox := panelStyle(reqFocus).Width(rightW - 2).Height(detailH - 2)
		
		var detail string
		if m.focus == FocusEdit {
			detail = detailStyleBox.Render(lipgloss.NewStyle().Foreground(cMauve).Bold(true).Render(" Edit Request") + "\n\n" + m.renderInlineEdit(rightW-6, detailH-4))
		} else {
			detail = detailStyleBox.Render(lipgloss.NewStyle().Foreground(cMauve).Bold(true).Render(" Request Details") + "\n\n" + m.renderDetail(rightW-6, detailH-4))
		}
	
		// Response Panel
		resFocus := m.focus == FocusResponse
		resStyleBox := panelStyle(resFocus).Width(rightW - 2).Height(responseH - 2)
		
		var response string
		if m.isLoading {
			response = m.renderWaiting(rightW-6, responseH-4)
		} else if m.err != nil {
			response = m.renderError(rightW-6, responseH-4)
		} else if m.response != nil {
			response = m.viewport.View()
		} else {
			response = m.renderEmpty(rightW-6, responseH-4)
		}
		
		responsePane := resStyleBox.Render(lipgloss.NewStyle().Foreground(cMauve).Bold(true).Render(" Response") + "\n\n" + response)
	
		right = lipgloss.JoinVertical(lipgloss.Left, detail, responsePane)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, sidebar, right)
}

func (m AppModel) renderDetail(w, h int) string {
	req := m.activeRequest
	if req == nil {
		return lipgloss.NewStyle().Width(w).Height(h).Render(lipgloss.NewStyle().Foreground(cOverlay).Render("No request selected"))
	}

	var b strings.Builder
	
	methodBadge := badgeGET
	switch req.Method {
	case "POST": methodBadge = badgePOST
	case "PUT": methodBadge = badgePUT
	case "PATCH": methodBadge = badgePATCH
	case "DELETE": methodBadge = badgeDELETE
	}

	topBarColor := cText
	if m.focus == FocusRequest && m.reqSubFocus == 0 {
		topBarColor = cMauve // Highlight
	}

	topStr := methodBadge.Render(req.Method) + " " + lipgloss.NewStyle().Bold(true).Foreground(topBarColor).Render(req.URL) + "\n"
	if req.Title != "" {
		topStr += lipgloss.NewStyle().Foreground(cSubtext).Render(req.Title)
	} else {
		topStr += " "
	}
	
	if m.focus == FocusRequest && m.reqSubFocus == 0 {
		b.WriteString(lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, false, true).BorderForeground(cMauve).PaddingLeft(1).Render(topStr) + "\n\n")
	} else {
		b.WriteString(topStr + "\n\n")
	}

	tabs := []string{"Body", "Headers", "Auth", "Params"}
	var tabStr strings.Builder
	for i, t := range tabs {
		if i == int(m.reqTab) {
			tabStr.WriteString(lipgloss.NewStyle().Foreground(cMauve).Bold(true).Render(" [ " + t + " ] "))
		} else {
			tabStr.WriteString(lipgloss.NewStyle().Foreground(cOverlay).Render("   " + t + "   "))
		}
	}
	b.WriteString(tabStr.String() + "\n")
	
	divW := w
	if divW < 0 { divW = 0 }
	b.WriteString(lipgloss.NewStyle().Foreground(cSurface).Render(strings.Repeat("─", divW)) + "\n\n")

	contentH := h - 6
	if contentH < 0 { contentH = 0 }
	
	var content string
	switch m.reqTab {
	case ReqTabBody:
		content = m.editBody.View()
	case ReqTabHeaders:
		content = m.editHeaders.View()
	case ReqTabAuth:
		if req.Auth == nil {
			content = lipgloss.NewStyle().Foreground(cOverlay).Render("Inheriting Workspace Auth.")
		} else {
			var hb strings.Builder
			hb.WriteString(lipgloss.NewStyle().Foreground(cBlue).Render("Type: ") + req.Auth.Type + "\n")
			for k, v := range req.Auth.Params {
				hb.WriteString(fmt.Sprintf("%s: %s\n", k, v))
			}
			content = hb.String()
		}
	case ReqTabParams:
		content = m.editParams.View()
	}
	
	if m.focus == FocusRequest && m.reqSubFocus == 1 {
		content = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, false, true).BorderForeground(cMauve).PaddingLeft(1).Render(content)
	}

	b.WriteString(content)
	return lipgloss.NewStyle().Width(w).Height(h).Render(b.String())
}

func (m AppModel) renderWaiting(w, h int) string {
	sp := lipgloss.NewStyle().Foreground(cMauve).Render(m.spinner.View() + " Sending...")
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, sp)
}

func (m AppModel) renderError(w, h int) string {
	errStr := lipgloss.NewStyle().Foreground(cRed).Render(fmt.Sprintf("❌ %v", m.err))
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, errStr)
}

func (m AppModel) renderEmpty(w, h int) string {
	msg := lipgloss.NewStyle().Foreground(cOverlay).Render("Press Enter on a request to execute\nPress a for auth, c for collections")
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, msg)
}

func (m AppModel) renderMultiRun(bodyH int) string {
	return panelStyle(true).Width(m.width - 2).Height(bodyH - 2).Render(lipgloss.NewStyle().Foreground(cMauve).Bold(true).Render(" Multi-Run Results") + "\n\n" + m.viewport.View())
}

// ---- Edit Mode ----

func (m AppModel) renderInlineEdit(w, h int) string {
	if m.activeRequest == nil {
		return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, lipgloss.NewStyle().Foreground(cOverlay).Render("No request selected to edit"))
	}

	var b strings.Builder
	
	if m.reqSubFocus == 0 {
		// Method
		if m.editActive == 0 {
			b.WriteString(lipgloss.NewStyle().Foreground(cMauve).Render("▸ Method: ") + m.editMethod.View() + "\n")
		} else {
			b.WriteString(lipgloss.NewStyle().Foreground(cSubtext).Render("  Method: ") + m.editMethod.View() + "\n")
		}
	
		// Title
		if m.editActive == 1 {
			b.WriteString(lipgloss.NewStyle().Foreground(cMauve).Render("▸ Title:  ") + m.editTitle.View() + "\n")
		} else {
			b.WriteString(lipgloss.NewStyle().Foreground(cSubtext).Render("  Title:  ") + m.editTitle.View() + "\n")
		}
	
		// URL
		if m.editActive == 2 {
			b.WriteString(lipgloss.NewStyle().Foreground(cMauve).Render("▸ URL:    ") + m.editURL.View() + "\n")
		} else {
			b.WriteString(lipgloss.NewStyle().Foreground(cSubtext).Render("  URL:    ") + m.editURL.View() + "\n")
		}
		
		b.WriteString("\n" + lipgloss.NewStyle().Foreground(cOverlay).Render("tab: next field  enter: save  esc: cancel"))
	} else {
		// Tab Content Edit
		switch m.reqTab {
		case ReqTabBody:
			b.WriteString(lipgloss.NewStyle().Foreground(cMauve).Render("▸ Edit Body:") + "\n" + m.editBody.View() + "\n")
		case ReqTabHeaders:
			b.WriteString(lipgloss.NewStyle().Foreground(cMauve).Render("▸ Edit Headers (Key: Value):") + "\n" + m.editHeaders.View() + "\n")
		case ReqTabParams:
			b.WriteString(lipgloss.NewStyle().Foreground(cMauve).Render("▸ Edit Params (Key=Value):") + "\n" + m.editParams.View() + "\n")
		}
		
		b.WriteString("\n" + lipgloss.NewStyle().Foreground(cOverlay).Render("ctrl+s: save  esc: cancel"))
	}

	return lipgloss.NewStyle().Width(w).Height(h).Render(b.String())
}

// ---- Splash Screen ----

func (m AppModel) renderSplash() string {
	logoBlock := lipgloss.NewStyle().Foreground(cBlue).Render(splashLogo)
	titleBlock := lipgloss.NewStyle().Foreground(cText).Bold(true).Render(splashText)
	
	content := lipgloss.JoinVertical(lipgloss.Center, logoBlock, titleBlock, "\n", lipgloss.NewStyle().Foreground(cMauve).Render(m.spinner.View()+" Loading..."))
	
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

// ---- Auth Modal ----

func (m AppModel) renderAuthModal(bodyH int) string {
	var b strings.Builder

	at := m.authType.View()
	tok := m.authToken.View()
	un := m.authUsername.View()
	pw := m.authPassword.View()

	authType := strings.ToLower(strings.TrimSpace(m.authType.Value()))
	showToken := authType == "bearer" || authType == "apikey" || authType == "cookie"
	showBasic := authType == "basic"

	if m.authActive == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(cMauve).Render("▸ Type:   ") + at + "\n")
	} else {
		b.WriteString("  Type:   " + at + "\n")
	}

	if showToken {
		if m.authActive == 1 {
			b.WriteString(lipgloss.NewStyle().Foreground(cMauve).Render("▸ Token:  ") + tok + "\n")
		} else {
			b.WriteString("  Token:  " + tok + "\n")
		}
	}

	if showBasic {
		if m.authActive == 2 {
			b.WriteString(lipgloss.NewStyle().Foreground(cMauve).Render("▸ User:   ") + un + "\n")
		} else {
			b.WriteString("  User:   " + un + "\n")
		}
		if m.authActive == 3 {
			b.WriteString(lipgloss.NewStyle().Foreground(cMauve).Render("▸ Pass:   ") + pw + "\n")
		} else {
			b.WriteString("  Pass:   " + pw + "\n")
		}
	}

	b.WriteString("\n" + lipgloss.NewStyle().Foreground(cOverlay).Render("tab: next field  enter: save  esc: cancel"))

	content := lipgloss.NewStyle().Padding(1, 2).Render(b.String())
	return lipgloss.Place(m.width, bodyH, lipgloss.Center, lipgloss.Center, panelStyle(true).Render(content))
}

// ---- New Collection Screen ----

func (m AppModel) renderNewCollectionModal(bodyH int) string {
	var b strings.Builder
	
	b.WriteString(lipgloss.NewStyle().Foreground(cMauve).Bold(true).Render(" Create New Collection ") + "\n\n")
	b.WriteString("Enter a name for the new collection (e.g. users, authentication):\n\n")
	
	b.WriteString(lipgloss.NewStyle().Foreground(cMauve).Render("▸ Name: ") + m.newColInput.View() + "\n\n")
	
	b.WriteString(lipgloss.NewStyle().Foreground(cOverlay).Render("enter: create  esc: cancel"))

	content := lipgloss.NewStyle().Padding(1, 2).Render(b.String())
	return lipgloss.Place(m.width, bodyH, lipgloss.Center, lipgloss.Center, panelStyle(true).Render(content))
}

// ---- Collection Browser ----

func (m AppModel) renderCollectionBrowser(bodyH int) string {
	var b strings.Builder

	for i, col := range m.collections {
		marker := "  "
		if i == m.activeColIndex {
			marker = lipgloss.NewStyle().Foreground(cMauve).Render("▶ ")
		}
		b.WriteString(fmt.Sprintf("%s%s  (%d requests)\n",
			marker, lipgloss.NewStyle().Foreground(cText).Render(col.Name), len(col.Requests)))
	}

	b.WriteString("\n" + lipgloss.NewStyle().Foreground(cOverlay).Render("↑↓ navigate  ↵ select  esc back"))

	boxW := 50
	boxH := len(m.collections)*2 + 6
	
	renderedBox := panelStyle(true).
		Width(boxW).
		Height(boxH).
		Render(b.String())

	return lipgloss.Place(m.width, bodyH, lipgloss.Center, lipgloss.Center, renderedBox)
}

// ---- Help bar ----

func (m AppModel) renderHelp() string {
	colCount := len(m.collections)
	reqCount := len(m.requests)
	stats := fmt.Sprintf(" 📁 %d collections   📝 %d requests", colCount, reqCount)

	keys := []string{"TAB Cycle", "[ ] Tabs", "SPC Select", "R Run Selected", "RET Execute", "A Auth", "C Collections", "Y Copy", "Q Quit"}
	if m.showAuth {
		keys = []string{"TAB Next Field", "RET Save", "ESC Cancel"}
	} else if m.focus == FocusCollections {
		keys = []string{"↑↓ Navigate", "RET Select", "ESC Back"}
	}

	topLine := lipgloss.PlaceHorizontal(m.width, lipgloss.Left, lipgloss.NewStyle().Foreground(cTeal).Render(stats))
	
	// Format keys
	var keyStr strings.Builder
	for _, k := range keys {
		parts := strings.SplitN(k, " ", 2)
		if len(parts) == 2 {
			keyStr.WriteString(lipgloss.NewStyle().Foreground(cBlue).Bold(true).Render(parts[0]) + " " + lipgloss.NewStyle().Foreground(cText).Render(parts[1]) + "  ")
		} else {
			keyStr.WriteString(lipgloss.NewStyle().Foreground(cText).Render(k) + "  ")
		}
	}
	
	bottomLine := lipgloss.PlaceHorizontal(m.width, lipgloss.Left, " "+keyStr.String())
	
	return statusBarStyle.Render(topLine + "\n" + bottomLine)
}
