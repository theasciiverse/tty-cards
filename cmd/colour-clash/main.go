// Command colour-clash plays Colour Clash in the terminal against computer players.
package main

import "github.com/Pieczasz/tty-cards/internal/local"

func main() {
	local.Main(local.Game{
		Slug:   "uno",
		Name:   "Colour Clash",
		About:  "Match colour, number or symbol. Skips, reverses and draw cards; wilds name the colour.",
		Seats:  4,
		Binary: "colour-clash",
	})
}
