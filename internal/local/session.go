package local

import (
	"log/slog"
	"time"

	"github.com/Pieczasz/tty-cards/internal/game"

	tea "charm.land/bubbletea/v2"
)

// snapshotEvery is how often the session log records the table.
const snapshotEvery = time.Second

type snapshotMsg struct{}

// demoKeyMsg is the demo autopilot pressing a key for nobody.
type demoKeyMsg struct{ code rune }

// session wraps the router: every key goes to the log before the views see it, the
// table is written to the log every second, and in demo mode it deals and moves on
// from finished games by itself.
type session struct {
	tea.Model
	table *Table
	// finishedAt is when the demo saw the current game end, zero while one runs.
	finishedAt time.Time
	idleSince  time.Time
}

func newSession(inner tea.Model, t *Table) *session {
	// The first deal comes two seconds in, not five: a preview shows play early.
	return &session{Model: inner, table: t, idleSince: time.Now().Add(-3 * time.Second)}
}

func snapshotTick() tea.Cmd {
	return tea.Tick(snapshotEvery, func(time.Time) tea.Msg { return snapshotMsg{} })
}

func (s *session) Init() tea.Cmd { return tea.Batch(s.Model.Init(), snapshotTick()) }

func (s *session) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.KeyPressMsg:
		slog.Debug("key", "key", m.String())
	case tea.MouseClickMsg:
		slog.Debug("click", "x", m.X, "y", m.Y)
	case tea.WindowSizeMsg:
		slog.Debug("resize", "width", m.Width, "height", m.Height)
	case demoKeyMsg:
		msg = tea.KeyPressMsg{Code: m.code}
	case snapshotMsg:
		s.logTable()
		return s, tea.Batch(snapshotTick(), s.demoStep())
	}
	var cmd tea.Cmd
	s.Model, cmd = s.Model.Update(msg)
	return s, cmd
}

// demoStep presses enter on the setup screen and on a finished game's results, a
// few seconds after each comes up.
func (s *session) demoStep() tea.Cmd {
	if !s.table.Demo {
		return nil
	}
	const dwell = 5 * time.Second
	engine := s.activeEngine()
	switch {
	case engine == nil:
		if time.Since(s.idleSince) < dwell {
			return nil
		}
		s.idleSince = time.Now()
	case engine.IsFinished():
		if s.finishedAt.IsZero() {
			s.finishedAt = time.Now()
			return nil
		}
		if time.Since(s.finishedAt) < dwell {
			return nil
		}
		s.finishedAt = time.Time{}
		s.idleSince = time.Now()
	default:
		s.idleSince = time.Now()
		return nil
	}
	return func() tea.Msg { return demoKeyMsg{code: tea.KeyEnter} }
}

func (s *session) activeEngine() *game.Engine {
	l := s.table.Manager.FindLobbyByPlayer(s.table.Human)
	if l == nil {
		return nil
	}
	return l.ActiveGame()
}

// logTable writes whose turn it is and every seat's hand size: what the table shows.
func (s *session) logTable() {
	engine := s.activeEngine()
	if engine == nil {
		slog.Debug("snapshot", "screen", "setup", "seats", s.table.Seats)
		return
	}
	finished := engine.IsFinished() // before WithState: the engine lock is not reentrant
	engine.WithState(func(st *game.State) {
		hands := make([]any, 0, 2*len(st.Players))
		for _, p := range st.Players {
			hands = append(hands, p.Name, len(p.Cards))
		}
		turn := ""
		if st.CurrentTurn < len(st.Players) {
			turn = st.Players[st.CurrentTurn].Name
		}
		slog.Debug("snapshot", append([]any{"turn", turn, "finished", finished}, hands...)...)
	})
}
