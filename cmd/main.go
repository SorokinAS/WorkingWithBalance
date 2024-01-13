package main

import (
	"postgres-test/config"
	"postgres-test/service"
)

func main() {
	config.GetConfig()
	service.Run()
}
