package model

import "github.com/gofiber/fiber/v3/log"

func handleErr(err error) {
	log.Error(err)
}
