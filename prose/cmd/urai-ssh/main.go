package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"net"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"

	gossh "golang.org/x/crypto/ssh"

	"urai/internal/app"
	"urai/internal/config"
	"urai/internal/recovery"
)

const version = "1.0.8"

func main() {
	var (
		addr        string
		hostKeyPath string
		configPath  string
		showVersion bool
	)
	flag.StringVar(&addr, "addr", "0.0.0.0:2222", "listen address (host:port)")
	flag.StringVar(&hostKeyPath, "hostkey", "", "path to SSH host private key (default: ~/.config/urai/ssh_host_key)")
	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.BoolVar(&showVersion, "version", false, "print version and exit")
	flag.Parse()

	if showVersion {
		fmt.Println("urai-ssh", version)
		os.Exit(0)
	}

	if hostKeyPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("cannot determine home directory: %v", err)
		}
		hostKeyPath = filepath.Join(home, ".config", "urai", "ssh_host_key")
	}

	if err := ensureHostKey(hostKeyPath); err != nil {
		log.Fatalf("host key: %v", err)
	}

	var cfg *config.Config
	if configPath != "" {
		var err error
		cfg, err = config.LoadFrom(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: config error (%v) — using defaults\n", err)
			cfg = config.DefaultConfig()
		}
	} else {
		var err error
		cfg, err = config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: config error (%v) — using defaults\n", err)
			cfg = config.DefaultConfig()
		}
	}

	s, err := wish.NewServer(
		wish.WithAddress(addr),
		wish.WithHostKeyPath(hostKeyPath),
		// All connections accepted — no auth configured.
		// To restrict to specific public keys, add wish.WithPublicKeyAuth(fn).
		wish.WithMiddleware(
			bubbletea.Middleware(makeHandler(cfg)),
			logging.Middleware(),
		),
	)
	if err != nil {
		log.Fatalf("could not create server: %v", err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	log.Printf("urai-ssh listening on %s", addr)
	go func() {
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, net.ErrClosed) {
			log.Printf("server stopped: %v", err)
			stop <- syscall.SIGTERM
		}
	}()

	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}

func makeHandler(cfg *config.Config) bubbletea.Handler {
	return func(s ssh.Session) (tea.Model, []tea.ProgramOption) {
		pty, _, hasPTY := s.Pty()
		if !hasPTY {
			fmt.Fprintln(s, "urai requires a terminal. Connect with: ssh -t <host> urai")
			return nil, nil
		}
		_ = pty

		fp := ""
		if args := s.Command(); len(args) > 0 {
			fp = args[0]
		}

		m := app.New(fp, cfg)

		// wish/bubbletea middleware automatically wires session I/O and forwards
		// window-resize events as tea.WindowSizeMsgs, so no extra options needed.
		opts := []tea.ProgramOption{tea.WithAltScreen()}

		// Panic recovery: preserve content before the goroutine unwinds.
		go func() {
			defer func() {
				if r := recover(); r != nil {
					recovery.Write(m.Filepath(), m.Content()) //nolint:errcheck
					fmt.Fprintf(os.Stderr, "urai-ssh: session panic: %v\n", r)
				}
			}()
		}()

		return m, opts
	}
}

// ensureHostKey generates an Ed25519 host key at path if one doesn't exist.
func ensureHostKey(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil // already exists
	}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("create key directory: %w", err)
	}

	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return fmt.Errorf("generate key: %w", err)
	}

	b, err := gossh.MarshalPrivateKey(priv, "")
	if err != nil {
		return fmt.Errorf("marshal key: %w", err)
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("create key file: %w", err)
	}
	defer f.Close()

	if err := pem.Encode(f, b); err != nil {
		return fmt.Errorf("write key: %w", err)
	}

	log.Printf("generated new host key at %s", path)
	return nil
}
