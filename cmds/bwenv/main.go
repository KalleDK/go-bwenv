package main

import (
	"fmt"
	"log"
	"os"

	"github.com/KalleDK/go-bwenv/bwenv"
)

func main() {

	BW_SESSION := os.Getenv("BW_SESSION")
	BW_FOLDER := os.Getenv("BW_FOLDER")

	bw := bwenv.EnvConfig{
		Config: bwenv.Config{
			Key: BW_SESSION,
		},
		Folder: BW_FOLDER,
	}.New()

	item, err := bw.GetEnv("git")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(item)

}
