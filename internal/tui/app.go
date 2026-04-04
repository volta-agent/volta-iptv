package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/volta-agent/volta-iptv/internal/api"
	"github.com/volta-agent/volta-iptv/internal/models"
	"github.com/volta-agent/volta-iptv/internal/player"
	"github.com/volta-agent/volta-iptv/internal/storage"
)

const (
	tabChannels = iota
	tabCountries
	tabLanguages
	tabCategories
	tabGuides
	tabFavorites
	tabHistory
	numTabs = 7
)

type ResultItem struct {
	ID         string
	Title      string
	Subtitle   string
	Info       string
	IsFavorite bool
	Data       interface{}
}

type Model struct {
	api *api.Client
	player *player.Player
	storage *storage.Storage

	currentTab int
	searchInput textinput.Model
	results []ResultItem
	selectedIndex int
	searchQuery string
	loading bool
	err error
	statusMessage string
	showHelp bool
	onlyWithStreams bool
	kidsMode bool

	// Stream selection mode
	showStreamPicker bool
	pickerChannel    *models.Channel
	pickerStreams    []models.Stream
	pickerIndex      int

	width int
	height int
}

func NewModel() Model {
	apiClient := api.NewClient()
	pl := player.NewPlayer()
	st := storage.NewStorage()

	ti := textinput.New()
	ti.Placeholder = "Search..."
	ti.CharLimit = 100

	return Model{
		api:         apiClient,
		player:      pl,
		storage:     st,
		searchInput: ti,
		results:     []ResultItem{},
		currentTab:  tabChannels,
		loading:     true,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		loadDataCmd(m.api),
		textinput.Blink,
	)
}

type loadDataMsg struct {
	err error
}

type refreshMsg struct {
	err error
}

func loadDataCmd(apiClient *api.Client) tea.Cmd {
	return func() tea.Msg {
		err := apiClient.LoadData()
		return loadDataMsg{err: err}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Handle stream picker mode separately
	if m.showStreamPicker {
		return m.updatePicker(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEscape:
			if m.searchInput.Focused() {
				m.searchInput.Blur()
				return m, nil
			}
			return m, tea.Quit

		case tea.KeyTab:
			m.currentTab = (m.currentTab + 1) % numTabs
			m.results = m.filterResults(m.searchQuery)
			m.selectedIndex = 0
			m.statusMessage = ""
			return m, nil

		case tea.KeyShiftTab:
			m.currentTab = (m.currentTab - 1 + numTabs) % numTabs
			m.results = m.filterResults(m.searchQuery)
			m.selectedIndex = 0
			m.statusMessage = ""
			return m, nil

		case tea.KeyUp, tea.KeyCtrlP:
			if m.searchInput.Focused() {
				m.searchInput.Blur()
			} else if m.selectedIndex > 0 {
				m.selectedIndex--
				m.statusMessage = ""
			}

		case tea.KeyDown, tea.KeyCtrlN:
			if !m.searchInput.Focused() {
				if m.selectedIndex < len(m.results)-1 {
					m.selectedIndex++
					m.statusMessage = ""
				}
			}

		case tea.KeyRunes:
			if string(msg.Runes) == "/" && !m.searchInput.Focused() {
				m.searchInput.Focus()
				m.showHelp = false
				return m, nil
			}
			if string(msg.Runes) == "r" && !m.searchInput.Focused() && !m.showHelp {
				m.loading = true
				m.statusMessage = "Refreshing data..."
				return m, func() tea.Msg {
					err := m.api.Refresh()
					return refreshMsg{err: err}
				}
			}
			if string(msg.Runes) == "?" && !m.searchInput.Focused() {
				m.showHelp = !m.showHelp
				m.statusMessage = ""
				return m, nil
			}
			if string(msg.Runes) == "f" && !m.searchInput.Focused() && !m.showHelp {
				return m.handleFavorite()
			}
		if string(msg.Runes) == "s" && !m.searchInput.Focused() && !m.showHelp {
			m.onlyWithStreams = !m.onlyWithStreams
			m.results = m.filterResults(m.searchQuery)
			m.selectedIndex = 0
			if m.onlyWithStreams {
				m.statusMessage = "Filter: Only channels with streams"
			} else {
				m.statusMessage = "Filter: All channels"
			}
			return m, nil
		}
		if string(msg.Runes) == "k" && !m.searchInput.Focused() && !m.showHelp {
			m.kidsMode = !m.kidsMode
			m.results = m.filterResults(m.searchQuery)
			m.selectedIndex = 0
			if m.kidsMode {
				m.statusMessage = "Kids mode: ON (adult content filtered)"
			} else {
				m.statusMessage = "Kids mode: OFF"
			}
			return m, nil
		}
		if string(msg.Runes) == "q" && !m.searchInput.Focused() {
				if m.showHelp {
					m.showHelp = false
					return m, nil
				}
				return m, tea.Quit
			}

		case tea.KeyEnter:
			if m.searchInput.Focused() {
				m.searchInput.Blur()
				m.searchQuery = m.searchInput.Value()
				m.results = m.filterResults(m.searchQuery)
				m.selectedIndex = 0
				return m, nil
			}
			return m.handlePlay()

		case tea.KeyBackspace:
			if !m.searchInput.Focused() && m.searchQuery != "" {
				m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
				m.searchInput.SetValue(m.searchQuery)
				m.results = m.filterResults(m.searchQuery)
				m.selectedIndex = 0
				return m, nil
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case loadDataMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			m.statusMessage = fmt.Sprintf("Error loading data: %v", msg.err)
		} else {
			m.statusMessage = fmt.Sprintf("Loaded %d channels, %d streams",
				len(m.api.GetChannels()), len(m.api.GetStreams()))
		}

	case refreshMsg:
		m.loading = false
		if msg.err != nil {
			m.statusMessage = fmt.Sprintf("Refresh failed: %v", msg.err)
		} else {
			m.statusMessage = "Data refreshed successfully"
		}
	}

	var tiCmd tea.Cmd
	m.searchInput, tiCmd = m.searchInput.Update(msg)
	if m.searchInput.Focused() {
		newQuery := m.searchInput.Value()
		if newQuery != m.searchQuery {
			m.statusMessage = ""
		}
		m.searchQuery = newQuery
		m.results = m.filterResults(m.searchQuery)
	}
	cmds = append(cmds, tiCmd)

	return m, tea.Batch(cmds...)
}

// updatePicker handles input for stream quality picker
func (m *Model) updatePicker(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEscape, tea.KeyCtrlC:
			m.showStreamPicker = false
			return m, nil
		case tea.KeyUp, tea.KeyCtrlP:
			if m.pickerIndex > 0 {
				m.pickerIndex--
			}
		case tea.KeyDown, tea.KeyCtrlN:
			if m.pickerIndex < len(m.pickerStreams)-1 {
				m.pickerIndex++
			}
		case tea.KeyEnter:
			// Play selected stream
			stream := m.pickerStreams[m.pickerIndex]
			m.storage.AddToHistory(m.pickerChannel.ID, m.pickerChannel.Name, stream.URL)
			m.statusMessage = fmt.Sprintf("Launching: %s (%s)", m.pickerChannel.Name, stream.Quality)
			if err := m.player.Play(m.pickerChannel.Name, stream.URL); err != nil {
				m.statusMessage = fmt.Sprintf("Error: %v", err)
			}
			m.showStreamPicker = false
			return m, nil
		}
	}
	return m, nil
}

func (m *Model) handleFavorite() (tea.Model, tea.Cmd) {
	if m.selectedIndex >= len(m.results) {
		return m, nil
	}

	result := m.results[m.selectedIndex]
	if channel, ok := result.Data.(models.Channel); ok {
		if m.storage.IsFavorite(channel.ID) {
			m.storage.RemoveFavorite(channel.ID)
			m.statusMessage = fmt.Sprintf("Removed %s from favorites", channel.Name)
		} else {
			m.storage.AddFavorite(channel.ID)
			m.statusMessage = fmt.Sprintf("Added %s to favorites", channel.Name)
		}
		m.results = m.filterResults(m.searchQuery)
	}
	return m, nil
}

func (m *Model) handlePlay() (tea.Model, tea.Cmd) {
	if m.selectedIndex >= len(m.results) {
		return m, nil
	}

	result := m.results[m.selectedIndex]

	// Handle filterable items (Country, Language, Category) - drill down instead of play
	switch data := result.Data.(type) {
	case models.Country:
		// Filter channels by this country
		m.searchQuery = ""
		m.searchInput.SetValue("")
		m.currentTab = tabChannels
		m.results = m.filterByCountry(data.Code)
		m.selectedIndex = 0
		m.statusMessage = fmt.Sprintf("Showing channels from %s", data.Name)
		return m, nil
	case models.Language:
		// Filter channels by this language
		m.searchQuery = ""
		m.searchInput.SetValue("")
		m.currentTab = tabChannels
		m.results = m.filterByLanguage(data.Code)
		m.selectedIndex = 0
		m.statusMessage = fmt.Sprintf("Showing channels in %s", data.Name)
		return m, nil
	case models.Category:
		// Filter channels by this category
		m.searchQuery = ""
		m.searchInput.SetValue("")
		m.currentTab = tabChannels
		m.results = m.filterByCategory(data.ID)
		m.selectedIndex = 0
		m.statusMessage = fmt.Sprintf("Showing %s channels", data.Name)
		return m, nil
	}

	// Handle playable items (Channel, Stream)
	if !m.player.IsAvailable() {
		m.statusMessage = "Error: No player found. Install mpv or vlc."
		return m, nil
	}

	var channel *models.Channel
	var streams []models.Stream
	var streamURL string

	switch data := result.Data.(type) {
	case models.Channel:
		channel = &data
		streams = m.api.GetStreamsForChannel(channel.ID)
		// If multiple streams, show picker
		if len(streams) > 1 {
			m.showStreamPicker = true
			m.pickerChannel = channel
			m.pickerStreams = streams
			m.pickerIndex = 0
			return m, nil
		}
		if len(streams) > 0 {
			streamURL = streams[0].URL
		}
	case models.Stream:
		streamURL = data.URL
		channel = m.api.GetChannelByID(data.Channel)
	case models.HistoryEntry:
		// Play directly from history
		streamURL = data.URL
		channel = m.api.GetChannelByID(data.ChannelID)
	}

	if streamURL == "" {
		m.statusMessage = "No stream available for this channel"
		return m, nil
	}

	if channel != nil {
		m.storage.AddToHistory(channel.ID, channel.Name, streamURL)
		m.statusMessage = fmt.Sprintf("Launching: %s", channel.Name)
	} else {
		m.storage.AddToHistory("", result.Title, streamURL)
		m.statusMessage = fmt.Sprintf("Launching: %s", result.Title)
	}

	if err := m.player.Play(result.Title, streamURL); err != nil {
		m.statusMessage = fmt.Sprintf("Error: %v", err)
	}

	return m, nil
}

func (m Model) filterResults(query string) []ResultItem {
	query = strings.ToLower(query)
	var results []ResultItem

	switch m.currentTab {
	case tabChannels:
		for _, ch := range m.api.GetChannels() {
			// Skip channels without streams if filter is enabled
			if m.onlyWithStreams && !m.hasStream(ch.ID) {
				continue
			}
			// Skip adult content in kids mode
			if m.kidsMode && m.isAdultContent(ch) {
				continue
			}
			// Search by channel name, ID, country name, or category
			if m.matchesChannel(ch, query) {
				country := m.api.FindCountryByCode(ch.Country)
				countryName := ch.Country
				if country != nil {
					countryName = country.Flag + " " + country.Name
				}
				results = append(results, ResultItem{
					ID:         ch.ID,
					Title:      ch.Name,
					Subtitle:   countryName,
					Info:       strings.Join(ch.Categories, ", "),
					IsFavorite: m.storage.IsFavorite(ch.ID),
					Data:       ch,
				})
			}
		}

	case tabCountries:
		for _, c := range m.api.GetCountries() {
			if query == "" ||
				strings.Contains(strings.ToLower(c.Name), query) ||
				strings.Contains(strings.ToLower(c.Code), query) {
				count := 0
				for _, ch := range m.api.GetChannels() {
					if ch.Country == c.Code {
						count++
					}
				}
				results = append(results, ResultItem{
					ID:       c.Code,
					Title:    fmt.Sprintf("%s %s", c.Flag, c.Name),
					Subtitle: c.Code,
					Info:     fmt.Sprintf("%d channels", count),
					Data:     c,
				})
			}
		}

	case tabLanguages:
		for _, l := range m.api.GetLanguages() {
			if query == "" ||
				strings.Contains(strings.ToLower(l.Name), query) ||
				strings.Contains(strings.ToLower(l.Code), query) {
				results = append(results, ResultItem{
					ID:       l.Code,
					Title:    l.Name,
					Subtitle: l.Code,
					Info:     "",
					Data:     l,
				})
			}
		}

	case tabCategories:
		for _, cat := range m.api.GetCategories() {
			if query == "" ||
				strings.Contains(strings.ToLower(cat.Name), query) {
				count := 0
				for _, ch := range m.api.GetChannels() {
					for _, c := range ch.Categories {
						if c == cat.ID {
							count++
							break
						}
					}
				}
				results = append(results, ResultItem{
					ID:       cat.ID,
					Title:    cat.Name,
					Subtitle: cat.ID,
					Info:     fmt.Sprintf("%d channels", count),
					Data:     cat,
				})
			}
		}

	case tabGuides:
		// Only load guides when searching - too many to display all
		if query != "" {
			for _, g := range m.api.GetGuides() {
				if strings.Contains(strings.ToLower(g.SiteName), query) {
					channel := m.api.GetChannelByID(g.Channel)
					channelName := g.Channel
					if channel != nil {
						channelName = channel.Name
					}
					results = append(results, ResultItem{
						ID:        g.Channel,
						Title:     g.SiteName,
						Subtitle:  channelName,
						Info:      g.Site,
						Data:      g,
					})
				}
			}
		}

	case tabFavorites:
		for _, fav := range m.storage.GetFavorites() {
			channel := m.api.GetChannelByID(fav.ChannelID)
			if channel == nil {
				continue
			}
			if query == "" ||
				strings.Contains(strings.ToLower(channel.Name), query) ||
				strings.Contains(strings.ToLower(channel.ID), query) {
				country := m.api.FindCountryByCode(channel.Country)
				countryName := channel.Country
				if country != nil {
					countryName = country.Flag + " " + country.Name
				}
				results = append(results, ResultItem{
					ID:         channel.ID,
					Title:      channel.Name,
					Subtitle:   countryName,
					Info:       strings.Join(channel.Categories, ", "),
					IsFavorite: true,
					Data:       *channel,
				})
			}
		}

	case tabHistory:
		for _, h := range m.storage.GetHistory() {
			if query == "" ||
				strings.Contains(strings.ToLower(h.Name), query) {
				results = append(results, ResultItem{
					ID:       h.ChannelID,
					Title:    h.Name,
					Subtitle: fmt.Sprintf("Played %d times", h.PlayCount),
					Info:     h.URL,
					Data:     h,
				})
			}
		}
	}

	return results
}

func (m Model) filterByCountry(countryCode string) []ResultItem {
	var results []ResultItem
	for _, ch := range m.api.GetChannels() {
		if ch.Country == countryCode {
			// Skip channels without streams if filter is enabled
			if m.onlyWithStreams && !m.hasStream(ch.ID) {
				continue
			}
			country := m.api.FindCountryByCode(ch.Country)
			countryName := ch.Country
			if country != nil {
				countryName = country.Flag + " " + country.Name
			}
			results = append(results, ResultItem{
				ID:         ch.ID,
				Title:      ch.Name,
				Subtitle:   countryName,
				Info:       strings.Join(ch.Categories, ", "),
				IsFavorite: m.storage.IsFavorite(ch.ID),
				Data:       ch,
			})
		}
	}
	return results
}

func (m Model) filterByLanguage(langCode string) []ResultItem {
	var results []ResultItem
	for _, c := range m.api.GetCountries() {
		for _, lang := range c.Languages {
			if lang == langCode {
				for _, ch := range m.api.GetChannels() {
					if ch.Country == c.Code {
						// Skip channels without streams if filter is enabled
						if m.onlyWithStreams && !m.hasStream(ch.ID) {
							continue
						}
						country := m.api.FindCountryByCode(ch.Country)
						countryName := ch.Country
						if country != nil {
							countryName = country.Flag + " " + country.Name
						}
						results = append(results, ResultItem{
							ID:         ch.ID,
							Title:      ch.Name,
							Subtitle:   countryName,
							Info:       strings.Join(ch.Categories, ", "),
							IsFavorite: m.storage.IsFavorite(ch.ID),
							Data:       ch,
						})
					}
				}
			}
		}
	}
	return results
}

func (m Model) filterByCategory(categoryID string) []ResultItem {
	var results []ResultItem
	for _, ch := range m.api.GetChannels() {
		for _, cat := range ch.Categories {
			if cat == categoryID {
				// Skip channels without streams if filter is enabled
				if m.onlyWithStreams && !m.hasStream(ch.ID) {
					break
				}
				country := m.api.FindCountryByCode(ch.Country)
				countryName := ch.Country
				if country != nil {
					countryName = country.Flag + " " + country.Name
				}
				results = append(results, ResultItem{
					ID:         ch.ID,
					Title:      ch.Name,
					Subtitle:   countryName,
					Info:       strings.Join(ch.Categories, ", "),
					IsFavorite: m.storage.IsFavorite(ch.ID),
					Data:       ch,
				})
				break
			}
		}
	}
	return results
}

func (m Model) hasStream(channelID string) bool {
	streams := m.api.GetStreamsForChannel(channelID)
	return len(streams) > 0
}

// matchesChannel checks if a channel matches the search query
// Searches by channel name, ID, country name, language, or category
func (m Model) matchesChannel(ch models.Channel, query string) bool {
	if query == "" {
		return true
	}

	// Check channel name and ID
	if strings.Contains(strings.ToLower(ch.Name), query) ||
		strings.Contains(strings.ToLower(ch.ID), query) {
		return true
	}

	// Check country name
	if country := m.api.FindCountryByCode(ch.Country); country != nil {
		if strings.Contains(strings.ToLower(country.Name), query) {
			return true
		}
	}

	// Check categories
	for _, catID := range ch.Categories {
		if cat := m.api.FindCategoryByID(catID); cat != nil {
			if strings.Contains(strings.ToLower(cat.Name), query) {
				return true
			}
		}
	}

	// Check languages (from country)
	if country := m.api.FindCountryByCode(ch.Country); country != nil {
		for _, langCode := range country.Languages {
			if lang := m.api.FindLanguageByCode(langCode); lang != nil {
				if strings.Contains(strings.ToLower(lang.Name), query) {
					return true
				}
			}
		}
	}

	return false
}

// isAdultContent checks if a channel has adult categories
func (m Model) isAdultContent(ch models.Channel) bool {
	for _, cat := range ch.Categories {
		if cat == "xxx" || cat == "adult" {
			return true
		}
	}
	// Also check channel name for common adult keywords
	name := strings.ToLower(ch.Name)
	adultKeywords := []string{"xxx", "porn", "adult", "sex", "18+", "erotic"}
	for _, kw := range adultKeywords {
		if strings.Contains(name, kw) {
			return true
		}
	}
	return false
}

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Show stream picker overlay
	if m.showStreamPicker {
		return m.renderStreamPicker()
	}

	if m.showHelp {
		return m.renderHelp()
	}

	var sections []string

	title := TitleStyle.Render("📺 volta-iptv - Free IPTV Stream Finder")
	sections = append(sections, title)

	tabs := renderTabs(m.currentTab)
	sections = append(sections, tabs)

	if !m.player.IsAvailable() {
		sections = append(sections, ErrorStyle.Render("⚠ No player found! Install mpv or vlc to play streams"))
	}

	searchBox := SearchBoxStyle.Render("Search: " + m.searchInput.View())
	sections = append(sections, searchBox)

	if m.loading {
		sections = append(sections, HelpStyle.Render("Loading data from iptv-org..."))
	} else if len(m.results) == 0 && m.searchQuery == "" {
		if m.currentTab == tabGuides {
			sections = append(sections, HelpStyle.Render("160K+ guides - type to search (required)"))
		} else {
			sections = append(sections, HelpStyle.Render("Type to search or press '/' to focus search"))
		}
	} else if len(m.results) == 0 {
		sections = append(sections, HelpStyle.Render("No results found"))
	} else {
		visibleCount := min(10, len(m.results))
		startIdx := max(0, min(m.selectedIndex-5, len(m.results)-visibleCount))
		endIdx := startIdx + visibleCount
		if endIdx > len(m.results) {
			endIdx = len(m.results)
		}

		var items []string
		for i := startIdx; i < endIdx; i++ {
			r := m.results[i]
			style := ListItemStyle
			if i == m.selectedIndex {
				style = SelectedListItemStyle
			}

			fav := " "
			if r.IsFavorite {
				fav = "★"
			}

			line := fmt.Sprintf("%s %s", fav, r.Title)
			if r.Subtitle != "" {
				line = fmt.Sprintf("%s %s (%s)", fav, r.Title, truncate(r.Subtitle, 20))
			}
			if r.Info != "" {
				line = fmt.Sprintf("%s [%s]", line, truncate(r.Info, 15))
			}
			items = append(items, style.Render(line))
		}
		sections = append(sections, strings.Join(items, "\n"))

		pageInfo := fmt.Sprintf("%d/%d results", m.selectedIndex+1, len(m.results))
		sections = append(sections, HelpStyle.Render(pageInfo))
	}

	footer := HelpStyle.Render("?: Help | Tab: Switch | /: Search | Enter: Play | f: Fav | s: Streams | k: Kids | r: Refresh | q: Quit")
	sections = append(sections, footer)

	if m.statusMessage != "" {
		sections = append(sections, StatusStyle.Render(m.statusMessage))
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m Model) renderHelp() string {
	helpTitle := TitleStyle.Render("📖 Help - volta-iptv")

	helpContent := `
Navigation:
 Tab / Shift+Tab Switch between search categories
 ↑/↓ or j/k Navigate results
 Enter Play selected stream / Drill down

Search:
 / Focus search input
 Type Search as you type
 Backspace Delete last character
 Esc Blur search input

Actions:
 f Toggle favorite (★)
 s Toggle stream filter (show only channels with streams)
 k Toggle kids mode (filter adult content)
 r Refresh data from API

General:
 ? Toggle this help screen
 q or Ctrl+C Quit

Tabs:
 1. Channels Search by channel name
 2. Countries Browse by country (Enter to filter)
 3. Languages Browse by language (Enter to filter)
 4. Categories Browse by genre (Enter to filter)
 5. Guides EPG program guides
 6. Favorites Your saved favorite channels
 7. History Recently played streams

Data:
 Channels, streams, and metadata are cached locally
 at ~/.cache/volta-iptv/ and refreshed every 24 hours.

 Favorites: ~/.config/volta-iptv/favorites.json
 History: ~/.config/volta-iptv/history.json

Player:
  Requires mpv or vlc installed.
  mpv is preferred for better streaming support.
`

	helpBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(primaryColor).
		Padding(1, 2).
		Render(helpContent)

	pressQ := HelpStyle.Render("Press ? or q to close this help screen")

	return lipgloss.JoinVertical(lipgloss.Left, helpTitle, helpBox, pressQ)
}

// renderStreamPicker shows quality selection for channels with multiple streams
func (m Model) renderStreamPicker() string {
	title := TitleStyle.Render(fmt.Sprintf("📺 Select Stream Quality - %s", m.pickerChannel.Name))

	var items []string
	for i, stream := range m.pickerStreams {
		style := ListItemStyle
		if i == m.pickerIndex {
			style = SelectedListItemStyle
		}
		quality := stream.Quality
		if quality == "" {
			quality = "unknown"
		}
		line := fmt.Sprintf("  %s", quality)
		items = append(items, style.Render(line))
	}

	footer := HelpStyle.Render("↑/↓: Navigate | Enter: Play | Esc: Cancel")

	content := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(accentColor).
		Padding(1, 2).
		Render(strings.Join(items, "\n"))

	return lipgloss.JoinVertical(lipgloss.Left, title, content, footer)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
