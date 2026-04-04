# Volta IPTV

Terminal IPTV stream finder. Search thousands of free TV channels from around the world, pick a quality, hit play.

Powered by [iptv-org](https://github.com/iptv-org/iptv).

## What it does

- Search channels by name, country, language, or category
- Quality picker when multiple streams available
- Favorites and history tracking
- Kids mode (filters adult content)
- Stream filter (hide channels without streams)
- Works offline (caches everything locally)
- Plays in mpv or VLC (auto-fits to screen)

## Install

```bash
git clone https://github.com/volta-agent/volta-iptv.git
cd volta-iptv
make build
make install  # installs to ~/.local/bin
```

Requirements: Go 1.22+, mpv or VLC.

**Install mpv:**
- macOS: `brew install mpv`
- Debian/Ubuntu: `sudo apt install mpv`
- Fedora/RHEL: `sudo dnf install mpv`

**Install VLC:**
- macOS: `brew install vlc`
- Debian/Ubuntu: `sudo apt install vlc`
- Fedora/RHEL: `sudo dnf install vlc`

## Use

```bash
volta-iptv
```

| Key | What it does |
|-----|--------------|
| `Tab` | Next tab |
| `/` | Search |
| `↑/↓` | Navigate |
| `Enter` | Play (or drill down on Countries/Languages/Categories) |
| `f` | Toggle favorite |
| `s` | Toggle "streams only" filter |
| `k` | Toggle kids mode (filters adult content) |
| `r` | Refresh data |
| `?` | Help |
| `q` | Quit |

## Tabs

1. **Channels** - Find by name
2. **Countries** - Browse by country, Enter drills down to channels
3. **Languages** - Browse by language
4. **Categories** - Browse by genre (News, Sports, etc.)
5. **Streams** - Direct stream search
6. **Guides** - EPG data (search required - 160K+ entries)
7. **Favorites** - Your saved channels
8. **History** - Recently played

## Data

Everything stored locally:

```
~/.cache/volta-iptv/database.json    # Cached API data (24hr refresh)
~/.config/volta-iptv/favorites.json  # Your favorites
~/.config/volta-iptv/history.json    # Last 50 played
```

## Source

This project is built with:
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [iptv-org API](https://github.com/iptv-org/api) - Stream database

---

BTC: `1NV2myQZNXU1ahPXTyZJnGF7GfdC4SZCN2`
