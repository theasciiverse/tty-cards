package poker

import (
	"math/rand/v2"

	"github.com/Pieczasz/tty-cards/internal/deck"
	"github.com/Pieczasz/tty-cards/internal/game"
)

var _ game.Bot = (*Rules)(nil)

// BotAction rates its two cards with whatever is on the board, weighs that against
// the price of staying in, and bets, calls, checks or folds on it, with a little
// randomness so it cannot be read like a table.
func (r *Rules) BotAction(state *game.State) game.Action {
	extra := extraOf(state)
	if extra.HandComplete() {
		if extra.MatchComplete {
			return nil
		}
		return ActionNextHand{}
	}
	p := state.Players[state.CurrentTurn]
	seat := extra.Seats[p.ID]
	if seat == nil || seat.Folded || seat.AllIn {
		return nil
	}

	strength := handStrength(p.Cards, extra.Table) + (rand.Float64()-0.5)*0.1 //nolint:gosec // table talk, not crypto
	toCall := extra.ToCall(p.ID)
	potOdds := 0.0
	if toCall > 0 {
		potOdds = float64(toCall) / float64(extra.Pool+toCall)
	}

	if lo, hi, ok := RaiseBounds(state, p.ID); ok && strength > 0.62 {
		target := extra.CurrentBet + max(extra.Pool/2, extra.BigBlind)
		if strength > 0.9 {
			target = extra.CurrentBet + extra.Pool
		}
		amount := min(max(target, lo), hi)
		if a := (ActionRaiseTo{Amount: amount}); r.ValidateAction(state, a) == nil {
			return a
		}
	}
	if toCall == 0 {
		return ActionCheck{}
	}
	if strength >= potOdds+0.12 || (toCall <= extra.BigBlind && strength > 0.3) {
		if r.ValidateAction(state, ActionCall{}) == nil {
			return ActionCall{}
		}
		if r.ValidateAction(state, ActionAllIn{}) == nil && strength > 0.7 {
			return ActionAllIn{}
		}
	}
	return ActionFold{}
}

// handStrength is a rough 0 to 1 rating: the two cards alone before the flop, the
// best five after it, marked down when the board makes the hand for everyone.
func handStrength(hole, board []deck.Card) float64 {
	if len(hole) < 2 {
		return 0
	}
	a, b := deck.RankValue(hole[0].Rank), deck.RankValue(hole[1].Rank)
	if len(board) < 3 {
		hi, lo := max(a, b), min(a, b)
		s := float64(hi+lo) / 28 * 0.55
		if a == b {
			s = 0.5 + float64(a)/28
		}
		if hole[0].Suit == hole[1].Suit {
			s += 0.05
		}
		if hi-lo == 1 {
			s += 0.04
		}
		return s
	}

	cards := append(append([]deck.Card(nil), hole...), board...)
	rank := classifyHand(cards).Rank
	strength := map[handRank]float64{
		rankHighCard: 0.15, rankPair: 0.45, rankTwoPair: 0.65, rankThreeOfAKind: 0.78,
		rankStraight: 0.86, rankFlush: 0.89, rankFullHouse: 0.94, rankFourOfAKind: 0.98,
		rankStraightFlush: 1,
	}[rank]
	if rank <= rankThreeOfAKind && !holeHelps(a, b, board) {
		return 0.2
	}
	if rank == rankPair && max(a, b) < highestBoardRank(board) && a != b {
		strength -= 0.12 // a pair under the board's top card
	}
	return strength
}

// holeHelps is a pocket pair or a hole card that pairs the board.
func holeHelps(a, b int, board []deck.Card) bool {
	if a == b {
		return true
	}
	for _, c := range board {
		if v := deck.RankValue(c.Rank); v == a || v == b {
			return true
		}
	}
	return false
}

func highestBoardRank(board []deck.Card) int {
	best := 0
	for _, c := range board {
		best = max(best, deck.RankValue(c.Rank))
	}
	return best
}
