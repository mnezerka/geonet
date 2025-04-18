package main

import (
	"embed"
	"mnezerka/geonet/cmd"
)

//go:embed templates/*
var templatesContent embed.FS

func main() {

	cmd.SetTemplatesContent(&templatesContent)

	cmd.Execute()
}
