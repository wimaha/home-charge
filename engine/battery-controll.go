package engine

import (
	"log"
	"time"

	sonnenbatterie "github.com/wimaha/home-charge/battery"
	"github.com/wimaha/home-charge/database"
)

func DoScheduleCommands(sonnenbatterie sonnenbatterie.Sonnenbatterie) {
	scheduleCommands := database.GetScheduleCommands()

	for _, scheduleCommand := range scheduleCommands {
		// Trigger Time
		if scheduleCommand.TriggerType == "time" {
			dateNow := time.Now() // Aktueller Zeitstempel
			dateConvert := scheduleCommand.TriggerTime

			//fmt.Println("DateNow:", dateNow.Format("2006-01-02 15:04:05"), "; DateTrigger:", dateConvert.Format("2006-01-02 15:04:05"))
			//fmt.Printf("DateNow B: %v; DateTrigger B: %v", dateNow, dateConvert)

			if dateNow.After(dateConvert) {
				//fmt.Println("Hier A")
				scheduleCommand.Triggered = true
				database.UpdateScheduleCommand(scheduleCommand)
				triggerCommand(sonnenbatterie, scheduleCommand)
			} else {
				//fmt.Println("Hier B")
			}
		}
		// Trigger SOC
		if scheduleCommand.TriggerType == "SOC" {
			dateNow := time.Now() // Aktueller Zeitstempel
			dateConvert := scheduleCommand.TriggerTime
			soc := sonnenbatterie.Soc()

			if dateNow.After(dateConvert) && soc >= scheduleCommand.TriggerSOC {
				scheduleCommand.Triggered = true
				database.UpdateScheduleCommand(scheduleCommand)
				triggerCommand(sonnenbatterie, scheduleCommand)
			}
		}
	}
}

func triggerCommand(sonnenbatterie sonnenbatterie.Sonnenbatterie, scheduleCommand database.ScheduleCommand) {
	database.LogBatteryCommand(scheduleCommand)

	switch scheduleCommand.BatteryCommandId {
	case 1:
		sonnenbatterie.SetOperationMode(1)
		sonnenbatterie.ChargeBattery()
	case 2:
		sonnenbatterie.StopChargeBattery()
	case 3:
		sonnenbatterie.SetOperationMode(1)
		sonnenbatterie.StopDischargeBattery()
	case 4:
		sonnenbatterie.SetOperationMode(2)
	}

	log.Println("Operation ausgeführt.")
}
