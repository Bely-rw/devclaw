// Package commands – setup_skills.go provides embedded default skill templates
// that can be installed during the interactive setup wizard. Each skill is a
// SKILL.md file (ClawdHub/OpenClaw format) that the agent reads as instructions.
package commands

import (
	"fmt"
	"os"
	"path/filepath"
)

// embeddedSkill holds a default skill template.
type embeddedSkill struct {
	Name        string
	Label       string // Display label for the setup wizard.
	Description string
	Content     string // Full SKILL.md content.
}

// defaultSkillTemplates returns the list of skills available during setup.
func defaultSkillTemplates() []embeddedSkill {
	return []embeddedSkill{
		{
			Name:        "web-search",
			Label:       "🌐 Web Search — search the web via Brave API or DuckDuckGo",
			Description: "Web search via Brave Search API or DuckDuckGo",
			Content: `---
name: web-search
description: "Search the web for current information using Brave Search API or DuckDuckGo"
---
# Web Search

You can search the web for current information.

## Using Brave Search API (preferred, if BRAVE_API_KEY is available)

` + "```bash" + `
# Web search
curl -s "https://api.search.brave.com/res/v1/web/search?q=QUERY&count=5" \
  -H "Accept: application/json" \
  -H "X-Subscription-Token: $BRAVE_API_KEY" | jq '.web.results[] | {title, url, description}'

# News search
curl -s "https://api.search.brave.com/res/v1/web/search?q=QUERY&count=5&freshness=week&news=true" \
  -H "Accept: application/json" \
  -H "X-Subscription-Token: $BRAVE_API_KEY" | jq '.web.results[] | {title, url, description}'
` + "```" + `

## Using DuckDuckGo (no API key needed, fallback)

` + "```bash" + `
curl -s "https://html.duckduckgo.com/html/?q=QUERY" | grep -oP 'class="result__a"[^>]*href="\K[^"]+' | head -5
` + "```" + `

## Tips
- URL-encode the query (replace spaces with +).
- Use freshness parameter for time-filtered results: day, week, month.
- Be specific in queries for better results.
- Check if BRAVE_API_KEY is set; if not, fall back to DuckDuckGo.
- Combine with web_fetch to read the full content of interesting results.
`,
		},
		{
			Name:        "web-fetch",
			Label:       "📄 Web Fetch — fetch and extract readable content from URLs",
			Description: "Fetch URL content and extract readable text/markdown",
			Content: `---
name: web-fetch
description: "Fetch URL content and extract readable text"
---
# Web Fetch

You can fetch and read the content of any URL.

## Fetching a web page

` + "```bash" + `
# Fetch page content (text only, no HTML)
curl -sL "URL" | sed 's/<[^>]*>//g' | sed '/^$/d' | head -200

# Using readability-cli if installed
readable "URL" 2>/dev/null || curl -sL "URL" | sed 's/<[^>]*>//g' | sed '/^$/d' | head -200
` + "```" + `

## Fetching JSON APIs

` + "```bash" + `
curl -s "API_URL" -H "Accept: application/json" | jq '.'
` + "```" + `

## Tips
- Always use -sL (silent + follow redirects).
- For large pages, pipe through head -N to limit output.
- Strip HTML tags with sed for readability.
- Check Content-Type header to decide parsing strategy.
- Respect robots.txt and rate limits.
`,
		},
		{
			Name:        "github",
			Label:       "🐙 GitHub — issues, PRs, releases, CI via gh CLI",
			Description: "Full GitHub integration via gh CLI",
			Content: `---
name: github
description: "GitHub integration via gh CLI"
metadata: {"openclaw":{"requires":{"anyBins":["gh"]}}}
---
# GitHub

You can interact with GitHub using the gh CLI.

## Common operations

` + "```bash" + `
# List repos
gh repo list --limit 10

# View repo info
gh repo view OWNER/REPO

# Issues
gh issue list -R OWNER/REPO --limit 10
gh issue create -R OWNER/REPO --title "TITLE" --body "BODY"
gh issue view NUMBER -R OWNER/REPO

# Pull requests
gh pr list -R OWNER/REPO --limit 10
gh pr create -R OWNER/REPO --title "TITLE" --body "BODY"
gh pr view NUMBER -R OWNER/REPO
gh pr merge NUMBER -R OWNER/REPO --squash

# Releases
gh release list -R OWNER/REPO --limit 5
gh release create TAG -R OWNER/REPO --title "TITLE" --notes "NOTES"

# Actions / CI
gh run list -R OWNER/REPO --limit 5
gh run view RUN_ID -R OWNER/REPO

# Gists
gh gist list
gh gist create FILE --public --desc "DESCRIPTION"
` + "```" + `

## Tips
- Use -R OWNER/REPO to target a specific repo.
- Use --json to get structured output: gh issue list --json number,title,state
- Use jq for filtering: gh issue list --json number,title | jq '.[] | select(.title | contains("bug"))'
- Check if gh is authenticated: gh auth status
`,
		},
		{
			Name:        "weather",
			Label:       "🌤  Weather — forecasts via wttr.in (no API key needed)",
			Description: "Weather information and forecasts (no API key required)",
			Content: `---
name: weather
description: "Weather information and forecasts using wttr.in"
metadata: {"openclaw":{"always":true}}
---
# Weather

You can check weather using wttr.in (no API key needed).

## Current weather

` + "```bash" + `
# Current weather for a city
curl -s "wttr.in/CITY?format=3"

# Detailed current weather
curl -s "wttr.in/CITY?format=%l:+%c+%t+%h+%w+%p"

# Full forecast (3 days)
curl -s "wttr.in/CITY?lang=pt"
` + "```" + `

## JSON format (for parsing)

` + "```bash" + `
curl -s "wttr.in/CITY?format=j1" | jq '{
  location: .nearest_area[0].areaName[0].value,
  temp_c: .current_condition[0].temp_C,
  feels_like: .current_condition[0].FeelsLikeC,
  humidity: .current_condition[0].humidity,
  description: .current_condition[0].weatherDesc[0].value,
  wind_kmph: .current_condition[0].windspeedKmph
}'
` + "```" + `

## Tips
- Replace CITY with the city name (use + for spaces: New+York).
- Use lang=pt for Portuguese, lang=en for English.
- The user's timezone and location are in USER.md — use them as defaults.
- wttr.in supports airport codes (e.g. GRU, JFK).
`,
		},
		{
			Name:        "summarize",
			Label:       "📊 Summarize — summarize URLs, articles, and text",
			Description: "Summarize URLs, articles, videos, and long texts",
			Content: `---
name: summarize
description: "Summarize URLs, articles, and long texts"
metadata: {"openclaw":{"always":true}}
---
# Summarize

You can summarize web pages, articles, and long texts.

## Summarizing a URL

1. First, fetch the content:

` + "```bash" + `
curl -sL "URL" | sed 's/<[^>]*>//g' | sed '/^$/d' | head -500
` + "```" + `

2. Then summarize the extracted text using your own reasoning capabilities.

## Summarizing YouTube videos

` + "```bash" + `
# If yt-dlp is installed, get the transcript/subtitles
yt-dlp --write-auto-subs --skip-download --sub-lang pt,en -o "/tmp/%(id)s" "VIDEO_URL" 2>/dev/null
cat /tmp/*.vtt 2>/dev/null | grep -v "^[0-9]" | grep -v "^$" | grep -v "WEBVTT" | grep -v "-->" | sort -u | head -300
` + "```" + `

## Tips
- For long texts, break into sections and summarize each, then combine.
- Ask the user what level of detail they want (brief, detailed, bullet points).
- Preserve key facts, names, dates, and numbers.
- For technical content, keep important code snippets and terminology.
- Default to the user's language (check USER.md).
`,
		},
		{
			Name:        "timer",
			Label:       "⏱️  Timer — timers, alarmes e Pomodoro em segundo plano",
			Description: "Timers, alarms, and Pomodoro sessions",
			Content: `---
name: timer
description: "Set timers, alarms, and Pomodoro sessions"
---
# Timer

You can set timers that run in background. Use bash with background mode or the scheduler.

## Quick timers

` + "```bash" + `
# 5-minute timer
sleep 300 && echo "⏰ Timer de 5 minutos finalizado!"

# Custom message
sleep 600 && echo "⏰ Hora de verificar o forno!"

# 30 seconds
sleep 30 && echo "⏰ 30 segundos!"
` + "```" + `

> Run timers in background mode so the user can keep chatting.

## Pomodoro

` + "```bash" + `
# Work (25 min)
sleep 1500 && echo "🍅 Pomodoro finalizado! Pausa de 5 min."
# Break (5 min)
sleep 300 && echo "🔔 Pausa acabou! Volte ao trabalho."
` + "```" + `

## Time reference
| Input | Seconds |
|-------|---------|
| 30s | 30 |
| 1m | 60 |
| 5m | 300 |
| 15m | 900 |
| 25m | 1500 |
| 1h | 3600 |

## Tips
- Always run in background so user can keep chatting.
- Convert natural language: "5 minutos" = sleep 300.
- For recurring timers, use the scheduler with cron expressions.
- Notify clearly when timer completes.
`,
		},
		{
			Name:        "reminders",
			Label:       "🔔 Reminders — lembretes com data e hora",
			Description: "Time-based reminders with scheduling",
			Content: `---
name: reminders
description: "Create and manage time-based reminders"
---
# Reminders

Create reminders using the GoClaw scheduler (cron_add).

## Creating reminders

` + "```bash" + `
# Reminder at 3pm today (cron: minute hour day month weekday)
cron_add --id "rem-123" --schedule "0 15 14 2 *" --payload "📋 Reunião às 15h"

# Daily at 9am
cron_add --id "daily-water" --schedule "0 9 * * *" --payload "💧 Beber água!"

# Weekdays at 8:30am
cron_add --id "standup" --schedule "30 8 * * 1-5" --payload "🏃 Standup em 30min!"

# Weekly (Monday 10am)
cron_add --id "review" --schedule "0 10 * * 1" --payload "📊 Revisão semanal"

# List and remove
cron_list
cron_remove --id "rem-123"
` + "```" + `

## Natural language → cron
| User says | Cron |
|-----------|------|
| todo dia 8h | 0 8 * * * |
| seg a sex 9h | 0 9 * * 1-5 |
| toda segunda | 0 9 * * 1 |
| dia 15/mês | 0 9 15 * * |

## Tips
- Generate unique IDs for each reminder.
- For less than 1 hour, use the timer skill instead.
- Always confirm time with user before creating.
- Use user's timezone from config.
`,
		},
		{
			Name:        "notes",
			Label:       "📝 Notes — notas rápidas, listas e ideias",
			Description: "Quick notes, lists, and ideas stored locally",
			Content: `---
name: notes
description: "Quick notes, lists, and ideas — stored as local markdown"
---
# Notes

Save and manage notes as markdown files in ~/.goclaw/notes/.

## Creating notes

` + "```bash" + `
mkdir -p ~/.goclaw/notes

# Quick note
cat > ~/.goclaw/notes/$(date +%Y%m%d-%H%M%S)-note.md << 'EOF'
# Quick note
Content here.
EOF

# Shopping list
cat > ~/.goclaw/notes/shopping-list.md << 'EOF'
# Shopping List
- [ ] Leite
- [ ] Pão
- [ ] Ovos
EOF

# Append to list
echo "- [ ] Café" >> ~/.goclaw/notes/shopping-list.md
` + "```" + `

## Reading & searching

` + "```bash" + `
ls -lt ~/.goclaw/notes/ | head -20
cat ~/.goclaw/notes/shopping-list.md
grep -rl "TERM" ~/.goclaw/notes/
` + "```" + `

## Editing

` + "```bash" + `
# Mark todo as done
sed -i 's/- \[ \] Leite/- [x] Leite/' ~/.goclaw/notes/shopping-list.md
` + "```" + `

## Tips
- Use descriptive filenames for easy retrieval.
- Checkboxes: - [ ] todo, - [x] done.
- Read back after creating for confirmation.
- Tags at bottom: Tags: #work #urgent.
`,
		},
		{
			Name:        "translate",
			Label:       "🌍 Translate — traduções entre idiomas",
			Description: "Translate text between languages",
			Content: `---
name: translate
description: "Translate text between any languages"
---
# Translate

Translate text using your multilingual capabilities. For verification, use external APIs.

## Built-in translation (preferred)
As a multilingual LLM, translate directly when asked. Fast and accurate for most use cases.

## External verification (LibreTranslate)

` + "```bash" + `
curl -s -X POST "https://libretranslate.com/translate" \
  -H "Content-Type: application/json" \
  -d '{"q": "TEXT", "source": "en", "target": "pt"}' | jq -r '.translatedText'

# Detect language
curl -s -X POST "https://libretranslate.com/detect" \
  -H "Content-Type: application/json" \
  -d '{"q": "TEXT"}' | jq '.[0]'
` + "```" + `

## Common language codes
| Language | Code |
|----------|------|
| Portuguese | pt |
| English | en |
| Spanish | es |
| French | fr |
| German | de |
| Japanese | ja |
| Chinese | zh |

## Tips
- For casual translations, use built-in capabilities.
- Preserve formatting during translation.
- Don't translate proper nouns unless asked.
- For technical/legal text, suggest professional review.
`,
		},
	}
}

// installEmbeddedSkills copies selected skill templates to the skills directory.
func installEmbeddedSkills(selectedNames []string) {
	if len(selectedNames) == 0 {
		return
	}

	skillsDir := "./skills"
	templates := defaultSkillTemplates()

	// Build a lookup map.
	templateMap := make(map[string]embeddedSkill, len(templates))
	for _, t := range templates {
		templateMap[t.Name] = t
	}

	fmt.Println()
	fmt.Println("  Installing selected skills...")

	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		fmt.Printf("    ✗ Failed to create skills directory: %v\n", err)
		return
	}

	installed := 0
	for _, name := range selectedNames {
		tmpl, ok := templateMap[name]
		if !ok {
			fmt.Printf("    ✗ %s — unknown skill\n", name)
			continue
		}

		targetDir := filepath.Join(skillsDir, name)
		if err := os.MkdirAll(targetDir, 0o755); err != nil {
			fmt.Printf("    ✗ %s — %v\n", name, err)
			continue
		}

		skillFile := filepath.Join(targetDir, "SKILL.md")

		// Don't overwrite existing skills.
		if _, err := os.Stat(skillFile); err == nil {
			fmt.Printf("    • %s — already exists, skipping\n", name)
			installed++
			continue
		}

		if err := os.WriteFile(skillFile, []byte(tmpl.Content), 0o644); err != nil {
			fmt.Printf("    ✗ %s — %v\n", name, err)
			continue
		}

		fmt.Printf("    ✓ %s\n", name)
		installed++
	}

	fmt.Printf("  %d/%d skill(s) installed.\n", installed, len(selectedNames))
}
