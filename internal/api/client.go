package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/adrg/xdg"
	"github.com/volta-agent/volta-iptv/internal/models"
)

const (
	BaseURL   = "https://iptv-org.github.io/api"
	CacheTTL  = 24 * time.Hour
	UserAgent = "volta-iptv/1.0"
)

var endpoints = map[string]string{
	"channels":   BaseURL + "/channels.json",
	"streams":    BaseURL + "/streams.json",
	"countries":  BaseURL + "/countries.json",
	"languages":  BaseURL + "/languages.json",
	"categories": BaseURL + "/categories.json",
	"guides":     BaseURL + "/guides.json",
}

type Client struct {
	httpClient *http.Client
	cacheDir   string
	data       *models.Database
	dataMu     sync.RWMutex
	lastFetch  time.Time
}

func NewClient() *Client {
	cacheDir := filepath.Join(xdg.CacheHome, "volta-iptv")
	return &Client{
		httpClient: &http.Client{Timeout: 60 * time.Second},
		cacheDir:   cacheDir,
		data:       &models.Database{},
	}
}

func (c *Client) LoadData() error {
	if err := c.loadFromCache(); err == nil && c.isCacheValid() {
		return nil
	}

	if err := c.fetchAll(); err != nil {
		if c.data.Channels != nil {
			return nil
		}
		return fmt.Errorf("failed to fetch data: %w", err)
	}

	return c.saveToCache()
}

func (c *Client) loadFromCache() error {
	cacheFile := filepath.Join(c.cacheDir, "database.json")
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, c.data)
}

func (c *Client) saveToCache() error {
	if err := os.MkdirAll(c.cacheDir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c.data, "", "  ")
	if err != nil {
		return err
	}
	cacheFile := filepath.Join(c.cacheDir, "database.json")
	return os.WriteFile(cacheFile, data, 0644)
}

func (c *Client) isCacheValid() bool {
	cacheFile := filepath.Join(c.cacheDir, "database.json")
	info, err := os.Stat(cacheFile)
	if err != nil {
		return false
	}
	return time.Since(info.ModTime()) < CacheTTL
}

func (c *Client) fetchAll() error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(endpoints))

	c.dataMu.Lock()
	defer c.dataMu.Unlock()

	for name, url := range endpoints {
		wg.Add(1)
		go func(name, url string) {
			defer wg.Done()
			if err := c.fetchEndpoint(name, url); err != nil {
				errChan <- fmt.Errorf("%s: %w", name, err)
			}
		}(name, url)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	c.lastFetch = time.Now()
	return nil
}

func (c *Client) fetchEndpoint(name, url string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", UserAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	switch name {
	case "channels":
		return json.Unmarshal(data, &c.data.Channels)
	case "streams":
		return json.Unmarshal(data, &c.data.Streams)
	case "countries":
		return json.Unmarshal(data, &c.data.Countries)
	case "languages":
		return json.Unmarshal(data, &c.data.Languages)
	case "categories":
		return json.Unmarshal(data, &c.data.Categories)
	case "guides":
		return json.Unmarshal(data, &c.data.Guides)
	}
	return nil
}

func (c *Client) Refresh() error {
	return c.fetchAll()
}

func (c *Client) GetDatabase() *models.Database {
	c.dataMu.RLock()
	defer c.dataMu.RUnlock()
	return c.data
}

func (c *Client) GetChannels() []models.Channel {
	c.dataMu.RLock()
	defer c.dataMu.RUnlock()
	return c.data.Channels
}

func (c *Client) GetStreams() []models.Stream {
	c.dataMu.RLock()
	defer c.dataMu.RUnlock()
	return c.data.Streams
}

func (c *Client) GetCountries() []models.Country {
	c.dataMu.RLock()
	defer c.dataMu.RUnlock()
	return c.data.Countries
}

func (c *Client) GetLanguages() []models.Language {
	c.dataMu.RLock()
	defer c.dataMu.RUnlock()
	return c.data.Languages
}

func (c *Client) GetCategories() []models.Category {
	c.dataMu.RLock()
	defer c.dataMu.RUnlock()
	return c.data.Categories
}

func (c *Client) GetGuides() []models.Guide {
	c.dataMu.RLock()
	defer c.dataMu.RUnlock()
	return c.data.Guides
}

func (c *Client) GetChannelByID(id string) *models.Channel {
	c.dataMu.RLock()
	defer c.dataMu.RUnlock()
	for _, ch := range c.data.Channels {
		if ch.ID == id {
			return &ch
		}
	}
	return nil
}

func (c *Client) GetStreamsForChannel(channelID string) []models.Stream {
	c.dataMu.RLock()
	defer c.dataMu.RUnlock()
	var streams []models.Stream
	for _, s := range c.data.Streams {
		if s.Channel == channelID {
			streams = append(streams, s)
		}
	}
	return streams
}

func (c *Client) FindCountryByCode(code string) *models.Country {
	c.dataMu.RLock()
	defer c.dataMu.RUnlock()
	for _, c := range c.data.Countries {
		if c.Code == code {
			return &c
		}
	}
	return nil
}

func (c *Client) FindLanguageByCode(code string) *models.Language {
	c.dataMu.RLock()
	defer c.dataMu.RUnlock()
	for _, l := range c.data.Languages {
		if l.Code == code {
			return &l
		}
	}
	return nil
}

func (c *Client) FindCategoryByID(id string) *models.Category {
	c.dataMu.RLock()
	defer c.dataMu.RUnlock()
	for _, cat := range c.data.Categories {
		if cat.ID == id {
			return &cat
		}
	}
	return nil
}
