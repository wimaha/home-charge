package html

import (
	"embed"
	"html/template"
	"io"

	"github.com/wimaha/home-charge/database"
)

//go:embed *
var files embed.FS

var (
	dashboard           = parse("dashboard.html")
	editScheduleCommand = parse("edit-schedule-command.html")
)

type DashboardParams struct {
	OperationMode     int
	OperationModeText string
	SOC               string
	BatteryCharging   string
	Pac_total_W       string
	WallboxStatus     int
	WallboxStatusText string
	ScheduleComands   []database.ScheduleCommand
}

func Dashboard(w io.Writer, p DashboardParams, partial string) error {
	if partial == "" {
		partial = "layout.html"
	}
	return dashboard.ExecuteTemplate(w, partial, p)
}

type EditScheduleCommandParams struct {
	BatteryCommands []database.BatteryCommand
	Title           string
}

func EditScheduleCommand(w io.Writer, p EditScheduleCommandParams, partial string) error {
	if partial == "" {
		partial = "layout.html"
	}
	return editScheduleCommand.ExecuteTemplate(w, partial, p)
}

func parse(file string) *template.Template {
	return template.Must(
		template.New("layout.html").ParseFS(files, "layout.html", file))
}
