package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	sonnenbatterie "github.com/wimaha/home-charge/battery"
	"github.com/wimaha/home-charge/database"
	"github.com/wimaha/home-charge/engine"
	"github.com/wimaha/home-charge/html"
	"github.com/wimaha/home-charge/wallbox"
	yaml "gopkg.in/yaml.v3"
)

type conf struct {
	ApiToken  string `yaml:"apiToken"`
	BatteryIP string `yaml:"batteryIP"`
	WallboxIP string `yaml:"wallboxIP"`
}

func (c *conf) getConf() *conf {
	yamlFile, err := os.ReadFile("settings/config.yaml")
	if err != nil {
		fmt.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		fmt.Printf("Unmarshal: %v", err)
	}

	return c
}

var config *conf
var battery *sonnenbatterie.Sonnenbatterie
var wallboxInstance *wallbox.Mennekes

func main() {
	log.Println("HomeCharge is loading ...")
	var c conf
	config = c.getConf()
	battery = sonnenbatterie.NewSonnenbatterie(config.ApiToken, config.BatteryIP)

	go startAutoControl()
	database.Setup()

	wallboxInstance = wallbox.NewMennekes(config.WallboxIP)

	log.Println("HomeCharge is running")
	startWebserver()
}

func startAutoControl() {
	for {
		//println("AutoControl")
		time.Sleep(10 * time.Second)
		battery.Reload()
		engine.DoScheduleCommands(*battery)
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
	battery.Reload()
	//sonnenbatterie := sonnenbatterie.NewSonnenbatterie(config.ApiToken, config.BatteryIP)
	wStatus, wStatusText := wallboxInstance.StatusAndText()
	p := html.DashboardParams{
		OperationMode:     battery.OperationMode(),
		OperationModeText: battery.OperationModeText(),
		SOC:               battery.SocText(),
		BatteryCharging:   battery.BatteryCharging(),
		Pac_total_W:       battery.PacTotalW(),
		WallboxStatus:     wStatus,
		WallboxStatusText: wStatusText,
		ScheduleComands:   database.GetScheduleCommands(),
	}
	html.Dashboard(w, p, "")
}

func saveSettings(w http.ResponseWriter, r *http.Request) {
	battery.Reload()
	//sonnenbatterie := sonnenbatterie.NewSonnenbatterie(config.ApiToken, config.BatteryIP)

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
			battery.SetOperationMode(2)
		} else if batterie == "nicht_entladen" {
			battery.SetOperationMode(1)
			battery.StopDischargeBattery()
		} else if batterie == "laden" {
			battery.SetOperationMode(1)
			battery.ChargeBattery()
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
