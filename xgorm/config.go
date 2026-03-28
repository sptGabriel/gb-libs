package xgorm

import "fmt"

type Config struct {
	DBName     string
	DBHost     string
	DBUser     string
	DBPort     string
	DBEnabled  bool
	DBPassword string
}

func (c Config) Url() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s",
		c.DBHost, c.DBUser, c.DBPassword, c.DBName, c.DBPort)
}
