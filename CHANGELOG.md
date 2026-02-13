# Changelog

All notable changes to GoClaw are documented in this file.

## [1.0.0] — 2026-02-12

First stable release.

### Core

- Agent loop with multi-turn tool calling, auto-continue (up to 2 continuations), and reflection nudges every 8 turns
- 8-layer prompt composer (Core, Safety, Identity, Thinking, Bootstrap, Business, Skills, Memory, Temporal, Conversation, Runtime) with priority-based token budget trimming
- Session isolation per chat/group with JSONL persistence and auto-pruning
- Three compression strategies for session compaction: `summarize` (LLM), `truncate`, `sliding` — preventive trigger at 80% capacity
- Token-aware sliding window for conversation history (backwards construction, per-message truncation)
- Subagent system: spawn, track, wait, stop child agents with filtered tool sets
- Message queue with per-session debounce (configurable ms), deduplication (5s window), and burst handling
- Config hot-reload via file watcher (mtime + SHA256 hash)
- Token/cost tracking per session and global with model-specific pricing

### LLM Client

- OpenAI-compatible HTTP client with provider auto-detection
- Providers: OpenAI, Z.AI (API, Coding, Anthropic proxy), Anthropic, any OpenAI-compatible
- Model-specific defaults for temperature and max_tokens (GPT-5, Claude Opus 4.6, GLM-5, etc.)
- Model fallback chain with exponential backoff, `Retry-After` header support, and error classification
- SSE streaming with `[DONE]` terminator handling
- Prompt caching (`cache_control: ephemeral`) for Anthropic and Z.AI Anthropic proxy
- Context overflow handling: auto-compaction, tool result truncation, retry

### Security

- Encrypted vault (AES-256-GCM + Argon2id key derivation: 64 MB, 3 iterations, 4 threads)
- Secret resolution chain: vault → OS keyring → env vars → .env → config.yaml
- Tool Guard: per-tool role permissions (owner/admin/user), dangerous command regex blocking, protected paths, SSH host allowlist, audit logging
- Interactive execution approval via chat (`/approve`, `/deny`)
- SSRF protection for `web_fetch`: blocks private IPs, loopback, link-local, cloud metadata
- Script sandbox: none / Linux namespaces (PID, mount, net, user) / Docker container
- Pre-execution content scanning: eval, reverse shells, crypto mining, shell injection, obfuscation
- Input guardrails: rate limiting (sliding window), prompt injection detection, max input length
- Output guardrails: system prompt leak detection, empty response fallback

### Tools (35+)

- File I/O: `read_file`, `write_file`, `edit_file`, `list_files`, `search_files`, `glob_files` — full filesystem access
- Shell: `bash` (persistent cwd/env), `set_env`, `ssh`, `scp`
- Web: `web_search` (DuckDuckGo HTML parsing), `web_fetch` (SSRF-protected)
- Memory: `memory_save`, `memory_search`, `memory_list`
- Scheduler: `schedule_add`, `schedule_list`, `schedule_remove`
- Media: `describe_image` (LLM vision), `transcribe_audio` (Whisper API)
- Skills: `init_skill`, `edit_skill`, `add_script`, `list_skills`, `test_skill`, `install_skill`, `search_skills`, `remove_skill`
- Subagents: `spawn_subagent`, `list_subagents`, `wait_subagent`, `stop_subagent`
- Parallel tool execution with configurable semaphore (max 5 concurrent)

### Channels

- WhatsApp (native Go via whatsmeow): text, images, audio, video, documents, stickers, voice notes, locations, contacts, reactions, reply/quoting, typing indicators, read receipts, group messages
- Automatic media enrichment: vision (image description) and audio transcription (Whisper) on incoming media
- WhatsApp markdown formatting (bold, italic, strikethrough, code, code blocks)
- Message splitting for long responses (preserves code blocks, prefers paragraph/sentence boundaries)
- Plugin loader for additional channels (Go native `.so`)

### Access Control & Workspaces

- Per-user/group allowlist and blocklist with deny-by-default policy
- Roles: owner > admin > user > blocked
- Chat commands: `/allow`, `/block`, `/admin`, `/users`, `/group allow`
- Multi-tenant workspaces with isolated system prompts, skills, models, languages, and memory
- Workspace management via chat: `/ws create`, `/ws assign`, `/ws list`

### Skills

- Native Go skills (compiled, direct execution) and SKILL.md format (ClawHub/OpenClaw compatible)
- Skill installation from ClawHub, GitHub, HTTP URLs, and local paths
- Built-in skills: weather, calculator, web-search, web-fetch, summarize, github, gog, calendar
- Skill creation via chat (agent can author its own skills)
- Hot-reload on install/remove

### HTTP API Gateway

- OpenAI-compatible `POST /v1/chat/completions` with SSE streaming
- Session management: `GET/DELETE /api/sessions`
- Usage tracking: `GET /api/usage`
- System status: `GET /api/status`
- Webhook registration: `POST /api/webhooks`
- Bearer token authentication and CORS

### CLI

- Interactive setup wizard with arrow-key navigation (`charmbracelet/huh`), model auto-detection, vault creation
- CLI chat REPL with `readline` support: arrow-key history, reverse search (Ctrl+R), tab completion, persistent history
- Commands: `setup`, `serve`, `chat`, `config` (init, show, validate, vault-*, set-key, key-status), `skill` (list, search, install, create), `schedule` (list, add), `health`
- Chat commands: `/help`, `/model`, `/usage`, `/compact`, `/think`, `/new`, `/reset`, `/stop`, `/approve`, `/deny`

### Bootstrap System

- Template files in `configs/bootstrap/`: SOUL.md, AGENTS.md, IDENTITY.md, USER.md, TOOLS.md, HEARTBEAT.md
- Loaded at runtime into the prompt layer system
- Agent can read and update its own bootstrap files

### Scheduler & Heartbeat

- Cron-based task scheduler with file persistence
- Heartbeat: proactive agent behavior on configurable interval with active hours
- Agent reads HEARTBEAT.md for pending tasks

### Deployment

- Single binary, zero runtime dependencies
- Docker and Docker Compose support
- systemd service unit with hardening (ProtectSystem, PrivateTmp, MemoryMax)
- Makefile: build, run, setup, chat, test, lint, clean, docker-build, docker-up
