// Package local plays one tty.cards game on this machine: the player in their own
// terminal against computer seats, with no server, no account and no database. It
// reuses the real lobby, engine and table views; the computer seats are lobby guests
// that a driver goroutine plays through the rules' own game.Bot.
package local

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/Pieczasz/tty-cards/internal/game"
	"github.com/Pieczasz/tty-cards/internal/lobby"
)

// botNames are the computer seats, dealt in order.
var botNames = []string{"Ada", "Bo", "Cyd", "Dee", "Eli", "Fay", "Gus", "Ivy", "Jo"}

// Timing of a computer seat: a short think before each move, so a human can follow
// the table, and a longer one before dealing the next hand, so the result screen
// stays up long enough to read.
const (
	thinkMin     = 700 * time.Millisecond
	thinkSpread  = 700 * time.Millisecond
	nextHandWait = 6 * time.Second
)

// Table is the one table this program plays: the human, the manager holding the
// lobby, and the computer seats.
type Table struct {
	Manager *lobby.Manager
	Human   *game.Player
	Game    game.Module
	// Seats is how many sit down, the human included, within the game's limits.
	Seats int
	// Demo lets the computer play the human's seat too: the preview and --demo.
	Demo bool

	mu      sync.Mutex
	bots    []*game.Player
	drivers context.CancelFunc
}

// NewTable seats nobody yet; Lobby opens the table on first use.
func NewTable(m *lobby.Manager, human *game.Player, mod game.Module, seats int) *Table {
	return &Table{Manager: m, Human: human, Game: mod, Seats: clampSeats(mod, seats)}
}

func clampSeats(mod game.Module, n int) int { return min(max(n, mod.MinPlayers), mod.MaxPlayers) }

// SetSeats changes the table size for the next deal.
func (t *Table) SetSeats(n int) { t.Seats = clampSeats(t.Game, n) }

// Lobby is the human's table, opened if they sit at none.
func (t *Table) Lobby() (*lobby.Lobby, error) {
	if l := t.Manager.FindLobbyByPlayer(t.Human); l != nil {
		return l, nil
	}
	t.mu.Lock()
	t.bots = nil
	t.mu.Unlock()
	l, err := t.Manager.CreateLobby(t.Human,
		lobby.WithCardGame(t.Game.Slug), lobby.WithMaxPlayers(t.Game.MaxPlayers))
	if err != nil {
		return nil, fmt.Errorf("open the table: %w", err)
	}
	return l, nil
}

// Deal seats Seats-1 computer players, readies them and the human, which starts
// the game, and sets the computer seats playing. It returns the human's seat.
func (t *Table) Deal(ctx context.Context) (*game.BoundEngine, error) {
	l, err := t.Lobby()
	if err != nil {
		return nil, err
	}
	if err := t.fillSeats(ctx, l); err != nil {
		return nil, err
	}
	snap := l.Snapshot()
	for _, g := range snap.Guests {
		if !g.Ready {
			if err := t.Manager.ToggleReady(g.Player); err != nil {
				return nil, fmt.Errorf("ready %s: %w", g.Player.Name, err)
			}
		}
	}
	if !l.Snapshot().Leader.Ready {
		if err := t.Manager.ToggleReady(t.Human); err != nil {
			return nil, fmt.Errorf("start the game: %w", err)
		}
	}
	engine := l.ActiveGame()
	if engine == nil {
		return nil, fmt.Errorf("the game did not start")
	}
	t.startDrivers(engine)
	slog.Info("dealt", "game", t.Game.Slug, "seats", t.Seats)
	return l.Seat(t.Human.ID), nil
}

// fillSeats makes the computer seats exactly Seats-1.
func (t *Table) fillSeats(ctx context.Context, l *lobby.Lobby) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	want := t.Seats - 1
	if err := l.SetMaxPlayers(t.Human, t.Seats); err != nil && len(t.bots) <= want {
		// Raising first when growing; shrinking lowers the cap after the leaves.
		return fmt.Errorf("set the table size: %w", err)
	}
	for len(t.bots) > want {
		last := t.bots[len(t.bots)-1]
		t.Manager.LeaveLobby(last)
		t.bots = t.bots[:len(t.bots)-1]
	}
	for len(t.bots) < want {
		i := len(t.bots)
		bot := &game.Player{ID: fmt.Sprintf("cpu-%d", i+1), Name: botNames[i%len(botNames)]}
		if _, err := t.Manager.JoinLobbyByCode(ctx, l.Code(), bot); err != nil {
			return fmt.Errorf("seat %s: %w", bot.Name, err)
		}
		t.bots = append(t.bots, bot)
	}
	if err := l.SetMaxPlayers(t.Human, t.Seats); err != nil {
		return fmt.Errorf("set the table size: %w", err)
	}
	return nil
}

// Stop ends the computer seats' play.
func (t *Table) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.drivers != nil {
		t.drivers()
		t.drivers = nil
	}
}

// Close stands every computer seat up, which ends a table the human left mid-game.
func (t *Table) Close() {
	t.Stop()
	t.mu.Lock()
	bots := t.bots
	t.bots = nil
	t.mu.Unlock()
	for _, b := range bots {
		t.Manager.LeaveLobby(b)
	}
}

func (t *Table) startDrivers(engine *game.Engine) {
	t.Stop()
	ctx, cancel := context.WithCancel(context.Background())
	t.mu.Lock()
	t.drivers = cancel
	bots := append([]*game.Player(nil), t.bots...)
	t.mu.Unlock()
	rules := t.Game.Factory()
	bot, ok := rules.(game.Bot)
	if !ok {
		slog.Error("game has no computer player", "game", t.Game.Slug)
		return
	}
	if t.Demo {
		bots = append(bots, t.Human)
	}
	for _, b := range bots {
		go drive(ctx, engine, b.ID, bot)
	}
}

// drive plays seat id until the game ends or ctx is cancelled: on every engine event
// it checks whether the seat is on turn, thinks for a moment and submits its move.
func drive(ctx context.Context, engine *game.Engine, id string, bot game.Bot) {
	seat := game.Bind(engine, id)
	events, err := seat.Subscribe()
	if err != nil {
		slog.Error("computer seat could not watch the table", "seat", id, "error", err)
		return
	}
	defer seat.Unsubscribe(events)
	poll := time.NewTicker(time.Second)
	defer poll.Stop()

	for {
		if engine.IsFinished() {
			return
		}
		if engine.CurrentPlayerID() == id {
			var action game.Action
			handOver := false
			engine.WithState(func(s *game.State) {
				if s.CurrentTurn < len(s.Players) && s.Players[s.CurrentTurn].ID == id {
					action = bot.BotAction(s)
					if h, ok := s.Extra.(interface{ HandComplete() bool }); ok {
						handOver = h.HandComplete()
					}
				}
			})
			if action != nil {
				wait := thinkMin + time.Duration(rand.Int64N(int64(thinkSpread))) //nolint:gosec // pacing
				if handOver {
					wait = nextHandWait
				}
				select {
				case <-ctx.Done():
					return
				case <-time.After(wait):
				}
				if engine.CurrentPlayerID() == id {
					if err := seat.Submit(action); err != nil {
						slog.Warn("computer move refused", "seat", id, "action", action.Name(), "error", err)
					} else {
						slog.Debug("computer move", "seat", id, "action", action.Name())
					}
				}
				continue
			}
		}
		select {
		case <-ctx.Done():
			return
		case _, ok := <-events:
			if !ok {
				return
			}
		case <-poll.C:
		}
	}
}
