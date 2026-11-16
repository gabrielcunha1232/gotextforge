package desktop

import (
	"embed"
	"io/fs"

	"github.com/gabrielcunha1232/gotextforge/internal/core/editor"
	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

func Run(ed *editor.Editor) error {
	dist, err := fs.Sub(assets, "frontend/dist")
	if err != nil {
		return err
	}

	var app *application.App
	service := NewEditorService(ed,
		func() { app.Quit() },
		func(f Frame) { app.Event.Emit(FrameEvent, f) },
	)

	app = application.New(application.Options{
		Name: "gotextforge",
		Services: []application.Service{
			application.NewService(service),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(dist),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "gotextforge",
		Width:            1200,
		Height:           800,
		BackgroundColour: application.NewRGB(0x0b, 0x0e, 0x13),
		URL:              "/",
	})

	return app.Run()
}
