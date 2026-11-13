package local

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/Pieczasz/tty-cards/internal/tui/router"
	"github.com/Pieczasz/tty-cards/internal/tui/styles"
	"github.com/Pieczasz/tty-cards/internal/tui/views"

	tea "charm.land/bubbletea/v2"
	lg "charm.land/lipgloss/v2"
)

// setup is the screen before and between games: the table size and the deal. It
// stands in for the server's home, lobby and lobby browser.
type setup struct {
	global router.GlobalContext
	table  *Table
	about  string
	err    error
}

func newSetup(g router.GlobalContext, t *Table, about string) tea.Model {
	return &setup{global: g, table: t, about: about}
}

func (m *setup) Init() tea.Cmd { return nil }

func (m *setup) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if handled, cmd := views.HandleCommonMsg(msg, &m.global); handled {
		return m, cmd
	}
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}
	slog.Debug("key", "view", "setup", "key", key.String())
	switch key.String() {
	case "q", "esc":
		m.table.Close()
		return m, tea.Quit
	case "left", "h", "-", "down", "j":
		m.table.SetSeats(m.table.Seats - 1)
	case "right", "l", "+", "=", "up", "k":
		m.table.SetSeats(m.table.Seats + 1)
	case "enter", "space", "d":
		seat, err := m.table.Deal(m.global.RequestContext())
		if err != nil {
			slog.Error("deal failed", "error", err)
			m.err = err
			return m, nil
		}
		m.err = nil
		return m, router.NavigateToGame(m.table.Game.Slug, seat)
	}
	return m, nil
}

func (m *setup) View() tea.View {
	g, t := m.global, m.global.Palette()
	actions := []string{"enter - Deal", "q - Quit"}
	mod := m.table.Game
	if mod.MinPlayers != mod.MaxPlayers {
		actions = append([]string{"left/right - Players"}, actions...)
	}
	footer := t.RenderActionFooter(actions)
	width := styles.InnerWidth(g.Width)
	title := t.Title.Render(styles.RenderFigureASCII(mod.Name, width, styles.TitleHeightBudget(g.Height)))

	opponents := m.table.Seats - 1
	who := "computer player"
	if opponents != 1 {
		who += "s"
	}
	rows := []string{
		t.Muted.Render(m.about),
		"",
		fmt.Sprintf("You against %s %s", t.Accented.Render(fmt.Sprint(opponents)), who),
		t.Dim.Render(strings.Join(botNames[:min(opponents, len(botNames))], ", ")),
	}
	if mod.MinPlayers != mod.MaxPlayers {
		rows = append(rows, "", t.Muted.Render(fmt.Sprintf("%d to %d players at the table", mod.MinPlayers, mod.MaxPlayers)))
	}
	rows = append(rows, "", t.SuccessText.Render("Press enter to deal"))
	if m.err != nil {
		rows = append(rows, "", t.ErrorText.Render(m.err.Error()))
	}
	body := styles.JoinVertical(lg.Center, rows...)
	return tea.NewView(views.RenderCenteredLayout(g, title, body, footer))
}
