package utils

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

func PingDB(db *gorm.DB) error {
	ping := db.Exec("SELECT * FROM information_schema.information_schema_catalog_name")
	return ping.Error
}

func GetDSN(databaseURL string) string {
	var host string
	var user string
	var password string
	var dbname string
	var port string

	s1 := strings.Split(databaseURL, "://")
	s2 := strings.Split(s1[1], ":")
	user = s2[0]
	s3 := strings.Split(s2[1], "@")
	password = s3[0]
	host = s3[1]
	s4 := strings.Split(s2[2], "/")
	port = s4[0]
	dbname = s4[1]

	return fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v sslmode=disable", host, user, password, dbname, port)
}
