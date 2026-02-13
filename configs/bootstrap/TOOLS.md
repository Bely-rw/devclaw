# TOOLS.md — Local Notes

Skills define _how_ tools work. This file is for _your_ specifics — the stuff that's unique to your setup.

## What Goes Here

Things like:

- SSH hosts and aliases
- Server configurations
- Preferred voices for TTS
- Device nicknames
- API endpoints
- Anything environment-specific

## Examples

```
### SSH

- home-server → 192.168.1.100, user: admin
- prod → deploy@prod.example.com, port 2222
- staging → deploy@staging.example.com

### Servers

- Database: PostgreSQL on home-server:5432
- Redis: home-server:6379

### API Keys

- Weather API: stored in vault as "weather_api_key"
- GitHub token: stored in vault as "github_token"
```

## Why Separate?

Skills are shared. Your setup is yours. Keeping them apart means you can update skills without losing your notes, and share skills without leaking your infrastructure.

---

Add whatever helps you do your job. This is your cheat sheet.
