package shed

import (
	"github.com/Pieczasz/tty-cards/internal/deck"
	"github.com/Pieczasz/tty-cards/internal/game"
)

// SuitCounts is how many cards of each suit hand holds, ignoring the cards wild
// says to skip: the colour a bot names is the one it holds most of.
func SuitCounts(hand []deck.Card, wild func(deck.Card) bool) map[deck.Suit]int {
	counts := make(map[deck.Suit]int, 4)
	for _, c := range hand {
		if wild != nil && wild(c) {
			continue
		}
		if deck.IsSuit(c.Suit) {
			counts[c.Suit]++
		}
	}
	return counts
}

// BestSuit is the suit hand holds most of after playing played, among suits, the
// first of them on a tie or an empty hand.
func BestSuit(hand []deck.Card, played deck.Card, suits []deck.Suit, wild func(deck.Card) bool) deck.Suit {
	counts := SuitCounts(deck.RemoveOne(hand, played), wild)
	best, bestN := suits[0], -1
	for _, s := range suits {
		if counts[s] > bestN {
			best, bestN = s, counts[s]
		}
	}
	return best
}

// FewestCardsElsewhere is the smallest hand any other seat holds: the threat a
// shedding bot answers with its action cards. Hand sizes are face-up information.
func FewestCardsElsewhere(state *game.State) int {
	fewest := -1
	for i, p := range state.Players {
		if i == state.CurrentTurn {
			continue
		}
		if fewest < 0 || len(p.Cards) < fewest {
			fewest = len(p.Cards)
		}
	}
	return fewest
}
