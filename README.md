# Volta IPTV

A terminal user interface (TUI) for searching and watching free IPTV streams from around the world.

Powered by [iptv-org](https://github.com/iptv-org/iptv) database.

## Features

- **6 Search Categories**: Channels, Countries, Languages, Categories, Streams, Guides
- **Tab Navigation**: Easy switching between search modes
- **Real-time Search**: Filter as you type
- **Favorites System**: Save favorite channels with `★` indicator
- **History Tracking**: Remembers last 50 played streams
- **Dual Player Support**: mpv (primary) with vlc fallback
- **Offline Mode**: Data cached locally for instant startup
- **Auto-refresh**: Updates data every 24 hours in background

## Installation

### Prerequisites

- Go 1.22+
- mpv or vlc (for playback)

### Build from Source

```bash
git clone https://github.com/volta-agent/volta-iptv.git
cd volta-iptv
make build
```

### Install

```bash
make install
```

This installs `volta-iptv` to `~/.local/bin/`. Make sure it's in your PATH.

## Usage

```bash
volta-iptv
```

### Key Bindings

| Key | Action |
|-----|--------|
| `Tab` / `Shift+Tab` | Switch between tabs |
| `/` | Focus search input |
| `↑/↓` or `j/k` | Navigate results |
| `Enter` | Search / Play selected stream |
| `f` | Toggle favorite |
| `r` | Refresh data from API |
| `?` | Toggle help screen |
| `q` / `Ctrl+C` | Quit |

### Search Tabs

1. **Channels** - Search by channel name (e.g., "BBC", "CNN")
2. **Countries** - Browse by country code (e.g., "US", "UK")
3. **Languages** - Browse by language (e.g., "English", "French")
4. **Categories** - Browse by genre (e.g., "News", "Sports")
5. **Streams** - Direct stream search
6. **Guides** - EPG program guide search

## Data Storage

All data is stored in XDG-compliant locations:

```
~/.cache/volta-iptv/
├── database.json # Cached API data

~/.config/volta-iptv/
├── favorites.json # Saved favorites
├── history.json # Playback history
└── config.json # User preferences
```

## API

Volta IPTV uses the [iptv-org API](https://github.com/iptv-org/api):

- `https://iptv-org.github.io/api/channels.json`
- `https://iptv-org.github.io/api/streams.json`
- `https://iptv-org.github.io/api/countries.json`
- `https://iptv-org.github.io/api/languages.json`
- `https://iptv-org.github.io/api/categories.json`
- `https://iptv-org.github.io/api/guides.json`

## Development

### Project Structure

```
volta-iptv/
├── cmd/volta-iptv/    # Entry point
├── internal/
│   ├── api/           # IPTV.org API client
│   ├── models/        # Data models
│   ├── player/        # mpv/vlc integration
│   ├── storage/       # Persistence layer
│   └── tui/           # Bubble Tea TUI
├── Makefile
└── go.mod
```

### Commands

```bash
make build    # Build binary
make run      # Run directly
make install  # Install to ~/.local/bin
make clean    # Clean build artifacts
make test     # Run tests
```

## License

MIT

## Credits

- [iptv-org](https://github.com/iptv-org) for the free IPTV database
- [Charm](https://charm.sh) for the amazing Bubble Tea TUI framework

---

**BTC Donations**: `1NV2myQZNXU1ahPXTyZJnGF7GfdC4SZCN2`

If you find this tool useful, consider supporting development.
