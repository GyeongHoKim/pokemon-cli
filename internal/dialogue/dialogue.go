// Package dialogue provides self-authored, species-agnostic speech-bubble
// lines for the terminal pet UI. These are original text, not copyrighted
// in-game Pokédex or in-game encounter flavor text — see /NOTICE.
package dialogue

import (
	"fmt"
	"math/rand/v2"
)

var templates = []string{
	"Say hello to %s!",
	"Ooh, %s wandered in!",
	"Look who's here — %s!",
	"%s popped by to visit!",
	"A friendly %s showed up!",
	"Guess who dropped in — %s!",
	"%s wandered over to say hi!",
	"Here comes %s!",
	"%s just arrived, looking curious!",
	"Welcome %s to the terminal!",
}

// Arrival returns a random generic "a new companion showed up" line
// featuring displayName.
func Arrival(displayName string) string {
	return fmt.Sprintf(templates[rand.IntN(len(templates))], displayName)
}
