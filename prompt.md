# Anjal Collection Generator Prompt

**Target Audience:** AI Assistants (LLMs) and Human Developers.
**Purpose:** This document serves as a standard instruction set for generating API collections compatible with the **Anjal** terminal API client.

## What is Anjal?
Anjal is a keyboard-driven Terminal User Interface (TUI) for executing API requests. Unlike traditional API clients that store data in complex JSON structures, Anjal stores API collections entirely in **Markdown (`.md`) files**. 

When tasked with "creating an Anjal collection", you must generate a standard Markdown file adhering strictly to the syntax outlined below.

---

## The Anjal Markdown Schema

Anjal parses `.md` files by looking for a specific combination of Markdown Headers (`#`) and HTTP codeblocks (````http`). 

### Basic Structure

Each API request in the collection MUST follow this exact format:

1. **Title**: An `H1` markdown header (`# Title`) describing the request.
2. **Codeblock**: A markdown codeblock with the language tagged as `http`.
3. **HTTP Protocol**: Inside the block, the first line must be the `METHOD` followed by the `URL`.
4. **Directives (Optional)**: Lines starting with `@` control special Anjal features:
   - `@id req-users-1`: Unique identifier for the request.
   - `@query key value`: Query parameter injection.
   - `@auth bearer token`: Inline authentication (though global auth is preferred).
5. **Headers (Optional)**: Standard HTTP `Key: Value` pairs.
6. **Body (Optional)**: An empty line separating the headers from the payload, followed by the raw body (e.g., JSON).

### Example Collection

```markdown
# Get All Users

```http
GET https://api.example.com/v1/users
@id req-get-users
@query limit 10
@query sort desc
Authorization: Bearer my-token
```

# Create New Employee

```http
POST https://api.example.com/v1/employees
@id req-create-employee
Content-Type: application/json
Accept: application/json

{
  "first_name": "John",
  "last_name": "Doe",
  "department": "Engineering"
}
```
```

---

## Critical Rules for AI Generators

When an AI is prompted to generate or update an Anjal collection, it MUST adhere to the following strict rules:

1. **One Request Per H1 Header**: Anjal splits requests based on `# ` headers. Do not put multiple ````http` blocks under a single H1 header. If you have a new request, create a new `# Header`.
2. **Exact Syntax**: 
   - The method (GET, POST, PUT, DELETE, PATCH) must be capitalized.
   - The URL must be on the same line as the method, separated by a space.
3. **Empty Lines Matter**: There must be exactly one blank line between the HTTP Headers and the HTTP Body inside the codeblock. Do not add blank lines between headers.
4. **No UI Configuration Needed for Global Auth**: Do not include authentication tokens explicitly in headers if they are globally managed (unless specifically asked). Anjal supports workspace-level variables natively. 
   - Anjal stores environment variables and Auth tokens in hidden `.env` files located inside the `.anjal/` workspace directory.
   - For example, if a collection is named `users.md`, Anjal will automatically create and read from a `.users.env` file in the workspace to manage tokens securely without polluting the `.md` file!
5. **Human Readability**: You may add standard markdown text outside of the `# Header` and ````http` blocks. Anjal will simply ignore text that isn't part of an `H1` or `http` block, meaning you can freely document the APIs for human readers using standard markdown paragraphs.

## Instructions for Humans

If you are manually writing an Anjal file:
- Simply create a `.md` file inside your `.anjal/` workspace directory.
- Copy and paste the example above.
- Anjal will dynamically live-reload or parse this file the next time you open the application. 
- You do not need to use the Anjal UI to create endpoints; you can edit these files directly in your favorite text editor (Vim, VSCode, etc.).
