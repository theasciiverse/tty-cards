// Command hearts plays Hearts in the terminal against computer players.
package main

import "github.com/Pieczasz/tty-cards/internal/local"

func main() {
	local.Main(local.Game{
		Slug:   "hearts",
		Name:   "Hearts",
		About:  "Pass three, follow suit, dodge the hearts and the queen of spades. Match ends at 100.",
		Seats:  4,
		Binary: "hearts",
	})
}
