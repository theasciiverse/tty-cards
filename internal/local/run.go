package local

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"unicode"

	"github.com/Pieczasz/tty-cards/internal/catalog"
	"github.com/Pieczasz/tty-cards/internal/game"
	"github.com/Pieczasz/tty-cards/internal/lobby"
	"github.com/Pieczasz/tty-cards/internal/tui/router"

	tea "charm.land/bubbletea/v2"
)

// Game is one title built from the collection: which catalog entry it plays, what it
// is called on screen, a line about it, and how many sit down by default.
type Game struct {
	Slug  string
	Name  string
	About string
	Seats int
	// Binary names the program and its folders: logs, nothing else.
	Binary string
}

// Main runs g and exits with its status: the whole of a title's main.
func Main(g Game) {
	if err := Run(g, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, g.Binary+":", err)
		os.Exit(1)
	}
}

// Run plays g in this terminal until the player quits.
func Run(g Game, args []string) error {
	flags := flag.NewFlagSet(g.Binary, flag.ContinueOnError)
	seats := flags.Int("players", g.Seats, "players at the table, you included")
	name := flags.String("name", "", "your name at the table (default: your login)")
	demo := flags.Bool("demo", false, "the computer plays your seat too, game after game")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err //nolint:wrapcheck // flag's own message
	}

	closeLog := setupLogging(g.Binary)
	defer closeLog()

	entry, ok := findEntry(g.Slug)
	if !ok {
		return fmt.Errorf("no game %q in this build", g.Slug)
	}
	entry.Name = g.Name

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	games := []catalog.Entry{entry}
	mgr := lobby.NewManager(ctx, nil, game.NewRegistry(catalog.Modules(games)...))
	mgr.SetEngineOptions(game.WithTurnTimeout(0))

	human := &game.Player{ID: "you", Name: playerName(*name)}
	table := NewTable(mgr, human, entry.Module, *seats)
	table.Demo = *demo
	defer table.Close()

	global := router.GlobalContext{
		Player:       human,
		LobbyManager: mgr,
		Games:        catalog.Modules(games),
		SessionCtx:   ctx,
	}
	r := router.New(global)
	home := func(gc router.GlobalContext, _ router.Payload) tea.Model {
		table.Stop()
		if l := mgr.FindLobbyByPlayer(human); l == nil {
			table.Close() // left mid-game: stand the computer seats up too
		}
		return newSetup(gc, table, g.About)
	}
	for _, route := range []router.Route{
		router.RouteHome, router.RouteLobby, router.RouteProfile, router.RouteLeaderboard,
		router.RouteLobbyCreate, router.RouteLobbyJoin,
	} {
		r.Register(route, home)
	}
	r.Register(router.GameRoute(entry.Slug), func(gc router.GlobalContext, p router.Payload) tea.Model {
		if p.Table == nil {
			return newSetup(gc, table, g.About)
		}
		return entry.View(gc, p.Table, entry.Slug)
	})

	slog.Info("start", "game", g.Slug, "seats", table.Seats)
	_, err := tea.NewProgram(newSession(r, table)).Run()
	slog.Info("quit", "error", err)
	if err != nil {
		return fmt.Errorf("run: %w", err)
	}
	return nil
}

func findEntry(slug string) (catalog.Entry, bool) {
	for _, e := range catalog.Games {
		if e.Slug == slug {
			return e, true
		}
	}
	return catalog.Entry{}, false
}

// playerName is the name shown at the player's seat: the flag, else the login,
// capitalised, else "You".
func playerName(flagName string) string {
	n := strings.TrimSpace(flagName)
	for _, v := range []string{"USER", "USERNAME"} {
		if n == "" {
			n = strings.TrimSpace(os.Getenv(v))
		}
	}
	if n == "" {
		if u, err := user.Current(); err == nil {
			n = u.Username
			if i := strings.LastIndexAny(n, `\`); i >= 0 {
				n = n[i+1:]
			}
		}
	}
	if n == "" {
		return "You"
	}
	r := []rune(n)
	r[0] = unicode.ToUpper(r[0])
	if len(r) > 16 {
		r = r[:16]
	}
	return string(r)
}

// setupLogging writes the session log (RUST_LOG-style: TTY_CARDS_LOG=debug for
// every key and move) to the player's state folder, never to the terminal.
func setupLogging(binary string) func() {
	level := slog.LevelInfo
	if os.Getenv("TTY_CARDS_LOG") == "debug" {
		level = slog.LevelDebug
	}
	dir := stateDir(binary)
	var w io.Writer = io.Discard
	closer := func() {}
	if err := os.MkdirAll(dir, 0o755); err == nil {
		if f, err := os.OpenFile(filepath.Join(dir, binary+".log"),
			os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644); err == nil {
			w, closer = f, func() { _ = f.Close() }
		}
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{Level: level})))
	return closer
}

// stateDir is $XDG_STATE_HOME/<binary>, else the platform's usual place.
func stateDir(binary string) string {
	if d := os.Getenv("XDG_STATE_HOME"); d != "" {
		return filepath.Join(d, binary)
	}
	home, _ := os.UserHomeDir()
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", binary)
	case "windows":
		if d := os.Getenv("LocalAppData"); d != "" {
			return filepath.Join(d, binary)
		}
	}
	return filepath.Join(home, ".local", "state", binary)
}
