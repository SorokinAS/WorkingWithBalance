package config

import (
	"log"

	"github.com/joho/godotenv"
)

func GetConfig() {
	// err := godotenv.Load("./config/.env") // for Linux
	err := godotenv.Load(".\\config\\.env") // for Windows
	if err != nil {
		log.Fatal("Error with parse config ", err.Error())
	}
}
