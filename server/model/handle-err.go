package model

import (
	"log"

	"github.com/Mopsgamer/draqun/server/debug"
)

func handleErr(err error) {
	log.Println(err)
	debug.PrintStack()
}
