package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func Setup() {
	db, err := sql.Open("sqlite3", "database/home-charge.db")
	if err != nil {
		log.Fatal("Open: ", err)
	}

	sql := `SELECT * FROM weBatteryCommand`
	_, err = db.Exec(sql)
	if err != nil {
		CreateTables(db)
	}

	sql = `SELECT * FROM weHomeChargeStatus`
	_, err = db.Exec(sql)
	if err != nil {
		CreateTablesV2(db)
	}

	defer db.Close()
}

func CreateTables(db *sql.DB) {
	fmt.Println("Create Tables")

	sql := `CREATE TABLE IF NOT EXISTS weBatteryCommand (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL
	  );
	  INSERT INTO weBatteryCommand (id, name) VALUES
	  (1, 'Batterie laden'),
	  (2, 'Batterie laden stoppen'),
	  (3, 'Batterie entladen stoppen'),
	  (4, 'Automatischen Modus wieder aktivieren');`

	_, err := db.Exec(sql)

	if err != nil {
		log.Fatal("CREATE TABLE weBatteryCommand: ", err)
	}
	fmt.Println("table weBatteryCommand created")

	sql = `CREATE TABLE IF NOT EXISTS weBatteryCommandLog (
		id INTEGER PRIMARY KEY,
		weBatteryCommand_id INTEGER,
		triggerTime DATETIME DEFAULT CURRENT_TIMESTAMP
	  );`

	_, err = db.Exec(sql)

	if err != nil {
		log.Fatal("CREATE TABLE weBatteryCommandLog: ", err)
	}
	fmt.Println("table weBatteryCommandLog created")

	//triggerType: time,SOC
	sql = `CREATE TABLE IF NOT EXISTS weScheduleCommand (
		id INTEGER PRIMARY KEY,
		weBatteryCommand_id INTEGER NOT NULL,
		triggerType TEXT NOT NULL,
		triggerTime DATETIME NOT NULL,
		triggerSOC INTEGER NOT NULL DEFAULT '0',
		triggered BOOLEAN NOT NULL DEFAULT '0'
	  );`

	_, err = db.Exec(sql)

	if err != nil {
		log.Fatal("CREATE TABLE weScheduleCommand: ", err)
	}
	fmt.Println("table weScheduleCommand created")
}

func CreateTablesV2(db *sql.DB) {
	fmt.Println("Create TablesV2")

	sql := `CREATE TABLE IF NOT EXISTS weHomeChargeStatus (
		id INTEGER PRIMARY KEY,
		version INTEGER NOT NULL,
		wallboxAutomatic BOOLEAN NOT NULL DEFAULT '1'
	  );
	  INSERT INTO weHomeChargeStatus (id, version, wallboxAutomatic) VALUES
	  (1, 2, 1);`

	_, err := db.Exec(sql)

	if err != nil {
		log.Fatal("CREATE TABLE weHomeChargeStatus: ", err)
	}
	fmt.Println("table weHomeChargeStatus created")
}

type HomeChargeStatus struct {
	Id               int
	Version          int
	WallboxAutomatic bool
}

func GetHomeChargeStatus() (HomeChargeStatus, bool) {
	db, err := sql.Open("sqlite3", "database/home-charge.db")
	if err != nil {
		log.Fatal("Open: ", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM weHomeChargeStatus WHERE id = 1")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	if rows.Next() {
		var id int
		var version int
		var wallboxAutomatic bool

		err = rows.Scan(&id, &version, &wallboxAutomatic)
		if err != nil {
			log.Fatal(err)
		}

		homeChargeStatus := HomeChargeStatus{
			Id:               id,
			Version:          version,
			WallboxAutomatic: wallboxAutomatic,
		}
		return homeChargeStatus, false
	}

	return HomeChargeStatus{}, true
}

func UpdateHomeChargeStatus(homeChargeStatus HomeChargeStatus) {
	db, err := sql.Open("sqlite3", "database/home-charge.db")
	if err != nil {
		log.Fatal("Open: ", err)
	}
	defer db.Close()

	// Daten in die Tabelle einfügen
	updateSQL := `
		UPDATE weHomeChargeStatus
		SET version=?, wallboxAutomatic=?
		WHERE id=?
	`
	_, err = db.Exec(updateSQL, homeChargeStatus.Version, homeChargeStatus.WallboxAutomatic, homeChargeStatus.Id)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("weHomeChargeStatus wurde erfolgreich aktualisiert.")
}

type BatteryCommand struct {
	Id   int
	Name string
}

func GetBatteryCommands() []BatteryCommand {
	db, err := sql.Open("sqlite3", "database/home-charge.db")
	if err != nil {
		log.Fatal("Open: ", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM weBatteryCommand")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	batteryCommands := []BatteryCommand{}
	for rows.Next() {
		var id int
		var name string

		err = rows.Scan(&id, &name)
		if err != nil {
			log.Fatal(err)
		}

		batteryCommands = append(batteryCommands, BatteryCommand{
			Id:   id,
			Name: name,
		})
	}

	return batteryCommands
}

type ScheduleCommand struct {
	Id                 int
	BatteryCommandId   int
	BatteryCommandName string
	TriggerType        string
	TriggerTime        time.Time
	TriggerSOC         int
	Triggered          bool
}

func GetScheduleCommands(triggered_optional ...bool) []ScheduleCommand {
	triggered := false
	if len(triggered_optional) > 0 {
		triggered = triggered_optional[0]
	}

	db, err := sql.Open("sqlite3", "database/home-charge.db")
	if err != nil {
		log.Fatal("Open: ", err)
	}
	defer db.Close()

	stm, err := db.Prepare("SELECT weScheduleCommand.*, weBatteryCommand.name as batteryCommandName FROM weScheduleCommand INNER JOIN weBatteryCommand ON weScheduleCommand.weBatteryCommand_id = weBatteryCommand.id WHERE weScheduleCommand.triggered = ? ")
	if err != nil {
		log.Fatal(err)
	}
	defer stm.Close()

	rows, err := stm.Query(triggered)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	scheduleCommands := []ScheduleCommand{}
	for rows.Next() {
		var id int
		var batteryCommandId int
		var batteryCommandName string
		var triggerType string
		var triggerTime time.Time
		var triggerSOC int
		var triggered bool

		err = rows.Scan(&id, &batteryCommandId, &triggerType, &triggerTime, &triggerSOC, &triggered, &batteryCommandName)
		if err != nil {
			log.Fatal(err)
		}

		scheduleCommands = append(scheduleCommands, ScheduleCommand{
			Id:                 id,
			BatteryCommandId:   batteryCommandId,
			BatteryCommandName: batteryCommandName,
			TriggerType:        triggerType,
			TriggerTime:        triggerTime,
			TriggerSOC:         triggerSOC,
			Triggered:          triggered,
		})
	}

	return scheduleCommands
}

func AddScheduleCommand(scheduleCommand ScheduleCommand) {
	db, err := sql.Open("sqlite3", "database/home-charge.db")
	if err != nil {
		log.Fatal("Open: ", err)
	}
	defer db.Close()

	// Daten in die Tabelle einfügen
	insertSQL := `
		INSERT INTO weScheduleCommand (weBatteryCommand_id, triggerType, triggerTime, triggerSOC, triggered)
		VALUES (?, ?, ?, ?, ?);
	`

	_, err = db.Exec(insertSQL, scheduleCommand.BatteryCommandId, scheduleCommand.TriggerType, scheduleCommand.TriggerTime, scheduleCommand.TriggerSOC, scheduleCommand.Triggered)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("weScheduleCommand wurde erfolgreich eingefügt.")
}

func UpdateScheduleCommand(scheduleCommand ScheduleCommand) {
	db, err := sql.Open("sqlite3", "database/home-charge.db")
	if err != nil {
		log.Fatal("Open: ", err)
	}
	defer db.Close()

	// Daten in die Tabelle einfügen
	updateSQL := `
		UPDATE weScheduleCommand
		SET weBatteryCommand_id=?, triggerType=?, triggerTime=?, triggerSOC=?, triggered=?
		WHERE id=?
	`
	_, err = db.Exec(updateSQL, scheduleCommand.BatteryCommandId, scheduleCommand.TriggerType, scheduleCommand.TriggerTime, scheduleCommand.TriggerSOC, scheduleCommand.Triggered, scheduleCommand.Id)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("weScheduleCommand wurde erfolgreich aktualisiert.")
}

func DeleteScheduleCommand(id int) {
	db, err := sql.Open("sqlite3", "database/home-charge.db")
	if err != nil {
		log.Fatal("Open: ", err)
	}
	defer db.Close()

	deleteSQL := "DELETE FROM weScheduleCommand WHERE id = ?;"
	_, err = db.Exec(deleteSQL, id)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("weScheduleCommand mit id ", id, " wurde erfolgreich gelöscht.")
}

func ParseTime(timeStr string) time.Time {
	loc, _ := time.LoadLocation("Europe/Berlin")
	//2023-11-05T02:00
	// Das Format, das Ihre Eingabe entspricht
	layout := "2006-01-02T15:04"
	// Zeitumwandlung
	t, err := time.ParseInLocation(layout, timeStr, loc)
	if err != nil {
		log.Fatal("Fehler beim Parsen der Zeit:", err)
	}
	return t
}

func LogBatteryCommand(scheduleCommand ScheduleCommand) {
	db, err := sql.Open("sqlite3", "database/home-charge.db")
	if err != nil {
		log.Fatal("Open: ", err)
	}
	defer db.Close()

	// Daten in die Tabelle einfügen
	insertSQL := `
		INSERT INTO weBatteryCommandLog (weBatteryCommand_id)
		VALUES (?);
	`

	_, err = db.Exec(insertSQL, scheduleCommand.BatteryCommandId)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("weBatteryCommandLog wurde erfolgreich erstellt.")
}
