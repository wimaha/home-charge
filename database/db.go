package database

import (
	"database/sql"
	"fmt"
	"log"

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
