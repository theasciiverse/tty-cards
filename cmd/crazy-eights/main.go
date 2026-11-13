// Command crazy-eights plays Crazy Eights in the terminal against computer players.
package main

import "github.com/Pieczasz/tty-cards/internal/local"

func main() {
	local.Main(local.Game{
		Slug:   "crazy_eights",
		Name:   "Crazy Eights",
		About:  "Match the suit or the rank; an eight is wild and names the next suit. First empty hand wins.",
		Seats:  4,
		Binary: "crazy-eights",
	})
}
