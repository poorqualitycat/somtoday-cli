# SOMtoday REST API docs

## Somtoday-TUI (Go)

This repository now also contains a terminal UI app in Go: **Somtoday-TUI**.

### Run

```bash
go run .
```

Optional environment variables:

- `SOMTODAY_API_URL` (default: `https://api.somtoday.nl`)
- `SOMTODAY_ACCESS_TOKEN`
- `SOMTODAY_STUDENT_ID`

When you set these values in the TUI menu, the app also stores them in your OS secure credential store (for example Linux Secret Service/KWallet, macOS Keychain, or Windows Credential Manager). On startup, stored values are loaded automatically if env vars are not set.

### What it supports

The menu includes actions for:

- School discovery (`/organisaties.json`)
- Organisation UUID lookup (`/rest/v1/organisaties`)
- OAuth password login (`/oauth2/token`) with school UUID + username + password
- Access token refresh (`/oauth2/token` with `refresh_token`)
- Student, grades, schedule, absence, study guides, subjects, account, school years, vakkeuzes, waarnemingen, messages
- Schoolgegevens, vakanties, studiemateriaal, and iCalendar endpoints
- Homework reads (appointments/days/weeks) and homework updates (`swigemaakt`)
- A custom request mode for any endpoint

Responses are printed as formatted JSON where possible.

### OAuth note

Somtoday-TUI does not require a fixed API key. It uses OAuth tokens. After login/refresh, access and refresh tokens are stored in your OS keyring. Passwords are used only for the login request and are not stored.
