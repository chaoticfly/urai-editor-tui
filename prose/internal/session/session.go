package session

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Session holds the persisted editor state between launches.
type Session struct {
	LastFile string `json:"last_file"`
}

// Save persists the current session to disk.
func Save(lastFile string) error {
	p, err := path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(Session{LastFile: lastFile})
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o644)
}

// Load returns the last saved session. Returns an empty Session if none exists.
func Load() (Session, error) {
	p, err := path()
	if err != nil {
		return Session{}, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return Session{}, nil
		}
		return Session{}, err
	}
	var s Session
	if err := json.Unmarshal(data, &s); err != nil {
		return Session{}, err
	}
	return s, nil
}

func path() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "prose", "session.json"), nil
}
