package game

// Bot is a rules set that can play a seat itself: the computer opponents of a table
// played on one machine. BotAction is the move for the seat on turn, read under the
// engine lock like every other hook; it must be one ValidateAction accepts, and nil
// means the seat has nothing to do yet (it waits for the next event).
//
// A bot reads only what its seat could see at the table: its own hand and the cards
// face up. The state hands it everything, so that is a rule the implementations keep,
// not one the type enforces.
type Bot interface {
	BotAction(state *State) Action
}
