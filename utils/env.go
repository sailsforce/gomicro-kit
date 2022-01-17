package utils

import "github.com/joho/godotenv"

func InitEnv() error {
	return godotenv.Load()
}
