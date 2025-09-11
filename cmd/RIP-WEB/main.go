package main

import (
	"log"

	"RIP-WEB/internal/api"
)

func main() {
	log.Println("Apliction start")
	api.StartServer()
	log.Println("Aplication terminated")
}
