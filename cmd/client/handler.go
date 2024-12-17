package main

import (
	"fmt"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

func handlerPause(gs *gamelogic.GameState) func(state routing.PlayingState) {
	return func(state routing.PlayingState) {
		defer fmt.Print("> ")
		gs.HandlePause(state)
	}
}