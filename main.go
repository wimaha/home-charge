package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	sonnenbatterie "github.com/wimaha/home-charge/battery"
	"github.com/wimaha/home-charge/html"
	yaml "gopkg.in/yaml.v3"
)

type conf struct {
	ApiToken  string `yaml:"apiToken"`
	BatteryIP string `yaml:"batteryIP"`
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

func main() {
	fmt.Println("HomeCharge is loading ...")
	var c conf
	config = c.getConf()
	fmt.Println("HomeCharge is running")
	startWebserver()
}

func startWebserver() {
	http.HandleFunc("/", dashboard)
	http.HandleFunc("/save-settings", saveSettings)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.ListenAndServe(":7618", nil)
}

func dashboard(w http.ResponseWriter, r *http.Request) {
	sonnenbatterie := sonnenbatterie.NewSonnenbatterie(config.ApiToken, config.BatteryIP)
	p := html.DashboardParams{
		OperationMode:     sonnenbatterie.OperationMode(),
		OperationModeText: sonnenbatterie.OperationModeText(),
		SOC:               sonnenbatterie.Soc(),
		BatteryCharging:   sonnenbatterie.BatteryCharging(),
		Pac_total_W:       sonnenbatterie.PacTotalW(),
	}
	html.Dashboard(w, p, "")
}

func saveSettings(w http.ResponseWriter, r *http.Request) {
	sonnenbatterie := sonnenbatterie.NewSonnenbatterie(config.ApiToken, config.BatteryIP)

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		operationMode := r.FormValue("operationMode")
		if operationMode == "1" || operationMode == "2" || operationMode == "10" {
			mode, err := strconv.Atoi(operationMode)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			sonnenbatterie.SetOperationMode(mode)
		}

		batterie := r.FormValue("batterie")
		if batterie == "nicht_entladen" {
			sonnenbatterie.StopDischargeBattery()
		} else if batterie == "laden" {
			sonnenbatterie.ChargeBattery()
		}
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
