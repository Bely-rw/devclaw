# GoClaw

**AI assistant. Any OS.**

Open-source personal AI assistant framework in Go. Single binary, zero runtime dependencies, cross-compilable to Linux, macOS, Windows, ARM, and anything Go targets. CLI + messaging channels (WhatsApp, Discord, Telegram). Built on the [AgentGo](https://github.com/jholhewres/agent-go) SDK.

## About

GoClaw is heavily inspired by [OpenClaw](https://github.com/openclaw/openclaw) — a project that truly revolutionized what personal AI assistants can be. OpenClaw pioneered the concept of an extensible agent with community-driven skills, layered prompts, and multi-channel support. It demonstrated that an AI assistant is more than just a chatbot: it's a platform where skills, tools, and integrations compose together to create something genuinely useful in daily life.

GoClaw takes that same vision and brings it to Go: a single compiled binary with native WhatsApp support, a plugin system for channels, and a multi-tenant workspace architecture — all while keeping OpenClaw's skill ecosystem accessible through a compatibility layer.

## Why Go, and Why Security Matters

One core concern drove the design of GoClaw: **security in an open-source skill ecosystem**.

OpenClaw's community skill repository has 900+ skills contributed by hundreds of developers. That's incredible — but it also means no single maintainer can audit, verify, and keep up with every script that runs inside the assistant. A malicious or poorly written skill could access your files, exfiltrate data, mine crypto, or open reverse shells.

Creating, maintaining, and verifying that many skills is simply not feasible at scale. So instead of trying to trust every skill, **why not create a secure environment to execute them?**

That's the approach GoClaw takes:

- **Native Go skills** are compiled into the binary — fast, type-safe, and auditable
- **Community scripts** (Python, Node.js, Shell) run inside a **multi-level sandbox** with Linux namespace isolation or Docker containers
- **Environment filtering** blocks injection vectors (`LD_PRELOAD`, `NODE_OPTIONS`, `PYTHONPATH`, etc.)
- **Content scanning** detects `eval()`, reverse shells, crypto mining, and obfuscated code *before* execution
- **Network isolation** is the default — scripts can't phone home unless explicitly allowed

The result: you get access to the entire OpenClaw skill ecosystem, but every external script runs in a controlled, resource-limited environment. Trust the ecosystem, verify the execution.

## Quick Start

```bash
# Build from source
git clone https://github.com/jholhewres/goclaw.git
cd goclaw
make build

# Create config and set your phone number
make init
# Edit config.yaml → access.owners → your phone number

# Start (scan QR code on first run)
make run
```

Or install directly:

```bash
go install github.com/jholhewres/goclaw/cmd/copilot@latest
copilot config init
copilot serve
```

## Access Control

GoClaw does **not** respond to everyone. Only explicitly authorized contacts can interact with the assistant — just like OpenClaw.

```yaml
# config.yaml
access:
  default_policy: deny              # deny | allow | ask
  owners:
    - "5511999999999"               # Your phone number (full control)
  admins:
    - "5511888888888"               # Can manage users + workspaces
  allowed_users:
    - "5511777777777"               # Can interact with the bot
  allowed_groups:
    - "120363000000000000@g.us"     # Group JIDs
```

| Policy | Behavior |
|--------|----------|
| `deny` | Silently ignores unknown contacts (default, recommended) |
| `allow` | Responds to everyone except blocked contacts |
| `ask` | Sends a one-time "access not authorized" message |

Access levels: **owner** > **admin** > **user** > **blocked**.

Owners and admins can manage access in real time via chat commands:

```
/allow 5511777777777        Grant access
/block 5511666666666        Block a contact
/admin 5511888888888        Promote to admin
/users                      List all authorized users
/group allow                Allow the current group
/status                     Bot status
/help                       All commands
```

## Workspaces

Multiple people can use the same WhatsApp number with **completely isolated contexts**. Each workspace has its own system prompt, skills, LLM model, language, and conversation memory.

```yaml
workspaces:
  default_workspace: "default"
  workspaces:
    - id: "default"
      name: "Default"
      active: true

    - id: "personal"
      name: "Personal Assistant"
      instructions: |
        You are my personal assistant. Help with daily tasks,
        reminders, and planning. Be proactive.
      model: "gpt-4o"
      language: "pt-BR"
      skills: [weather, web-search, gog]
      members: ["5511999999999"]

    - id: "work"
      name: "Dev Team"
      instructions: |
        You are a technical assistant for our development team.
        Help with code reviews, architecture, and documentation.
      model: "gpt-4o-mini"
      language: "en"
      skills: [github, web-search, summarize]
      members: ["5511888888888", "5511777777777"]
      groups: ["120363000000000000@g.us"]
```

Managed via chat:

```
/ws create personal "My Assistant"
/ws assign 5511999999999 personal
/ws list
/ws info personal
```

## Architecture

GoClaw has three layers of extensibility: **Core**, **Plugins**, and **Skills**.

```
┌──────────────────────────────────────────────────────────────┐
│                           GoClaw                             │
├──────────────────────────────────────────────────────────────┤
│  CLI (cmd/copilot/)                                          │
│  chat · serve · schedule · skill · config · remember · health│
├───────────────┬──────────────────┬───────────────────────────┤
│  Core         │  Plugins (.so)   │  Skills (separate repo)   │
│  Compiled in  │  Loaded at       │  Installed via CLI         │
│  the binary   │  runtime         │  from goclaw-skills        │
│               │                  │                            │
│  ▸ WhatsApp   │  ▸ Extra         │  ▸ Prompt / Soul           │
│  ▸ Access     │    channels      │  ▸ Tools for the agent     │
│  ▸ Workspaces │  ▸ Webhooks      │  ▸ Triggers                │
│  ▸ Guardrails │  ▸ Custom        │  ▸ Config schema           │
│  ▸ Sessions   │    integrations  │  ▸ System prompt inject    │
│  ▸ Scheduler  │                  │                            │
│  ▸ Sandbox    │                  │                            │
├───────────────┴──────────────────┴───────────────────────────┤
│  AgentGo SDK (github.com/jholhewres/agent-go)                │
│  Agent · Models · Tools · Memory · Hooks · Guardrails        │
└──────────────────────────────────────────────────────────────┘
```

- **Core** = what GoClaw needs to run. WhatsApp, access control, workspaces, guardrails, sandbox — security is not optional.
- **Plugins** = runtime extensions via Go's native plugin system (`.so`). Add channels, webhooks, or integrations without recompiling.
- **Skills** = what the *agent* can do. A skill teaches the LLM new capabilities through prompt instructions, tools, and triggers. Skills live in a separate repository.

## Connecting to AgentGo SDK

GoClaw uses the [AgentGo SDK](https://github.com/jholhewres/agent-go) for agent execution, LLM models, tools, and memory.

```go
import (
    "github.com/jholhewres/agent-go/pkg/agentgo/agent"
    "github.com/jholhewres/agent-go/pkg/agentgo/models/openai"
)

model, _ := openai.New("gpt-4o-mini", openai.Config{
    APIKey: os.Getenv("OPENAI_API_KEY"),
})

ag, _ := agent.New(agent.Config{
    Name:         "Copilot",
    Model:        model,
    Instructions: "You are a helpful assistant.",
})

output, _ := ag.Run(context.Background(), "What's on my calendar?")
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

## Channels

### WhatsApp (core)

Native Go implementation using [whatsmeow](https://go.mau.fi/whatsmeow). Compiled into the binary — no Node.js, no Baileys.

Full support: text, images, audio, video, documents, stickers, voice notes, locations, contacts, reactions, reply/quoting, typing indicators, read receipts, group messages.

```yaml
channels:
  whatsapp:
    session_dir: "./sessions/whatsapp"
    trigger: "@copilot"
    respond_to_groups: true
    respond_to_dms: true
    auto_read: true
    send_typing: true
    media_dir: "./data/media"
    max_media_size_mb: 16
```

### Discord & Telegram (plugins)

Loaded as Go plugins (`.so`) at runtime. Not everyone needs every channel.

```go
// Build a channel plugin
// my_channel_plugin.go
package main

import "github.com/jholhewres/goclaw/pkg/goclaw/channels"

var Channel channels.Channel = &MyChannel{}

// go build -buildmode=plugin -o plugins/mychannel.so my_channel_plugin.go
```

## Skills

Skills teach the agent new capabilities. They live in a separate repository ([goclaw-skills](https://github.com/jholhewres/goclaw-skills)) and are managed via CLI.

### Two Formats

| Format | Files | Runtime | Security |
|--------|-------|---------|----------|
| **Native Go** | `skill.yaml` + `skill.go` | Compiled Go | Full trust (compiled) |
| **SKILL.md** | `SKILL.md` + `scripts/` | Python, Node.js, Shell | Sandboxed |

The SKILL.md format is fully compatible with [OpenClaw's skill repository](https://github.com/openclaw/skills) — GoClaw can run community skills from that ecosystem through its sandbox.

### Catalog

| Category | Skill | Tools | Requires |
|----------|-------|-------|----------|
| Builtin | **weather** | `get_weather`, `get_forecast`, `get_moon` | — |
| Builtin | **calculator** | `calculate`, `convert_units` | — |
| Data | **web-search** | `search`, `search_news` | `BRAVE_API_KEY` or `SEARXNG_URL` |
| Data | **web-fetch** | `fetch`, `fetch_headers`, `fetch_json` | — |
| Data | **summarize** | `summarize_url`, `transcribe_url`, `summarize_file` | `summarize` CLI |
| Development | **github** | 17 tools (issues, PRs, CI/CD, releases, search) | `gh` CLI |
| Productivity | **gog** | Gmail, Calendar, Drive (11 tools) | `gog` CLI |

See the full [Skills Catalog](docs/skills-catalog.md).

### CLI Management

```bash
copilot skill list                         # List installed
copilot skill search calendar              # Search available
copilot skill install calendar             # Install native Go skill
copilot skill install --from clawdhub 1password  # Install community skill (sandboxed)
copilot skill update --all                 # Update all
copilot skill create my-skill              # Scaffold a new skill
```

## Script Sandbox

This is the answer to the open-source security problem. When you have hundreds of community-contributed scripts, you can't audit them all. So you sandbox them.

### Isolation Levels

| Level | How | Use Case |
|-------|-----|----------|
| **none** | Direct `exec.Command` | Trusted/builtin skills only |
| **restricted** | Linux namespaces (PID, mount, network, user) | Community skills |
| **container** | Docker with purpose-built image | Untrusted scripts |

### What Gets Blocked

**Before execution** — script content is scanned for:

| Pattern | Severity | Examples |
|---------|----------|---------|
| Dangerous eval | Critical | `exec()`, `eval()`, `new Function()` |
| Shell injection | Critical | `subprocess.run(shell=True)` |
| Reverse shells | Critical | `/dev/tcp/`, `nc -e`, `socket.connect` |
| Crypto mining | Critical | `stratum+tcp`, `coinhive`, `xmrig` |
| Data exfiltration | Warning | Access to `/etc/passwd`, `.ssh/` |
| Obfuscation | Warning | Hex-encoded strings, `base64+exec` |

**During execution** — the sandbox enforces:

- Environment filtering (blocks `LD_PRELOAD`, `NODE_OPTIONS`, `PYTHONPATH`, etc.)
- Network isolation (no outbound connections by default)
- Resource limits (CPU, memory, timeout)
- Read-only filesystem (container mode)
- Non-root execution with dropped capabilities

```yaml
sandbox:
  default_isolation: restricted
  timeout: 60s
  max_memory_mb: 256
  allow_network: false
  docker:
    image: goclaw-sandbox:latest
    network: none
```

## Security

Security is applied at every stage of the message flow:

| Stage | Protection |
|-------|-----------|
| **Access** | Allowlist/blocklist, deny-by-default, per-user and per-group permissions |
| **Input** | Rate limiting, prompt injection detection, max input length |
| **Session** | Isolated per chat and per workspace, auto-pruning |
| **Prompt** | 8-layer system with token budget, no unbounded context |
| **Tools** | Whitelist per skill, confirmation for destructive actions |
| **Scripts** | Multi-level sandbox, env filtering, content scanning |
| **Output** | System prompt leak detection, empty response fallback |
| **Deploy** | systemd hardening (ProtectSystem, PrivateTmp, MemoryMax) |

## Configuration

```bash
make init          # Create config.yaml with defaults
make validate      # Validate without running
make run           # Build + serve (auto-detects config.yaml)
make run VERBOSE=1 # With debug logs
```

Full config reference: see [configs/copilot.example.yaml](configs/copilot.example.yaml).

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
./bin/copilot serve
```

## CLI Reference

| Command | Description |
|---------|-------------|
| `copilot chat [msg]` | Interactive chat or single message |
| `copilot serve` | Start daemon with messaging channels |
| `copilot config init` | Create default config |
| `copilot config show` | Show current config |
| `copilot config validate` | Validate config |
| `copilot skill list` | List installed skills |
| `copilot skill search <query>` | Search available skills |
| `copilot skill install <name>` | Install a skill |
| `copilot schedule list` | List scheduled tasks |
| `copilot schedule add <cron> <cmd>` | Add a scheduled task |
| `copilot remember <fact>` | Save to long-term memory |
| `copilot health` | Check service health |

## Project Structure

```
goclaw/
├── cmd/copilot/                # CLI application
│   └── commands/               # Cobra commands
├── pkg/goclaw/
│   ├── channels/               # Channel interface + Manager
│   │   └── whatsapp/           # WhatsApp (whatsmeow, core)
│   ├── copilot/                # Assistant orchestrator
│   │   ├── access.go           # Access control (allowlist/blocklist)
│   │   ├── workspace.go        # Multi-tenant workspaces
│   │   ├── commands.go         # Admin commands via chat
│   │   ├── assistant.go        # Main message flow
│   │   ├── prompt_layers.go    # 8-layer prompt composer
│   │   ├── session.go          # Session isolation
│   │   ├── loader.go           # YAML config loader
│   │   └── security/           # I/O guardrails
│   ├── plugins/                # Go native plugin loader (.so)
│   ├── sandbox/                # Script sandbox (namespaces/Docker)
│   ├── skills/                 # Skill system + ClawdHub loader
│   └── scheduler/              # Cron-based scheduling
├── skills/                     # Submodule → goclaw-skills
├── configs/                    # Example configs
├── Makefile
├── Dockerfile
├── docker-compose.yml
├── copilot.service
└── go.mod
```

## Key Dependencies

| Package | Purpose |
|---------|---------|
| [agent-go](https://github.com/jholhewres/agent-go) | Agent SDK (models, tools, memory, hooks) |
| [whatsmeow](https://go.mau.fi/whatsmeow) | WhatsApp (native Go, core) |
| [cobra](https://github.com/spf13/cobra) | CLI framework |
| [cron](https://github.com/robfig/cron) | Task scheduler |
| [yaml.v3](https://gopkg.in/yaml.v3) | Configuration |

No external dependencies for the sandbox — uses Go's `os/exec`, `syscall` (Linux namespaces), and Docker CLI.

## Roadmap

- [x] Core: channels, skills, scheduler, assistant, security guardrails
- [x] CLI: chat, serve, schedule, skill, config, remember, health
- [x] Prompt composer (8 layers with token budget)
- [x] Session isolation with auto-pruning
- [x] WhatsApp channel (whatsmeow — text, media, audio, video, docs, stickers, reactions)
- [x] Plugin loader (Go native `.so`)
- [x] Access control (allowlist, blocklist, deny-by-default)
- [x] Multi-tenant workspaces with isolated memory
- [x] Admin commands via chat (/allow, /block, /ws, /group)
- [x] YAML config loader with auto-discovery
- [x] Script sandbox (none / Linux namespaces / Docker)
- [x] OpenClaw SKILL.md compatibility layer
- [x] 10+ skills: weather, calculator, github, web-search, web-fetch, summarize, gog
- [ ] Full AgentGo SDK integration (agent.Run in message loop)
- [ ] Discord channel plugin
- [ ] Telegram channel plugin
- [ ] Memory persistence (SQLite)
- [ ] `copilot skill install --from clawdhub` implementation
- [ ] RAG with embeddings
- [ ] Web dashboard
- [ ] Multi-agent teams

## License

MIT
