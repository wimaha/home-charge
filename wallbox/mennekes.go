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

func (w *Mennekes) readModbusRegister(register uint16) (uint16, error) {
	// now that the client is created and configured, attempt to connect
	err := w.Client.Open()
	if err != nil {
		log.Println("Modbus-Client konnte nicht geöffnet werden.")
		return 0, err
	}
	defer w.Client.Close()

	var reg16 uint16
	reg16, err = w.Client.ReadRegister(register, modbus.HOLDING_REGISTER)
	if err != nil {
		// error out
		//log.Printf("Modbus error: %v\n", err)
		return 0, err
	} else {
		// use value
		//fmt.Printf("value: %v\n", reg16)        // as unsigned integer
		//fmt.Printf("value: %v\n", int16(reg16)) // as signed integer
		return reg16, nil
	}
}

/*
A = 1,
B = 2,
C = 3,
D = 4,
E = 5
*/
func (w *Mennekes) Status() (ChargeStatus, error) {
	status, err := w.readModbusRegister(122)
	if err != nil {
		return StatusNone, err
	}

	switch status {
	case 1:
		return StatusAvailable, nil
	case 2:
		return StatusOccupied, nil
	case 3, 4:
		return StatusCharging, nil
	default:
		return StatusNone, fmt.Errorf("unbekannter Status: %d", status)
	}
}

func (w *Mennekes) StatusText() string {
	status, err := w.Status()
	if err != nil {
		return "Fehler"
	}

	return statusTextWithStatus(status)
}

func (w *Mennekes) StatusAndText() (ChargeStatus, string) {
	status, err := w.Status()
	if err != nil {
		return StatusNone, "Fehler"
	}
	text := statusTextWithStatus(status)
	return status, text
}
