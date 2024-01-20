package aifui

import (
	"embed"
	"fmt"
	"log"
	"os"
)

//go:embed asset/*
var Asset embed.FS

var downloadDest = "sdk.tar.gz"
var extractDest = "asset/amis"

func updateBefore() {
	err := os.RemoveAll(extractDest)
	if err != nil {
		log.Println(err)
	}
	os.Remove("VERSION")
}

func updateVersion(ver string) {
	file, _ := os.Create("VERSION")
	file.WriteString(fmt.Sprintf("amis\t%v", ver))
}
