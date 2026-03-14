package recovery

import (
	"os"
	"path/filepath"
)

// Path returns the recovery file path for the given editor filepath.
// For unnamed files it falls back to the user data directory.
func Path(fp string) (string, error) {
	if fp == "" {
		dir, err := dataDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(dir, "untitled.urai-recover"), nil
	}
	return filepath.Join(filepath.Dir(fp), "."+filepath.Base(fp)+".urai-recover"), nil
}

// Write persists content to the recovery file for fp.
func Write(fp, content string) error {
	p, err := Path(fp)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	return os.WriteFile(p, []byte(content), 0o600)
}

// Clean removes the recovery file for fp (call after a successful save).
func Clean(fp string) {
	p, err := Path(fp)
	if err != nil {
		return
	}
	os.Remove(p)
}

// Exists reports whether a recovery file exists for fp and — for named files —
// whether it is newer than the saved file on disk.
func Exists(fp string) bool {
	p, err := Path(fp)
	if err != nil {
		return false
	}
	recStat, err := os.Stat(p)
	if err != nil {
		return false
	}
	if fp == "" {
		return true
	}
	origStat, err := os.Stat(fp)
	if err != nil {
		// Original doesn't exist yet — any recovery is valid.
		return true
	}
	return recStat.ModTime().After(origStat.ModTime())
}

// Read returns the content stored in the recovery file for fp.
func Read(fp string) (string, error) {
	p, err := Path(fp)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func dataDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "prose"), nil
}
