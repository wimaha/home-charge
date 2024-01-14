package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/wimaha/home-charge/battery"
	"github.com/wimaha/home-charge/database"
	"github.com/wimaha/home-charge/engine"
	"github.com/wimaha/home-charge/html"
	"github.com/wimaha/home-charge/mqtt"
	"github.com/wimaha/home-charge/settings"
	"github.com/wimaha/home-charge/wallbox"
)

var env = settings.Environment{}

func main() {
	log.Println("HomeCharge is loading ...")
	log.Println("Config loading ...")
	var c settings.Conf
	env.Config = c.GetConf()
	if !env.Config.CheckConf(true) {
		os.Exit(1)
	}

	env.Battery = battery.NewSonnenbatterie(env.Config.Sonnenbatterie.ApiToken, env.Config.Sonnenbatterie.Host)

	if env.Config.Mqtt != nil {
		env.MqttClient = mqtt.NewMqttClient(env.Config.Mqtt.Host, env.Config.Mqtt.Port)
	} else {
		env.MqttClient = nil
	}
	if env.Config.InfluxDB != nil {
		env.InfluxClient = database.NewInfluxClient(env.Config.InfluxDB.Host, env.Config.InfluxDB.Port, env.Config.InfluxDB.Token, env.Config.InfluxDB.Organisation, env.Config.InfluxDB.Querys.ProductionTotal)
	} else {
		env.InfluxClient = nil
	}

	go startAutoControl()
	database.Setup()

	if env.Config.Wallbox != nil {
		env.WallboxInstance = wallbox.NewMennekes(env.Config.Wallbox.Host)
	} else {
		env.WallboxInstance = nil
	}

	log.Println("HomeCharge is running")
	startWebserver()
}

func startAutoControl() {
	for {
		//println("AutoControl")
		time.Sleep(5 * time.Second)
		env.Battery.Reload()
		engine.DoScheduleCommands(*env.Battery, env.WallboxInstance)
		engine.Awtrix_doAll(&env)
	}
}

func startWebserver() {
	http.HandleFunc("/", dashboard)
	http.HandleFunc("/save-settings", saveSettings)
	http.HandleFunc("/add-schedule-command", addScheduleCommand)
	http.HandleFunc("/save-schedule-command", saveScheduleCommand)
	http.HandleFunc("/delete-schedule-command", deleteScheduleCommand)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.ListenAndServe(":7618", nil)
}

func dashboard(w http.ResponseWriter, r *http.Request) {
	env.Battery.Reload()
	var wStatus wallbox.ChargeStatus
	var wStatusText string
	if env.WallboxInstance != nil {
		wStatus, wStatusText = env.WallboxInstance.StatusAndText()
	} else {
		wStatus = wallbox.StatusNotConfig
		wStatusText = ""
	}
	homeChargeStatus, _ := database.GetHomeChargeStatus()
	p := html.DashboardParams{
		OperationMode:     env.Battery.OperationMode(),
		OperationModeText: env.Battery.OperationModeText(),
		SOC:               env.Battery.SocText(),
		BatteryCharging:   env.Battery.BatteryCharging(),
		Pac_total_W:       env.Battery.PacTotalW(),
		WallboxStatus:     wStatus,
		WallboxStatusText: wStatusText,
		ScheduleComands:   database.GetScheduleCommands(),
		HomeChargeStatus:  homeChargeStatus,
	}
	html.Dashboard(w, p, "")
}

func saveSettings(w http.ResponseWriter, r *http.Request) {
	env.Battery.Reload()

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		/*operationMode := r.FormValue("operationMode")
		if operationMode == "1" || operationMode == "2" || operationMode == "10" {
			mode, err := strconv.Atoi(operationMode)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			sonnenbatterie.SetOperationMode(mode)
		}*/

		batterie := r.FormValue("batterie")
		if batterie == "auto" {
			env.Battery.SetOperationMode(2)
		} else if batterie == "nicht_entladen" {
			env.Battery.SetOperationMode(1)
			env.Battery.StopDischargeBattery()
		} else if batterie == "laden" {
			env.Battery.SetOperationMode(1)
			env.Battery.ChargeBattery()
		}

		wallboxAutomatic := r.FormValue("wallboxAutomatic")
		if wallboxAutomatic == "true" {
			//fmt.Println("wallboxAutomatic: true")
			homeChargeStatus, err := database.GetHomeChargeStatus()
			if !err {
				homeChargeStatus.WallboxAutomatic = true
				database.UpdateHomeChargeStatus(homeChargeStatus)
			}
		} else {
			//fmt.Println("wallboxAutomatic: false")
			homeChargeStatus, err := database.GetHomeChargeStatus()
			if !err {
				homeChargeStatus.WallboxAutomatic = false
				database.UpdateHomeChargeStatus(homeChargeStatus)
			}
		}
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func addScheduleCommand(w http.ResponseWriter, r *http.Request) {
	p := html.EditScheduleCommandParams{
		BatteryCommands: database.GetBatteryCommands(),
		Title:           "Geplante Einstellung hizufügen",
	}
	html.EditScheduleCommand(w, p, "")
}

func deleteScheduleCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		id, err := strconv.Atoi(r.FormValue("schedule-command-id"))
		if err != nil {
			return
		}
		database.DeleteScheduleCommand(id)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func saveScheduleCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		//fmt.Printf("%v", r)
		//action:[1] trigger:[time] triggerSOC:[] triggerTime:[2023-11-05T02:00]
		// Daten aus der Map in die Struktur umwandeln
		batteryCommandId, _ := strconv.Atoi(r.FormValue("action"))
		triggerSOC, err := strconv.Atoi(r.FormValue("triggerSOC"))
		if err != nil {
			triggerSOC = 0
		}
		weScheduleCmd := database.ScheduleCommand{
			BatteryCommandId: batteryCommandId, // Ersetzen Sie dies durch die tatsächliche ID
			TriggerType:      r.FormValue("trigger"),
			TriggerTime:      database.ParseTime(r.FormValue("triggerTime")),
			TriggerSOC:       triggerSOC,
			Triggered:        false,
		}
		database.AddScheduleCommand(weScheduleCmd)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
