package sonnenbatterie

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type Sonnenbatterie struct {
	API_TOKEN           string
	BATTERY_IP          string
	StatusArray         map[string]interface{}
	OperationModeStored int
}

func NewSonnenbatterie(apiToken string, batteryIP string) *Sonnenbatterie {
	return &Sonnenbatterie{
		API_TOKEN:           apiToken,
		BATTERY_IP:          batteryIP,
		OperationModeStored: -1,
	}
}

func (s *Sonnenbatterie) Soc() string {
	status := s.status()
	return fmt.Sprintf("%.0f %%", status["USOC"].(float64))
}

func (s *Sonnenbatterie) BatteryCharging() string {
	status := s.status()
	batteryCharging := status["BatteryCharging"].(bool)
	batteryDischarging := status["BatteryDischarging"].(bool)

	if batteryCharging {
		return "lädt"
	} else if batteryDischarging {
		return "entlädt"
	} else {
		return "neutral"
	}
}

func (s *Sonnenbatterie) PacTotalW() string {
	status := s.status()
	return fmt.Sprintf("%.0f W", status["Pac_total_W"].(float64))
}

func (s *Sonnenbatterie) status() map[string]interface{} {
	if s.StatusArray != nil {
		return s.StatusArray
	}
	status, err := s.getStatus()
	if err != nil {
		fmt.Println("Error: ", err)
		return nil
	}
	s.StatusArray = status
	return s.StatusArray
}

func (s *Sonnenbatterie) getStatus() (map[string]interface{}, error) {
	url := fmt.Sprintf("http://%s/api/v2/status", s.BATTERY_IP)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Auth-Token", s.API_TOKEN)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP Request failed with status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var responseData map[string]interface{}
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		return nil, err
	}

	return responseData, nil
}

func (s *Sonnenbatterie) OperationMode() int {
	if s.OperationModeStored != -1 {
		return s.OperationModeStored
	}

	data := s.GetConfiguration("EM_OperatingMode")

	operationMode, err := strconv.ParseInt(data["EM_OperatingMode"].(string), 10, 64)
	if err != nil {
		return -1
	}
	s.OperationModeStored = int(operationMode)
	return int(operationMode)
}

func (s *Sonnenbatterie) OperationModeText() string {
	operationMode := s.OperationMode()
	operationModeText := "Unknown"

	switch operationMode {
	case 10:
		operationModeText = "Time-of-use"
	case 2:
		operationModeText = "Automatic"
	case 1:
		operationModeText = "Manual"
	}

	return operationModeText
}

func (s *Sonnenbatterie) GetConfiguration(config string) map[string]interface{} {
	url := fmt.Sprintf("http://%s/api/v2/configurations/%s", s.BATTERY_IP, config)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("cURL Error:", err)
		return nil
	}

	req.Header.Set("Auth-Token", s.API_TOKEN)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("cURL Error:", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("HTTP Request failed with status code %d\n", resp.StatusCode)
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return nil
	}

	var responseData map[string]interface{}
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		fmt.Println("Unable to decode JSON response:", err)
		return nil
	}

	return responseData
}

func (s *Sonnenbatterie) SetOperationMode(mode int) {
	s.PutConfiguration("EM_OperatingMode=2")
}

func (s *Sonnenbatterie) PutConfiguration(data string) {
	url := fmt.Sprintf("http://%s/api/v2/configurations", s.BATTERY_IP)
	method := "PUT"
	payload := []byte(data)

	req, err := http.NewRequest(method, url, bytes.NewBuffer(payload))
	if err != nil {
		panic(err)
	}

	req.Header.Set("Auth-Token", s.API_TOKEN)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("HTTP Request failed with status code %d\n", resp.StatusCode)
		return
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	fmt.Println(string(responseBody))
}

func (s *Sonnenbatterie) ChargeBattery() {
	s.PostCharge(3400)
}

func (s *Sonnenbatterie) StopChargeBattery() {
	s.PostCharge(0)
}

func (s *Sonnenbatterie) StopDischargeBattery() {
	s.PostCharge(0)
}

func (s *Sonnenbatterie) PostCharge(watt int) {
	url := fmt.Sprintf("http://%s/api/v2/setpoint/charge/%d", s.BATTERY_IP, watt)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		fmt.Println("cURL Error:", err)
		return
	}

	req.Header.Set("Auth-Token", s.API_TOKEN)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("cURL Error:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("HTTP Request failed with status code %d\n", resp.StatusCode)
		return
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	fmt.Println(string(responseBody))
}
