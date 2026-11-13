package local

import (
	"fmt"
	"testing"

	"github.com/Pieczasz/tty-cards/internal/catalog"
	"github.com/Pieczasz/tty-cards/internal/game"
)

// TestBotsPlayEveryGameToTheEnd seats only computer players at every table size of
// every game and plays whole matches: every move a bot makes must be one the rules
// accept, and every match must finish.
func TestBotsPlayEveryGameToTheEnd(t *testing.T) {
	for _, e := range catalog.Games {
		for seats := e.MinPlayers; seats <= e.MaxPlayers; seats++ {
			for round := range 3 {
				t.Run(fmt.Sprintf("%s/%d/%d", e.Slug, seats, round), func(t *testing.T) {
					playOut(t, e, seats)
				})
			}
		}
	}
}

func playOut(t *testing.T, e catalog.Entry, seats int) {
	t.Helper()
	rules := e.Factory()
	bot, ok := rules.(game.Bot)
	if !ok {
		t.Fatalf("%s has no computer player", e.Slug)
	}
	players := make([]*game.Player, seats)
	for i := range players {
		players[i] = &game.Player{ID: fmt.Sprintf("p%d", i), Name: fmt.Sprintf("P%d", i)}
	}
	engine := game.NewEngine(rules, players, game.WithTurnTimeout(0))
	if err := engine.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	defer engine.Close()

	const maxMoves = 20000
	for move := 0; move < maxMoves; move++ {
		if engine.IsFinished() {
			return
		}
		id := engine.CurrentPlayerID()
		var action game.Action
		engine.WithState(func(s *game.State) { action = bot.BotAction(s) })
		if action == nil {
			t.Fatalf("move %d: no move for %s", move, id)
		}
		if err := engine.SubmitAction(id, action); err != nil {
			t.Fatalf("move %d: %s refused %s: %v", move, id, action.Name(), err)
		}
	}
	t.Fatalf("no finish after %d moves", maxMoves)
}
