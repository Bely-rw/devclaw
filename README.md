# GoClaw

**Open-source personal AI assistant in Go.** CLI + messaging channels (WhatsApp, Discord, Telegram). Built on the [AgentGo](https://github.com/jholhewres/agent-go) SDK.

Single binary. Zero runtime dependencies. Cross-compilable.

## Why GoClaw

[OpenClaw](https://github.com/openclaw/openclaw) is an impressive project — 18+ channels, 45+ skills, 188K stars. But it's 52+ modules on Node.js, ~500MB in memory, and ~5s to start.

[NanoClaw](https://github.com/qwibitai/nanoclaw) nailed the philosophy — small, understandable, secure by isolation. But it's still Node.js, still Baileys for WhatsApp, still tied to Claude Code.

GoClaw gives you the same core functionality compiled into a single binary you can `scp` to any server.

| | NanoClaw | OpenClaw | **GoClaw** |
|---|----------|----------|------------|
| Single binary | ❌ Node.js | ❌ Node.js | ✅ |
| Memory footprint | ~200MB | ~500MB | **~30MB** |
| Startup time | ~2s | ~5s | **~50ms** |
| Runtime deps | Node.js 20+ | Node.js 22+ | **None** |
| Cross-compile | ❌ | ❌ | ✅ Linux/Mac/Win/ARM |
| WhatsApp | Baileys (JS) | Baileys (JS) | **Whatsmeow (native Go)** |
| LLM providers | Claude only | 15+ | **15+ (via AgentGo SDK)** |

## Quick Start

### Install

```bash
go install github.com/jholhewres/goclaw/cmd/copilot@latest
```

Or build from source:

```bash
git clone https://github.com/jholhewres/goclaw.git
cd goclaw
make build
./bin/copilot --version
```

### Chat (CLI)

```bash
export OPENAI_API_KEY=sk-...

# Single message
copilot chat "What's on my calendar today?"

# Interactive REPL
copilot chat
```

### Serve (daemon with channels)

```bash
copilot config init
# Edit config.yaml, then:
copilot serve --channel whatsapp
```

A QR code will be displayed on first run to link your WhatsApp.

## Architecture

```
┌──────────────────────────────────────────────────────┐
│                       GoClaw                          │
├──────────────────────────────────────────────────────┤
│  CLI (cmd/copilot/)                                   │
│  chat · serve · schedule · skill · config · remember  │
├──────────────────────────────────────────────────────┤
│  Channels        │  Copilot Core     │  Skills        │
│  ├── WhatsApp    │  ├── Assistant    │  ├── Registry  │
│  ├── Discord     │  ├── Prompt (8L)  │  ├── Loader    │
│  └── Telegram    │  ├── Sessions     │  └── Index     │
│                  │  └── Security     │                │
├──────────────────────────────────────────────────────┤
│  AgentGo SDK (github.com/jholhewres/agent-go)        │
│  Agent · Models · Tools · Memory · Hooks · Guardrails │
└──────────────────────────────────────────────────────┘
```

GoClaw **does not reimplement** the AI core — it uses the AgentGo SDK directly for agent execution, LLM models, tools, and memory.

## Connecting to AgentGo SDK

### Model + Agent

Any provider supported by the AgentGo SDK works out of the box:

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/jholhewres/agent-go/pkg/agentgo/agent"
    "github.com/jholhewres/agent-go/pkg/agentgo/models/openai"
    "github.com/jholhewres/agent-go/pkg/agentgo/tools/calculator"
    "github.com/jholhewres/agent-go/pkg/agentgo/tools/toolkit"
)

func main() {
    // 1. Create a model (any AgentGo provider)
    model, _ := openai.New("gpt-4o-mini", openai.Config{
        APIKey: os.Getenv("OPENAI_API_KEY"),
    })

    // 2. Create an agent with tools
    ag, _ := agent.New(agent.Config{
        Name:         "Copilot",
        Model:        model,
        Toolkits:     []toolkit.Toolkit{calculator.New()},
        Instructions: "You are a helpful assistant. Be concise.",
    })

    // 3. Run
    output, _ := ag.Run(context.Background(), "What is 42 * 58?")
    fmt.Println(output.Content)
}
```

### Supported Providers

| Provider | Import | Example model |
|----------|--------|---------------|
| OpenAI | `models/openai` | `gpt-4o-mini`, `gpt-4o` |
| Anthropic | `models/anthropic` | `claude-3-5-sonnet-20241022` |
| Google Gemini | `models/gemini` | `gemini-1.5-pro` |
| Ollama (local) | `models/ollama` | `llama2`, `mistral` |
| DeepSeek | `models/deepseek` | `deepseek-chat` |
| Groq | `models/groq` | `llama-3.1-70b-versatile` |
| Together | `models/together` | `meta-llama/Llama-3-70b` |
| OpenRouter | `models/openrouter` | Any model |
| LM Studio | `models/lmstudio` | Local models |

```go
// Anthropic Claude
model, _ := anthropic.New("claude-3-5-sonnet-20241022", anthropic.Config{
    APIKey: os.Getenv("ANTHROPIC_API_KEY"),
})

// Ollama (local, no API key)
model, _ := ollama.New("llama2", ollama.Config{
    BaseURL: "http://localhost:11434",
})
```

### Tools

GoClaw inherits all built-in tools from AgentGo and adds its own via Skills:

```go
import (
    "github.com/jholhewres/agent-go/pkg/agentgo/tools/calculator"
    "github.com/jholhewres/agent-go/pkg/agentgo/tools/websearch"
    "github.com/jholhewres/agent-go/pkg/agentgo/tools/toolkit"
)

// Built-in AgentGo tools
toolkits := []toolkit.Toolkit{
    calculator.New(),
    websearch.New(websearch.Config{APIKey: "..."}),
}

// Custom tool
myTool := toolkit.NewBaseToolkit("my_tool")
myTool.RegisterFunction(&toolkit.Function{
    Name:        "get_price",
    Description: "Get the current price of a product",
    Parameters: map[string]toolkit.Parameter{
        "product": {Type: "string", Description: "Product name", Required: true},
    },
    Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
        product := args["product"].(string)
        return fmt.Sprintf("The price of %s is $99.90", product), nil
    },
})
```

### Memory

```go
import "github.com/jholhewres/agent-go/pkg/agentgo/memory"

// Simple in-memory (default)
mem := memory.NewInMemory(100)

// Hybrid memory with vector DB for RAG
hybridMem := memory.NewHybridMemory(memory.HybridConfig{
    ShortTermSize: 50,
    VectorDB:      chromaDB,
    Embeddings:    embeddingFunc,
})
```

### Hooks & Guardrails

```go
import (
    "github.com/jholhewres/agent-go/pkg/agentgo/hooks"
    "github.com/jholhewres/agent-go/pkg/agentgo/guardrails"
)

// Pre-execution logging hook
preHook := hooks.HookFunc(func(ctx context.Context, input *hooks.HookInput) error {
    log.Printf("Input: %s", input.Input)
    return nil
})

// Prompt injection guardrail
injectionGuard := guardrails.NewPromptInjectionGuardrail()

ag, _ := agent.New(agent.Config{
    Model:     model,
    PreHooks:  []hooks.Hook{injectionGuard, preHook},
    PostHooks: []hooks.Hook{urlValidator},
})
```

## Channels

### WhatsApp

Uses [whatsmeow](https://go.mau.fi/whatsmeow) — native Go, no Node.js or Baileys.

```yaml
# config.yaml
channels:
  whatsapp:
    enabled: true
    session_dir: "./sessions/whatsapp"
    trigger: "@copilot"
```

```
User: @copilot how many meetings do I have today?
Copilot: You have 3 meetings today:
         • 9am - Team standup
         • 2pm - Client presentation
         • 4pm - Code review

User: @copilot remind me to call John at 3pm
Copilot: ✅ Reminder set for 3pm: "call John"
```

### Discord

Uses [discordgo](https://github.com/bwmarrin/discordgo).

```yaml
channels:
  discord:
    enabled: true
    token: "${DISCORD_TOKEN}"
    trigger: "!copilot"
```

### Telegram

Uses [telego](https://github.com/mymmrac/telego).

```yaml
channels:
  telegram:
    enabled: true
    token: "${TELEGRAM_TOKEN}"
    trigger: "/copilot"
```

## Skills

Skills are modules that add capabilities to the assistant. They can be built-in, installed from the registry, or community-contributed.

### CLI Management

```bash
# Search available skills
copilot skill search calendar

# Install a skill
copilot skill install github.com/jholhewres/goclaw-skills/calendar

# List installed
copilot skill list

# Update all
copilot skill update --all
```

### Creating a Skill

```yaml
# skill.yaml
name: my-weather
version: 1.0.0
author: your-name
description: Weather information
category: builtin
tags: [weather, forecast]

tools:
  - name: get_weather
    description: Get weather for a city
    parameters:
      city:
        type: string
        required: true

system_prompt: |
  You can check weather for any city using get_weather.

triggers:
  - "weather in"
  - "what's the weather"
```

```go
// skill.go
package weather

import (
    "context"
    "github.com/jholhewres/goclaw/pkg/goclaw/skills"
)

type WeatherSkill struct{}

func (s *WeatherSkill) Metadata() skills.Metadata {
    return skills.Metadata{
        Name:        "weather",
        Version:     "1.0.0",
        Description: "Weather information",
        Category:    "builtin",
    }
}

func (s *WeatherSkill) Tools() []skills.Tool {
    return []skills.Tool{
        {
            Name:        "get_weather",
            Description: "Get current weather for a city",
            Parameters: []skills.ToolParameter{
                {Name: "city", Type: "string", Required: true},
            },
            Handler: s.getWeather,
        },
    }
}

func (s *WeatherSkill) getWeather(ctx context.Context, args map[string]any) (any, error) {
    city := args["city"].(string)
    return map[string]any{"city": city, "temp": 25, "condition": "sunny"}, nil
}
```

## Scheduler

Schedule recurring tasks with cron expressions:

```bash
# Daily briefing at 9am on weekdays via WhatsApp
copilot schedule add "0 9 * * 1-5" "Send me a daily briefing" --channel whatsapp

# Weekly summary every Friday at 5pm
copilot schedule add "0 17 * * 5" "Generate a weekly summary" --channel whatsapp

# List schedules
copilot schedule list

# Remove
copilot schedule remove <id>
```

## Security

GoClaw applies guardrails at every stage of the message flow:

| Stage | Protection |
|-------|-----------|
| **Input** | Rate limiting (sliding window per user), prompt injection detection, max input length |
| **Session** | Isolated per chat/group, auto-pruning of inactive sessions |
| **Prompt** | 8-layer system with token budget, no unbounded context |
| **Tools** | Whitelist per skill, confirmation for destructive actions |
| **Output** | System prompt leak detection, empty response fallback |
| **Deploy** | systemd hardening (ProtectSystem, PrivateTmp, MemoryMax) |

## Configuration

```yaml
# config.yaml
assistant:
  name: "Copilot"
  trigger: "@copilot"
  model: "gpt-4o-mini"          # Any AgentGo model
  timezone: "America/Sao_Paulo"
  language: "en"
  instructions: |
    You are a helpful personal assistant.
    Be concise and practical.

channels:
  whatsapp:
    enabled: true
    session_dir: "./sessions/whatsapp"
  discord:
    enabled: false
    token: "${DISCORD_TOKEN}"
  telegram:
    enabled: false
    token: "${TELEGRAM_TOKEN}"

memory:
  type: "sqlite"                # sqlite, postgres, memory
  path: "./data/memory.db"
  max_messages: 100
  compression_strategy: "summarize"  # summarize, truncate, semantic

security:
  max_input_length: 4096
  rate_limit: 30                # messages/min/user
  enable_pii_detection: false
  enable_url_validation: true

scheduler:
  enabled: true
  storage: "./data/scheduler.db"

skills:
  builtin:
    - weather
    - calculator
    - websearch
```

## Deploy

### Docker

```bash
docker compose up -d
docker compose logs -f copilot
```

### systemd

```bash
sudo cp copilot.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now copilot
```

### Direct binary

```bash
make build
./bin/copilot serve --config config.yaml
```

## CLI Reference

| Command | Description |
|---------|-------------|
| `copilot chat [msg]` | Interactive chat or single message |
| `copilot serve` | Start daemon with messaging channels |
| `copilot schedule list` | List scheduled tasks |
| `copilot schedule add <cron> <cmd>` | Add a scheduled task |
| `copilot schedule remove <id>` | Remove a scheduled task |
| `copilot skill list` | List installed skills |
| `copilot skill search <query>` | Search available skills |
| `copilot skill install <name>` | Install a skill |
| `copilot skill update [name\|--all]` | Update skills |
| `copilot config init` | Create default config |
| `copilot config show` | Show current config |
| `copilot config set <key> <value>` | Set a config value |
| `copilot remember <fact>` | Save a fact to long-term memory |
| `copilot health` | Check service health |

## Project Structure

```
goclaw/
├── cmd/copilot/              # CLI application
│   ├── main.go
│   └── commands/             # Cobra commands
├── pkg/goclaw/
│   ├── channels/             # Channel interface + Manager
│   ├── skills/               # Skill interface + Registry
│   ├── copilot/              # Assistant + Prompt + Session
│   │   └── security/         # I/O guardrails
│   └── scheduler/            # Cron-based job scheduling
├── skills/                   # Submodule → goclaw-skills
├── configs/                  # Example configs
├── docs/                     # Plans & specs
├── Dockerfile
├── docker-compose.yml
├── Makefile
└── go.mod
```

## Key Dependencies

| Package | Purpose |
|---------|---------|
| [agent-go](https://github.com/jholhewres/agent-go) | Agent SDK (models, tools, memory, hooks) |
| [whatsmeow](https://go.mau.fi/whatsmeow) | WhatsApp (native Go) |
| [discordgo](https://github.com/bwmarrin/discordgo) | Discord |
| [telego](https://github.com/mymmrac/telego) | Telegram |
| [cobra](https://github.com/spf13/cobra) | CLI framework |
| [cron](https://github.com/robfig/cron) | Task scheduler |

## Roadmap

- [x] Core scaffolding: channels, skills, scheduler, assistant, security
- [x] CLI: chat, serve, schedule, skill, config, remember, health
- [x] Prompt composer with 8 layers and token budget
- [x] Security guardrails (input + output + tool policy)
- [x] Session isolation with auto-pruning
- [x] Docker + systemd + Makefile
- [x] Skills repository as submodule
- [ ] WhatsApp channel implementation (whatsmeow)
- [ ] Discord channel implementation (discordgo)
- [ ] Telegram channel implementation (telego)
- [ ] Full AgentGo SDK integration (agent.Run in message loop)
- [ ] Memory persistence (SQLite)
- [ ] RAG with embeddings
- [ ] 10+ skills in the repository
- [ ] Web dashboard
- [ ] Multi-agent teams

## License

MIT
