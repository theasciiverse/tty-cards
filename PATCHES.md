# Changes from upstream

This is the Asciiverse packaging of [tty.cards](https://github.com/Pieczasz/tty-cards)
by Pieczasz (MIT), built from upstream's `main` at `85469f0` (31 October 2026;
upstream has no tags). Upstream is the `upstream` remote; the `asciiverse`
branch carries the changes below.

Upstream is an SSH server: you `ssh` in and play against other people at the
same server, with accounts, ratings and PostgreSQL behind it. Asciiverse lists
five of its games as titles of their own, played on one machine against
computer players. The server, its tables and its views are unchanged; the
local games reuse them.

## Computer players (written to go upstream)

- **`game.Bot`** (`internal/game/bot.go`): an optional interface a rules set
  implements to play a seat itself, `BotAction(state) Action`, read under the
  engine lock like the other hooks and only from what that seat could see.
- **One per game**, beside its rules (`bot.go` in each game package):
  - Crazy Eights: a matching card of the suit the hand is longest in, highest
    first; an eight only when nothing else fits, naming the longest suit;
    else a draw.
  - Uno: coloured cards first, action cards saved for a seat about to go out,
    then Wild, then Wild Draw Four, naming the colour held most.
  - Hearts: passes the three most dangerous cards (the timeout pass), ducks
    under the card winning the trick, dumps the queen of spades and high
    hearts when void, leads low.
  - Gin Rummy: takes the upcard only when it lowers the best deadwood,
    discards the card that leaves the least, knocks as soon as it may.
  - Poker: rates its cards (alone before the flop, the best five after,
    marked down when the board makes the hand), weighs that against the pot
    odds, and raises, calls, checks or folds, with a little randomness.
  - `internal/game/shed/bot.go` holds what the two shedding games share.
- **`lobby.Manager.SetEngineOptions`**: options every engine the manager
  starts gets, so a local table can turn the turn clock off
  (`game.WithTurnTimeout(0)`). Upstream only set this from tests.
- **Test** (`internal/local/bots_test.go`): computer players alone play three
  whole matches of every game at every table size; every move must be one
  the rules accept, and every match must finish.

## Local play (ours)

- **`internal/local`**: one game in this terminal. It opens the real lobby
  with an in-memory manager (no database: finished matches are written
  nowhere), seats computer players as guests named Ada, Bo, Cyd and on,
  readies them and the player, and drives each computer seat from engine
  events with a short think before each move and a longer one before
  dealing the next hand, so the result screen stays up. The turn clock is
  off: nobody is hurried or auto-played.
- **A setup screen** stands in for the server's home, lobby and browser
  (`internal/local/setup.go`): the game's name, how many sit down (left and
  right, within the game's limits), enter to deal, q to quit. A finished game
  and a forfeit both come back to it; a forfeit stands the computer seats up.
- **Five commands**, one per title (`cmd/crazy-eights`, `cmd/holdem`,
  `cmd/colour-clash`, `cmd/hearts`, `cmd/gin-rummy`), each `local.Main` with
  the game, its on-screen name and its default table size (4, 6, 4, 4, 2).
  Uno is shown as **Colour Clash**; Poker as **Hold'em**.
- **Flags:** `--players N` (the human included), `--name` (default: the
  login name), `--demo` (the computer plays the human's seat too and deals
  game after game; used for the previews).
- **Session logging:** every key, click and resize, every computer move, and
  a snapshot of the table each second (whose turn, every seat's hand size)
  to `$XDG_STATE_HOME/<command>/<command>.log` (`~/.local/state/...` on
  Linux, Application Support on macOS, LocalAppData on Windows), at debug
  with `TTY_CARDS_LOG=debug`. Nothing is written to the terminal.
- Playing each other over SSH stays upstream's server (`cmd/server`), which
  needs PostgreSQL; the local games do not offer it.

## Packaging

- `.github/workflows/asciiverse-release.yml`: tests the computer players,
  then builds each game for Linux (static), macOS and Windows, x86-64 and
  ARM64, as `<id>-<target>` archives, plus SHA256SUMS and a .deb, .rpm, .apk
  and Arch package per game (`asciiverse/nfpm.yaml`, filled in per game).
- `.github/workflows/vuln.yml` and `.github/dependabot.yml` removed here: the
  weekly scan and the dependency updates are upstream's.
- `asciiverse/record-preview.sh` and the five `asciiverse/<id>.cast` previews.
- An Asciiverse note at the top of README.md.

## Fixes after the first release

- **Hearts drawn as a plain U+2665** (`internal/tui/components/card.go`,
  the Crazy Eights suit picker), like the other three suits. Upstream added
  a U+FE0E text-presentation selector after it; xterm.js and
  asciinema-player count that selector as a cell of its own, so every row
  with a heart shifted one column and the card frames broke apart (seen in
  the website previews, 2026-11-13). U+2665 already defaults to text
  presentation. Written to go upstream.
- **Hearts trick centred** (`internal/tui/views/gameview/hearts/view.go`): the
  gap between the left and right cards of the trick cross is three cells, not
  two. Cards are odd-width, so with an even gap the top and bottom cards sat
  half a cell right of centre. Written to go upstream.
