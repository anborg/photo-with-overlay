package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()
	err := wails.Run(&options.App{
		Title: "Field Photo Capture", Width: 1200, Height: 820, MinWidth: 900, MinHeight: 650,
		AssetServer:      &assetserver.Options{Assets: assets},
		BackgroundColour: &options.RGBA{R: 23, G: 32, B: 42, A: 1},
		OnStartup:        app.startup, OnShutdown: app.shutdown,
		Bind: []interface{}{app},
	})
	if err != nil {
		log.Fatal(err)
	}
}
