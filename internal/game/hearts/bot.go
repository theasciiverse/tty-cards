package hearts

import (
	"github.com/Pieczasz/tty-cards/internal/deck"
	"github.com/Pieczasz/tty-cards/internal/game"
)

var _ game.Bot = (*Rules)(nil)

// BotAction passes the three most dangerous cards, the absent player's pass, and
// plays a trick the way a careful beginner does: duck under the card winning the
// trick, dump the queen of spades and high hearts when void, and lead low.
func (r *Rules) BotAction(state *game.State) game.Action {
	extra := extraOf(state)
	p := state.Players[state.CurrentTurn]
	switch extra.Phase {
	case PhaseHandOver:
		if extra.MatchComplete {
			return nil
		}
		return ActionNextHand{}
	case PhasePassing:
		return ActionPassCards{Cards: threeMostDangerous(p.Cards)}
	case PhaseTrickPlay:
	}

	var legal []deck.Card
	for _, c := range p.Cards {
		if validatePlayCard(extra, p, c) == nil {
			legal = append(legal, c)
		}
	}
	if len(legal) == 0 {
		return nil
	}
	return ActionPlayCard{Card: chooseCard(state, extra, legal)}
}

func chooseCard(state *game.State, extra *State, legal []deck.Card) deck.Card {
	if extra.leadingTrick() {
		return lowestLead(legal)
	}
	if legal[0].Suit != extra.LedSuit {
		return dump(legal)
	}

	winning := 0
	for _, c := range extra.TrickCards {
		if c.Suit == extra.LedSuit {
			winning = max(winning, deck.RankValue(c.Rank))
		}
	}
	// Duck: the highest card that still loses the trick.
	if c, ok := pick(legal, func(c deck.Card) bool { return deck.RankValue(c.Rank) < winning }, true); ok {
		return c
	}
	// Every card wins it. Last to play on a clean trick, take it with the highest;
	// otherwise win it as cheaply as possible.
	last := len(extra.TrickCards) == len(state.Players)-1
	if last && trickPoints(extra.TrickCards) == 0 {
		c, _ := pick(legal, func(c deck.Card) bool { return c != queenOfSpades }, true)
		if c != (deck.Card{}) {
			return c
		}
	}
	c, _ := pick(legal, func(deck.Card) bool { return true }, false)
	return c
}

// lowestLead leads the lowest card that is neither a heart nor a spade that could
// catch the queen, falling back to the lowest card at all.
func lowestLead(legal []deck.Card) deck.Card {
	safe := func(c deck.Card) bool {
		if c.Suit == deck.Hearts {
			return false
		}
		return c.Suit != deck.Spades || deck.RankValue(c.Rank) < deck.RankValue(deck.Queen)
	}
	if c, ok := pick(legal, safe, false); ok {
		return c
	}
	c, _ := pick(legal, func(deck.Card) bool { return true }, false)
	return c
}

// dump is the void player's discard: the queen of spades, then the highest heart,
// then the ace or king of spades, then the highest card left.
func dump(legal []deck.Card) deck.Card {
	for _, c := range legal {
		if c == queenOfSpades {
			return c
		}
	}
	if c, ok := pick(legal, func(c deck.Card) bool { return c.Suit == deck.Hearts }, true); ok {
		return c
	}
	if c, ok := pick(legal, func(c deck.Card) bool {
		return c.Suit == deck.Spades && deck.RankValue(c.Rank) > deck.RankValue(deck.Queen)
	}, true); ok {
		return c
	}
	c, _ := pick(legal, func(deck.Card) bool { return true }, true)
	return c
}

// pick is the highest (or lowest) card of cards that ok accepts.
func pick(cards []deck.Card, ok func(deck.Card) bool, highest bool) (deck.Card, bool) {
	var best deck.Card
	found := false
	for _, c := range cards {
		if !ok(c) {
			continue
		}
		v, bv := deck.RankValue(c.Rank), deck.RankValue(best.Rank)
		if !found || (highest && v > bv) || (!highest && v < bv) {
			best, found = c, true
		}
	}
	return best, found
}
