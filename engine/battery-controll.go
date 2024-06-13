package engine

import (
	"log"
	"time"

	"github.com/wimaha/home-charge/battery"
	"github.com/wimaha/home-charge/database"
	"github.com/wimaha/home-charge/wallbox"
)

func DoScheduleCommands(sonnenbatterie battery.Sonnenbatterie, wallboxInstance *wallbox.Mennekes) {
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
			} /*else {
				//fmt.Println("Hier B")
			}*/
		}
		// Trigger SOC
		if scheduleCommand.TriggerType == "SOC" {
			dateNow := time.Now() // Aktueller Zeitstempel
			dateConvert := scheduleCommand.TriggerTime
			soc, ok := sonnenbatterie.Soc()

			if dateNow.After(dateConvert) && ok && soc >= scheduleCommand.TriggerSOC {
				scheduleCommand.Triggered = true
				database.UpdateScheduleCommand(scheduleCommand)
				triggerCommand(sonnenbatterie, scheduleCommand)
			}
		}
	}

	homeChargeStatus, err := database.GetHomeChargeStatus()
	if !err && homeChargeStatus.WallboxAutomatic && wallboxInstance != nil {
		status, err := wallboxInstance.Status()
		if err != nil {
			log.Printf("Fehler beim Abruf des Wallbox-Status: %v\n", err)
		} else {
			if status == wallbox.StatusCharging {
				if sonnenbatterie.OperationMode() != 1 {
					log.Println("Wallbox lädt -> Batterie nicht entladen")
					sonnenbatterie.SetOperationMode(1)
					sonnenbatterie.StopDischargeBattery()

					logCommand := database.ScheduleCommand{
						Id:               1,
						BatteryCommandId: 3,
					}
					database.LogBatteryCommand(logCommand)
				}
			} else {
				if sonnenbatterie.OperationMode() != 2 {
					log.Println("Wallbox laden beendet -> Batterie in Automatic-Modus")
					sonnenbatterie.SetOperationMode(2)

					logCommand := database.ScheduleCommand{
						Id:               1,
						BatteryCommandId: 2,
					}
					database.LogBatteryCommand(logCommand)
				}
			}
		}
	}
}

func triggerCommand(sonnenbatterie battery.Sonnenbatterie, scheduleCommand database.ScheduleCommand) {
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

	//log.Println("Operation ausgeführt.")
}
