package html

import (
	"embed"
	"html/template"
	"io"

	"github.com/wimaha/home-charge/database"
	"github.com/wimaha/home-charge/wallbox"
)

//go:embed *
var files embed.FS

var (
	dashboard           = parse("dashboard.html")
	editScheduleCommand = parse("edit-schedule-command.html")
)

const jsonString = `{"contactor_closed":{{.ContactorClosed}},"vehicle_connected":{{.VehicleConnected}},"session_s":{{.ChargingDuration}},"grid_v":238,"grid_hz":50.000,"vehicle_current_a":8,"currentA_a":{{.Current1}},"currentB_a":{{.Current3}},"currentC_a":{{.Current3}},"currentN_a":0.0,"voltageA_v":{{.Voltage1}},"voltageB_v":{{.Voltage2}},"voltageC_v":{{.Voltage3}},"relay_coil_v":6.1,"pcba_temp_c":14.1,"handle_temp_c":12.1,"mcu_temp_c":18.6,"uptime_s":42,"input_thermopile_uv":-172,"prox_v":1.5,"pilot_high_v":0.1,"pilot_low_v":0.1,"session_energy_wh":{{.SessionEnergyWh}},"config_status":5,"evse_state":11,"current_alerts":[]}`

var twc3Simulator, _ = template.New("").Parse(jsonString)

type DashboardParams struct {
	OperationMode     int
	OperationModeText string
	SOC               string
	BatteryCharging   string
	Pac_total_W       string
	WallboxStatus     wallbox.ChargeStatus
	WallboxStatusText string
	ScheduleComands   []database.ScheduleCommand
	HomeChargeStatus  database.HomeChargeStatus
	Connections       bool
	MqttStatus        string
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

type Twc3SimulatorParams struct {
	ContactorClosed  bool
	VehicleConnected bool
	ChargingDuration int
	Current1         float64
	Current2         float64
	Current3         float64
	Voltage1         float64
	Voltage2         float64
	Voltage3         float64
	SessionEnergyWh  int
}

func Twc3Simulator(w io.Writer, p Twc3SimulatorParams) error {
	return twc3Simulator.Execute(w, p)
}
