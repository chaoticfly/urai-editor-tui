package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"urai/internal/app"
	"urai/internal/config"
	"urai/internal/recovery"
	"urai/internal/session"
)

const version = "1.0.7"

func main() {
	var (
		configPath  string
		showVersion bool
	)

	flag.StringVar(&configPath, "config", "", "path to a custom config file")
	flag.BoolVar(&showVersion, "version", false, "print version and exit")
	flag.Parse()

	if showVersion {
		fmt.Println("urai", version)
		os.Exit(0)
	}

	// Load config — fall back to defaults on parse error so a corrupted
	// config file never prevents the editor from opening.
	var cfg *config.Config
	var cfgErr error
	if configPath != "" {
		cfg, cfgErr = config.LoadFrom(configPath)
	} else {
		cfg, cfgErr = config.Load()
	}
	if cfgErr != nil {
		fmt.Fprintf(os.Stderr, "warning: config error (%v) — using defaults\n", cfgErr)
		cfg = config.DefaultConfig()
	}

	fp := ""
	if flag.NArg() > 0 {
		fp = flag.Arg(0)
	} else {
		// No file given — restore the last session if the file still exists.
		if s, err := session.Load(); err == nil && s.LastFile != "" {
			if _, err := os.Stat(s.LastFile); err == nil {
				fp = s.LastFile
			}
		}
	}

	m := app.New(fp, cfg)

	// Panic recovery: write a last-resort recovery file before crashing so
	// no work is lost on an unhandled panic.
	defer func() {
		if r := recover(); r != nil {
			recovery.Write(m.Filepath(), m.Content()) //nolint:errcheck
			fmt.Fprintf(os.Stderr, "urai crashed: %v\nRecovery file written — reopen the file to restore.\n", r)
			os.Exit(2)
		}
	}()

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
