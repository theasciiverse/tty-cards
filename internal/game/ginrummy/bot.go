package ginrummy

import (
	"github.com/Pieczasz/tty-cards/internal/deck"
	"github.com/Pieczasz/tty-cards/internal/game"
)

var _ game.Bot = (*Rules)(nil)

// BotAction takes the upcard only when it lowers the hand's best deadwood, discards
// the card that leaves the least deadwood behind, and knocks as soon as the rules
// allow it, on gin above all.
func (r *Rules) BotAction(state *game.State) game.Action {
	extra := extraOf(state)
	if extra.HandComplete() {
		if extra.MatchComplete {
			return nil
		}
		return ActionNextHand{}
	}
	hand := state.Players[state.CurrentTurn].Cards

	switch extra.Phase {
	case PhaseAwaitingDraw:
		if up, ok := state.Discard.Peek(); ok {
			_, _, now := bestMeldSplit(hand)
			if _, after := bestDiscard(append(append([]deck.Card(nil), hand...), up), &up); after < now {
				return validOr(r, state, ActionDrawDiscard{}, ActionDrawStock{})
			}
		}
		return validOr(r, state, ActionDrawStock{}, ActionDrawDiscard{})
	case PhaseAwaitingDiscard:
		card, _ := bestDiscard(hand, extra.TakenUpcard)
		if card == (deck.Card{}) {
			return nil
		}
		return validOr(r, state, ActionKnock{Discard: card}, ActionDiscard{Card: card})
	case PhaseHandOver:
	}
	return nil
}

// bestDiscard is the legal discard that leaves the least deadwood, and that deadwood.
func bestDiscard(hand []deck.Card, forbidden *deck.Card) (deck.Card, int) {
	var best deck.Card
	bestPts := -1
	for _, c := range hand {
		if !mayDiscard(c, forbidden) {
			continue
		}
		_, _, pts := bestMeldSplit(deck.RemoveOne(hand, c))
		if bestPts < 0 || pts < bestPts ||
			(pts == bestPts && deck.PipValue(c.Rank) > deck.PipValue(best.Rank)) {
			best, bestPts = c, pts
		}
	}
	return best, bestPts
}

// validOr is a when the rules accept it, else b.
func validOr(r *Rules, state *game.State, a, b game.Action) game.Action {
	if r.ValidateAction(state, a) == nil {
		return a
	}
	return b
}
