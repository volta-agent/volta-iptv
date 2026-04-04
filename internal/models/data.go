package models

type Channel struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	AltNames   []string `json:"alt_names"`
	Network    string   `json:"network,omitempty"`
	Owners     []string `json:"owners,omitempty"`
	Country    string   `json:"country"`
	Categories []string `json:"categories,omitempty"`
	IsNSFW     bool     `json:"is_nsfw"`
	Launched   string   `json:"launched,omitempty"`
	Closed     string   `json:"closed,omitempty"`
	ReplacedBy string   `json:"replaced_by,omitempty"`
	Website    string   `json:"website,omitempty"`
}

type Stream struct {
	Channel   string `json:"channel"`
	Feed      string `json:"feed,omitempty"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	Referrer  string `json:"referrer,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
	Quality   string `json:"quality,omitempty"`
}

type Country struct {
	Name      string   `json:"name"`
	Code      string   `json:"code"`
	Languages []string `json:"languages"`
	Flag      string   `json:"flag"`
}

type Language struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

type Category struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Guide struct {
	Channel  string `json:"channel"`
	Feed     string `json:"feed,omitempty"`
	Site     string `json:"site"`
	SiteID   string `json:"site_id"`
	SiteName string `json:"site_name"`
	Lang     string `json:"lang"`
}

type Database struct {
	Channels   []Channel  `json:"channels"`
	Streams    []Stream   `json:"streams"`
	Countries  []Country  `json:"countries"`
	Languages  []Language `json:"languages"`
	Categories []Category `json:"categories"`
	Guides     []Guide    `json:"guides"`
}

type Favorite struct {
	ChannelID string `json:"channel_id"`
	AddedAt   int64  `json:"added_at"`
}

type HistoryEntry struct {
	ChannelID string `json:"channel_id"`
	Name      string `json:"name"`
	URL       string `json:"url"`
	PlayedAt  int64  `json:"played_at"`
	PlayCount int    `json:"play_count"`
}

type Config struct {
	Player      string `json:"player"`
	Quality     string `json:"quality_filter,omitempty"`
	Country     string `json:"country_filter,omitempty"`
	AutoRefresh bool   `json:"auto_refresh"`
	CacheTTL    int    `json:"cache_ttl_hours"`
}
