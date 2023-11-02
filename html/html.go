package html

import (
	"embed"
	"html/template"
	"io"
)

//go:embed *
var files embed.FS

var (
	dashboard = parse("dashboard.html")
)

type DashboardParams struct {
	OperationMode     int
	OperationModeText string
	SOC               string
	BatteryCharging   string
	Pac_total_W       string
}

func Dashboard(w io.Writer, p DashboardParams, partial string) error {
	if partial == "" {
		partial = "layout.html"
	}
	return dashboard.ExecuteTemplate(w, partial, p)
}

func parse(file string) *template.Template {
	return template.Must(
		template.New("layout.html").ParseFS(files, "layout.html", file))
}
