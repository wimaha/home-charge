package wallbox

import (
	"fmt"
	"log"
	"time"

	"github.com/simonvetter/modbus"
)

type Mennekes struct {
	WALLBOX_IP string
	Client     *modbus.ModbusClient
}

func NewMennekes(wallboxIP string) *Mennekes {
	return &Mennekes{
		WALLBOX_IP: wallboxIP,
		Client:     createClient(wallboxIP),
	}
}

func createClient(wallboxIP string) *modbus.ModbusClient {
	url := fmt.Sprintf("tcp://%s:502", wallboxIP)
	client, err := modbus.NewClient(&modbus.ClientConfiguration{
		URL:     url,
		Timeout: 1 * time.Second,
	})
	if err != nil {
		log.Println("Modbus-Client konnte nicht erstellt werden.")
	}
	return client
}

func (w *Mennekes) readModbusRegister(register uint16) uint16 {
	// now that the client is created and configured, attempt to connect
	err := w.Client.Open()
	if err != nil {
		log.Println("Modbus-Client konnte nicht geöffnet werden.")
		return 0
	}
	defer w.Client.Close()

	var reg16 uint16
	reg16, _ = w.Client.ReadRegister(register, modbus.HOLDING_REGISTER)
	/*if err != nil {
		// error out
	} else {
		// use value
		//fmt.Printf("value: %v\n", reg16)        // as unsigned integer
		//fmt.Printf("value: %v\n", int16(reg16)) // as signed integer
	}*/

	return reg16
}

/*
0 = Available
1 = Occupied
2 = Reserved
3 = Unavailable
4 = Faulted
5 = Preparing
6 = Charging
7 = Suspend-
edEVSE
8 = SuspendedEV
9 = Finishing
*/
func (w *Mennekes) Status() int {
	return int(w.readModbusRegister(104))
}

func (w *Mennekes) StatusText() string {
	status := w.Status()
	return w.StatusTextWithStatus(status)
}

func (w *Mennekes) StatusTextWithStatus(status int) string {
	switch status {
	case 0:
		return "Verfügbar"
	case 1:
		return "Belegt"
	case 2:
		return "Reserviert"
	case 3:
		return "Nicht verfügbar"
	case 4:
		return "Gestört"
	case 5:
		return "Vorbereiten"
	case 6:
		return "Laden"
	case 7:
		return "Unterbrochen durch Wallbox"
	case 8:
		return "Unterbrochen durch Fahrzeug"
	case 9:
		return "Beenden"
	}
	return "Unbekannt"
}

func (w *Mennekes) StatusAndText() (int, string) {
	status := w.Status()
	text := w.StatusTextWithStatus(status)
	return status, text
}
