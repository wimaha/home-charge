package wallbox

import (
	"errors"
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

func (w *Mennekes) readModbusRegister32(register uint16) (uint32, error) {
	// now that the client is created and configured, attempt to connect
	err := w.Client.Open()
	if err != nil {
		log.Println("Modbus-Client konnte nicht geöffnet werden.")
		return 0, err
	}
	defer w.Client.Close()

	var reg32 uint32
	reg32, err = w.Client.ReadUint32(register, modbus.HOLDING_REGISTER)
	if err != nil {
		// error out
		//log.Printf("Modbus error: %v\n", err)
		return 0, err
	} else {
		// use value
		//fmt.Printf("value: %v\n", reg16)        // as unsigned integer
		//fmt.Printf("value: %v\n", int16(reg16)) // as signed integer
		return reg32, nil
	}
}

func (w *Mennekes) readModbusRegister32s(register uint16, count uint16) ([]uint32, error) {
	var reg32s []uint32

	// now that the client is created and configured, attempt to connect
	err := w.Client.Open()
	if err != nil {
		log.Println("Modbus-Client konnte nicht geöffnet werden.")
		return reg32s, err
	}
	defer w.Client.Close()

	// read the same count*3 consecutive 16-bit input registers as count 32-bit integers
	reg32s, err = w.Client.ReadUint32s(register, count, modbus.HOLDING_REGISTER)

	return reg32s, err
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

func (w *Mennekes) ChargingDuration() (int, error) {
	duration, err := w.readModbusRegister32(718)
	if err != nil {
		return 0, err
	}
	return int(duration), nil
}

func (w *Mennekes) getPhaseValues(register uint16) (int, int, int, error) {
	// read the same 4 consecutive 16-bit input registers as 2 32-bit integers
	var reg32s []uint32
	var err error
	reg32s, err = w.readModbusRegister32s(register, 3)

	if err != nil {
		return 0, 0, 0, err
	} else if len(reg32s) == 3 {
		return int(reg32s[0]), int(reg32s[1]), int(reg32s[2]), nil
	} else {
		return 0, 0, 0, errors.New("falsche Anzahl Register")
	}
}

func (w *Mennekes) Current() (float64, float64, float64, error) {
	c1, c2, c3, err := w.getPhaseValues(212)
	if err != nil {
		return 0, 0, 0, err
	} else {
		return float64(c1) / 1000.0, float64(c2) / 1000.0, float64(c3) / 1000.0, nil
	}
}

func (w *Mennekes) Voltage() (float64, float64, float64, error) {
	c1, c2, c3, err := w.getPhaseValues(222)
	if err != nil {
		return 0, 0, 0, err
	} else {
		return float64(c1), float64(c2), float64(c3), nil
	}
}

func (w *Mennekes) SessionEnergyWh() (int, error) {
	chargedEnergy, err := w.readModbusRegister32(716)
	if err != nil {
		return 0, err
	}
	return int(chargedEnergy), nil
}
