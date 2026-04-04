package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/adrg/xdg"
	"github.com/volta-agent/volta-iptv/internal/models"
)

type Storage struct {
	configDir string
	favorites []models.Favorite
	history   []models.HistoryEntry
	config    models.Config
	favMu     sync.RWMutex
	histMu    sync.RWMutex
}

func NewStorage() *Storage {
	configDir := filepath.Join(xdg.ConfigHome, "volta-iptv")
	s := &Storage{
		configDir: configDir,
		config: models.Config{
			Player:      "mpv",
			AutoRefresh: true,
			CacheTTL:    24,
		},
	}
	s.loadFavorites()
	s.loadHistory()
	s.loadConfig()
	return s
}

func (s *Storage) loadFavorites() {
	file := filepath.Join(s.configDir, "favorites.json")
	data, err := os.ReadFile(file)
	if err != nil {
		s.favorites = []models.Favorite{}
		return
	}
	json.Unmarshal(data, &s.favorites)
}

func (s *Storage) saveFavorites() error {
	if err := os.MkdirAll(s.configDir, 0755); err != nil {
		return err
	}
	data, _ := json.MarshalIndent(s.favorites, "", "  ")
	return os.WriteFile(filepath.Join(s.configDir, "favorites.json"), data, 0644)
}

func (s *Storage) AddFavorite(channelID string) error {
	s.favMu.Lock()
	defer s.favMu.Unlock()

	for _, f := range s.favorites {
		if f.ChannelID == channelID {
			return nil
		}
	}

	s.favorites = append(s.favorites, models.Favorite{
		ChannelID: channelID,
		AddedAt:   currentTime(),
	})
	return s.saveFavorites()
}

func (s *Storage) RemoveFavorite(channelID string) error {
	s.favMu.Lock()
	defer s.favMu.Unlock()

	for i, f := range s.favorites {
		if f.ChannelID == channelID {
			s.favorites = append(s.favorites[:i], s.favorites[i+1:]...)
			break
		}
	}
	return s.saveFavorites()
}

func (s *Storage) IsFavorite(channelID string) bool {
	s.favMu.RLock()
	defer s.favMu.RUnlock()
	for _, f := range s.favorites {
		if f.ChannelID == channelID {
			return true
		}
	}
	return false
}

func (s *Storage) GetFavorites() []models.Favorite {
	s.favMu.RLock()
	defer s.favMu.RUnlock()
	return s.favorites
}

func (s *Storage) loadHistory() {
	file := filepath.Join(s.configDir, "history.json")
	data, err := os.ReadFile(file)
	if err != nil {
		s.history = []models.HistoryEntry{}
		return
	}
	json.Unmarshal(data, &s.history)
}

func (s *Storage) saveHistory() error {
	if err := os.MkdirAll(s.configDir, 0755); err != nil {
		return err
	}
	data, _ := json.MarshalIndent(s.history, "", "  ")
	return os.WriteFile(filepath.Join(s.configDir, "history.json"), data, 0644)
}

func (s *Storage) AddToHistory(channelID, name, url string) error {
	s.histMu.Lock()
	defer s.histMu.Unlock()

	now := currentTime()
	found := false
	for i, h := range s.history {
		if h.ChannelID == channelID {
			s.history[i].PlayedAt = now
			s.history[i].PlayCount++
			found = true
			break
		}
	}

	if !found {
		s.history = append([]models.HistoryEntry{{
			ChannelID: channelID,
			Name:      name,
			URL:       url,
			PlayedAt:  now,
			PlayCount: 1,
		}}, s.history...)
	}

	if len(s.history) > 50 {
		s.history = s.history[:50]
	}
	return s.saveHistory()
}

func (s *Storage) GetHistory() []models.HistoryEntry {
	s.histMu.RLock()
	defer s.histMu.RUnlock()
	return s.history
}

func (s *Storage) loadConfig() {
	file := filepath.Join(s.configDir, "config.json")
	data, err := os.ReadFile(file)
	if err != nil {
		return
	}
	json.Unmarshal(data, &s.config)
}

func (s *Storage) saveConfig() error {
	if err := os.MkdirAll(s.configDir, 0755); err != nil {
		return err
	}
	data, _ := json.MarshalIndent(s.config, "", "  ")
	return os.WriteFile(filepath.Join(s.configDir, "config.json"), data, 0644)
}

func (s *Storage) GetConfig() models.Config {
	return s.config
}

func (s *Storage) SetPlayer(player string) error {
	s.config.Player = player
	return s.saveConfig()
}

func (s *Storage) SetQualityFilter(quality string) error {
	s.config.Quality = quality
	return s.saveConfig()
}

func (s *Storage) SetCountryFilter(country string) error {
	s.config.Country = country
	return s.saveConfig()
}

func currentTime() int64 {
	return time.Now().Unix()
}
