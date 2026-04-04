package player

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

type Player struct {
	preferred string
	path      string
}

func NewPlayer() *Player {
	p := &Player{preferred: "mpv"}
	p.path = p.findPlayer()
	return p
}

func (p *Player) findPlayer() string {
	players := []string{"mpv", "vlc", "cvlc"}
	for _, player := range players {
		if path, err := exec.LookPath(player); err == nil {
			return path
		}
	}
	return ""
}

func (p *Player) SetPreferred(player string) {
	p.preferred = player
	p.path = p.findPlayer()
}

func (p *Player) GetPreferred() string {
	return p.preferred
}

func (p *Player) IsAvailable() bool {
	return p.path != ""
}

func (p *Player) Play(title, url string) error {
	if p.path == "" {
		return fmt.Errorf("no player found (mpv or vlc required)")
	}

	playerName := filepath.Base(p.path)

	// Log what we're trying to play
	logPath := filepath.Join(os.TempDir(), "volta-iptv-player.log")
	if logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		fmt.Fprintf(logFile, "\n=== %s ===\nPlayer: %s (%s)\nURL: %s\n", time.Now().Format("2006-01-02 15:04:05"), playerName, p.path, url)
		logFile.Close()
	}

	var cmd *exec.Cmd
	switch {
	case strings.Contains(playerName, "mpv"):
		cmd = exec.Command(p.path, "--force-window", "--autofit=100%", "--title="+title, "--really-quiet", url)
	case strings.Contains(playerName, "vlc") || strings.Contains(playerName, "cvlc"):
		cmd = exec.Command(p.path, "--play-and-exit", "--no-video-title-show", url)
	default:
		cmd = exec.Command(p.path, url)
	}

	// Detach from terminal - create new process group
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
		Pgid:    0,
	}

	// Don't connect to terminal
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start %s: %w", playerName, err)
	}

	// Log success
	if logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		fmt.Fprintf(logFile, "Started successfully, PID: %d\n", cmd.Process.Pid)
		logFile.Close()
	}

	return nil
}

func (p *Player) PlayWithReferrer(title, url, referrer, userAgent string) error {
	if p.path == "" {
		return fmt.Errorf("no player found (mpv or vlc required)")
	}

	playerName := filepath.Base(p.path)

	var args []string
	switch {
	case strings.Contains(playerName, "mpv"):
		args = []string{"--force-window", "--autofit=100%", "--title=" + title}
		if referrer != "" {
			args = append(args, "--referrer="+referrer)
		}
		if userAgent != "" {
			args = append(args, "--user-agent="+userAgent)
		}
		args = append(args, url)
	case strings.Contains(playerName, "vlc") || strings.Contains(playerName, "cvlc"):
		args = []string{"--play-and-exit", "--no-video-title-show"}
		if referrer != "" {
			args = append(args, "--http-referrer="+referrer)
		}
		if userAgent != "" {
			args = append(args, "--http-user-agent="+userAgent)
		}
		args = append(args, url)
	default:
		args = []string{url}
	}

	cmd := exec.Command(p.path, args...)

	// Detach from terminal
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
		Pgid:    0,
	}

	// Redirect stderr to log file
	logPath := filepath.Join(os.TempDir(), "volta-iptv-player.log")
	if logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		cmd.Stderr = logFile
	}

	cmd.Stdin = nil
	cmd.Stdout = nil

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start %s: %w", playerName, err)
	}

	cmd.Process.Release()
	return nil
}
