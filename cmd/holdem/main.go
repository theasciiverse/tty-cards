// Command holdem plays Hold'em in the terminal against computer players.
package main

import "github.com/Pieczasz/tty-cards/internal/local"

func main() {
	local.Main(local.Game{
		Slug:   "poker",
		Name:   "Hold'em",
		About:  "No-limit Texas Hold'em, a ten-hand match. 1000 chips each, blinds 25/50, real side pots.",
		Seats:  6,
		Binary: "holdem",
	})
}
