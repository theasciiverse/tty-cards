// Command gin-rummy plays Gin Rummy in the terminal against computer players.
package main

import "github.com/Pieczasz/tty-cards/internal/local"

func main() {
	local.Main(local.Game{
		Slug:   "gin_rummy",
		Name:   "Gin Rummy",
		About:  "Draw, meld, discard. Knock at 10 deadwood or less, gin at none. Match ends at 100.",
		Seats:  2,
		Binary: "gin-rummy",
	})
}
