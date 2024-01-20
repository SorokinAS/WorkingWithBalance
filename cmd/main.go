package main

import (
	"postgres-test/config"
	"postgres-test/handler"
)

func main() {
	config.GetConfig()
	handler.Run()
}
