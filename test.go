package main

import (
	"log"
	"os"
)

func openTestHTML(fileName string) *os.File {
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	return file
}
