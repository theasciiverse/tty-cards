package crazyeight

import (
	"github.com/Pieczasz/tty-cards/internal/deck"
	"github.com/Pieczasz/tty-cards/internal/game"
	"github.com/Pieczasz/tty-cards/internal/game/shed"
)

var _ game.Bot = (*Rules)(nil)

var suits = []deck.Suit{deck.Spades, deck.Hearts, deck.Diamonds, deck.Clubs}

func isEight(c deck.Card) bool { return c.Rank == deck.Eight }

// BotAction plays the card that keeps the most options open: a matching card of
// the suit the hand is longest in, the highest first so a lost hand costs less; an
// eight only when nothing else fits, naming the suit left longest; otherwise a draw.
func (r *Rules) BotAction(state *game.State) game.Action {
	hand := state.Players[state.CurrentTurn].Cards
	counts := shed.SuitCounts(hand, isEight)

	var best deck.Card
	found := false
	for _, c := range hand {
		if isEight(c) || r.ValidateAction(state, ActionPlayCard{Card: c}) != nil {
			continue
		}
		if !found || betterShed(c, best, counts) {
			best, found = c, true
		}
	}
	if found {
		return ActionPlayCard{Card: best}
	}
	for _, c := range hand {
		if isEight(c) {
			return ActionPlayCard{Card: c, ChosenSuit: shed.BestSuit(hand, c, suits, isEight)}
		}
	}
	return ActionDrawCard{}
}

// betterShed prefers the suit held most, then the higher card.
func betterShed(c, than deck.Card, counts map[deck.Suit]int) bool {
	if counts[c.Suit] != counts[than.Suit] {
		return counts[c.Suit] > counts[than.Suit]
	}
	return deck.RankValue(c.Rank) > deck.RankValue(than.Rank)
}
