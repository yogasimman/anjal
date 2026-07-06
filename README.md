<div align="center">
  <h1>🕊️ Anjal</h1>
  <p><strong>A modern, keyboard-driven API testing tool for the terminal.</strong></p>
</div>

Anjal is a powerful cross-platform API testing tool designed for developers who love automation and speed. It acts as a lightweight Postman alternative, allowing you to organize, edit, and execute HTTP requests from a clean Graphical User Interface (GUI) or purely from the terminal via the Terminal User Interface (TUI). 

Our goal with Anjal is to bridge the gap between traditional UI-heavy API clients and automation-friendly scripts, creating an environment that is heavily geared towards developer experience, automation, and speed.

---

## Installation

Anjal provides two distinct applications depending on your workflow: **Anjal Desktop** (GUI) and **Anjal CLI** (TUI/Headless).

### 1. Anjal Desktop (GUI)
The easiest way to install Anjal is by downloading the pre-compiled installer for your operating system from the **[GitHub Releases page](../../releases)**. 
We provide fully automated builds for all major platforms:
- **macOS**: `Anjal-macOS.dmg` (Universal binary for Apple Silicon & Intel)
  - *Note for Mac Users: Because Anjal is open-source and not signed with a paid Apple Developer certificate, macOS Gatekeeper may say the app is "damaged" or "malware". To fix this, after copying Anjal to your Applications folder, open your terminal and run:*
  - `xattr -cr /Applications/anjal-desktop.app`
- **Windows**: `Anjal-Windows.exe` (NSIS Installer)
- **Linux**: `.AppImage`, `.deb`, and `.rpm` packages

### 2. Anjal CLI (TUI)
If you prefer to stay in the terminal, you can install the CLI version instantly using Go:

```bash
go install github.com/yogasimman/anjal/cmd/anjal@latest
```
*Note: Ensure your `~/go/bin` is in your system's `$PATH` so you can simply type `anjal` into your terminal!*

---

## Developing the GUI Locally

Anjal Desktop is built using [Wails](https://wails.io/), combining a Go backend with a React (TypeScript + Vite) frontend.

To run the GUI in live development mode (with hot-reloading for the frontend):
1. Install the Wails CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
2. Navigate to the `gui` directory: `cd gui`
3. Run the development server: `wails dev`

---

## The UI

Anjal's UI is designed with a sleek, rigid three-pane layout that maximizes terminal real estate and is driven entirely by keyboard shortcuts:
- **Left Panel (Sidebar)**: Your API Collection navigator. Lists all endpoints in your active `.md` file.
- **Top Right (Request Detail)**: The active request viewer and editor, showcasing HTTP methods, URLs, Headers, and JSON Body.
- **Bottom Right (Response)**: Streams the response JSON (prettified) with Status Codes and Latency information.

> *![Anjal Main UI](docs/assets/ui-main.png)*
> *(Tip: Replace this with a screenshot of the main Anjal UI)*

## Workspace & Collections (`.anjal/`)

Anjal intelligently detects your workspace. By default, it searches for a `.anjal/` directory in your current path. 
Inside this directory, Anjal reads Markdown (`.md`) files, treating each file as a distinct **Collection** of API requests.

When you run `anjal`, it will parse these `.md` files and populate your left sidebar.

## The `.md` File Format

Anjal stores all API collections as Markdown files, making them portable, version-controllable, and easily editable without the UI. 
A request is defined using standard Markdown headers and codeblocks. 

Here is how you can create an API block manually:

```markdown
# Get Workforce Data

```http
POST http://localhost:3001/api/v5/leaveManagement/getWorkforceData
@id req-12345
Content-Type: application/json

{
  "page": 1,
  "pageSize": 10,
  "status": "PENDING"
}
```
```

> **Note**: A separate `prompt.md` document will discuss this schema in greater detail, specifically on how AI assistants can automatically generate these files for testing purposes.

## Navigation and Editing

Anjal is completely keyboard friendly. You can fluidly move between different sections of the app using intuitive keybinds:
- **`Tab` / `Shift+Tab`**: Cycle focus through the Sidebar, Request Pane, and Response Pane.
- **`Up/Down` or `j/k`**: Scroll through lists and JSON responses.
- **`e`**: Enter **Edit Mode** on the focused request. This opens an inline form where you can tweak the URL, Body, Headers, etc. 

> *![Edit Mode](docs/assets/ui-edit.png)*

## Global Authentication

Instead of attaching tokens to every individual request, Anjal supports Collection-level Auth.
- Press **`a`** to open the Authentication Modal.
- Specify your Auth Type (e.g., `bearer`) and your token.
- This token is saved to the environment and automatically injected into every request executed in that collection.

> *![Auth Modal](docs/assets/ui-auth.png)*

## Multi-Run Execution

Need to run an entire suite of tests at once? Anjal's Multi-Run feature behaves like a full test runner.
- Press **`r`** (Run All).
- The Request and Response panels collapse into a single unified dashboard.
- Anjal sequentially fires all APIs in the collection and streams detailed results—including Method, URL, precise Latency, Status Colors, and the prettified Response Body—into the view.
- *You can also press `Spacebar` on specific requests to select a subset to run!*

> *![Multi Run Results](docs/assets/ui-multirun.png)*

## Headless Mode (`--noui`)

For CI/CD pipelines or users who just want quick terminal output, Anjal supports a Headless execution mode.

**Interactive CLI:**
```bash
go run ./cmd/anjal --noui
```
This discovers your collections, prints them to stdout, and prompts you to select one by number.

**Direct Execution:**
```bash
go run ./cmd/anjal --noui cmd/anjal/.anjal/leave_management.md
```
This immediately parses the specified file, executes all endpoints sequentially, prints the detailed formatted output directly to the terminal, and exits.
