package uno

import (
	"github.com/Pieczasz/tty-cards/internal/deck"
	"github.com/Pieczasz/tty-cards/internal/game"
	"github.com/Pieczasz/tty-cards/internal/game/shed"
)

var _ game.Bot = (*Rules)(nil)

var colors = []deck.Suit{ColorRed, ColorYellow, ColorGreen, ColorBlue}

// threatCards is the hand size at which a bot starts spending its action cards on
// the seat about to go out.
const threatCards = 2

// BotAction plays a coloured card when it can: an action card while someone is close
// to going out, otherwise a number in the colour it holds most of, the highest
// first. A Wild comes next and a Wild Draw Four last, both naming the colour the hand
// is longest in; with nothing playable it draws.
func (r *Rules) BotAction(state *game.State) game.Action {
	hand := state.Players[state.CurrentTurn].Cards
	counts := shed.SuitCounts(hand, IsWild)
	threatened := shed.FewestCardsElsewhere(state) <= threatCards

	var best deck.Card
	bestScore := -1
	for _, c := range hand {
		if IsWild(c) || r.ValidateAction(state, ActionPlayCard{Card: c}) != nil {
			continue
		}
		if s := playScore(c, counts, threatened); s > bestScore {
			best, bestScore = c, s
		}
	}
	if bestScore >= 0 {
		return ActionPlayCard{Card: best}
	}
	for _, rank := range []deck.Rank{Wild, WildDrawFour} {
		for _, c := range hand {
			if c.Rank != rank {
				continue
			}
			a := ActionPlayCard{Card: c, ChosenSuit: shed.BestSuit(hand, c, colors, IsWild)}
			if r.ValidateAction(state, a) == nil {
				return a
			}
		}
	}
	return ActionDrawCard{}
}

// playScore ranks a playable coloured card: the colour held most counts most, an
// action card jumps ahead under threat and waits otherwise, and a higher number
// sheds more points.
func playScore(c deck.Card, counts map[deck.Suit]int, threatened bool) int {
	const colourWeight, actionBonus = 100, 1000
	score := counts[c.Suit] * colourWeight
	action := c.Rank == Skip || c.Rank == Reverse || c.Rank == DrawTwo
	switch {
	case action && threatened:
		score += actionBonus
	case action:
		score += 5
	default:
		score += numberValue(c.Rank)
	}
	return score
}

func numberValue(r deck.Rank) int {
	if r == Zero {
		return 0
	}
	if r == One {
		return 1
	}
	return deck.RankValue(r)
}
