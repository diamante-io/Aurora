package env

import (
	"log"
	"os"
	"strconv"
)

// String returns the value of the environment variable "name".
// If name is not set, it returns value.
func String(name string, value string) string {
	if s := os.Getenv(name); s != "" {
		value = s
	}
	return value
}

// Int returns the value of the environment variable "name" as an int.
// If name is not set, it returns value.
func Int(name string, value int) int {
	if s := os.Getenv(name); s != "" {
		var err error
		value, err = strconv.Atoi(s)
		if err != nil {
			log.Println(name, err)
			os.Exit(1)
		}
	}
	return value
}
