# Dandy Dashboard

A personal, modular dashboard built with a Go backend and Svelte frontend. Drop in the widgets you want — adding a new one is three files.

![Dashboard screenshot placeholder](docs/screenshot.png)

## Widgets

| Widget | Description |
|---|---|
| **Word of the Day** | Daily Japanese vocabulary with reading, JLPT level, meanings, and example sentences (via [Jotoba](https://jotoba.de)) |
| **Upcoming Events** | Next 7 days from Google Calendar |
| **Claude AI** | Streaming chat with Claude, your personal assistant |

## Tech Stack

- **Backend** — Go 1.25 + [Echo](https://echo.labstack.com/)
- **Frontend** — [Svelte 5](https://svelte.dev/) + TypeScript + Vite + Tailwind CSS
- **AI** — [Anthropic API](https://docs.anthropic.com/) (Claude 3.5 Sonnet, streaming SSE)
- **Calendar** — Google Calendar API (service account or OAuth2)

## Getting Started

### 1. Clone & configure

```bash
git clone https://github.com/dandydeveloper/dandy-dashboard
cd dandy-dashboard
cp .env.example .env
# Edit .env and add your ANTHROPIC_API_KEY at minimum
```

### 2. Run in development

**Terminal 1 — backend** (with hot reload via [air](https://github.com/air-verse/air)):
```bash
make dev-backend
```

**Terminal 2 — frontend** (Vite HMR):
```bash
make dev-frontend
```

Open [http://localhost:5173](http://localhost:5173).

### 3. Production build

```bash
make build
./bin/server
```

Or with Docker:
```bash
docker compose up --build
```

## Configuration

| Variable | Required | Description |
|---|---|---|
| `ANTHROPIC_API_KEY` | Yes | From [console.anthropic.com](https://console.anthropic.com/keys) |
| `GOOGLE_CREDENTIALS_JSON` | No | Path to service account JSON, or raw JSON string |
| `GOOGLE_CALENDAR_ID` | No | Calendar ID (default: `primary`) |
| `PORT` | No | Server port (default: `8080`) |
| `ALLOWED_ORIGINS` | No | CORS origins (default: `http://localhost:5173`) |
| `DASHBOARD_KEY` | No | Shared secret for `X-Dashboard-Key` header auth |

## Adding a Widget

The plugin system is explicit — no magic. Three steps each side.

**Backend** — create `internal/widgets/mywidget/`:
```go
// widget.go — implement the Widget interface
func (w *Widget) Slug() string { return "mywidget" }
func (w *Widget) RegisterRoutes(g *echo.Group) {
    g.GET("/data", w.handler.Data)
}
```
Then add one line to `cmd/server/main.go`:
```go
registry.Register(mywidget.New(cfg))
```

**Frontend** — create `web/src/widgets/mywidget/`:
```typescript
// index.ts
export const myWidget: WidgetDescriptor = {
  id: 'mywidget',
  title: 'My Widget',
  description: '...',
  component: MyWidget,   // Svelte component
  defaultSize: 'md',
}
```
Then add one line to `web/src/widgets/registry.ts`:
```typescript
import { myWidget } from './mywidget'
export const widgetRegistry = [..., myWidget]
```

Done — the widget appears in the grid automatically.

## Project Structure

```
dandy-dashboard/
├── cmd/server/main.go              # Entry point — wires config, registry, Echo
├── internal/
│   ├── config/                     # Env-based configuration
│   ├── widget/                     # Widget interface + registry
│   └── widgets/
│       ├── claude/                 # Claude AI chat (streaming SSE)
│       ├── japanese/               # Word of the day (Jotoba API + embedded wordlist)
│       └── calendar/               # Google Calendar events
└── web/src/
    ├── widgets/
    │   ├── types.ts                # WidgetDescriptor interface
    │   ├── registry.ts             # All widgets registered here
    │   ├── claude/
    │   ├── japanese/
    │   └── calendar/
    └── components/
        ├── WidgetCard.svelte       # Shared card shell (header + body)
        └── App.svelte              # Dashboard grid layout
```

## License

MIT
