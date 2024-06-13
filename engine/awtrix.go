package engine

import (
	"fmt"
	"strings"
	"time"

	"github.com/wimaha/home-charge/settings"
)

func Awtrix_doAll(env *settings.Environment) {
	doCurrentWattAndBattery(env)
	doCurrentConsumption(env)
	doTotalProduction(env)
}

/*func awtrixReset() {
	currentProduction_W = -1
	currentSoc = -1
	currentConsumption_W = -1
	currentTotalProduction_Wh = -1
}*/

var currentProduction_W = -1
var currentSoc = -1
var prodSocHidden = false
var production_W_lastNonZero = time.Now()
var currentSoc_lastNonZero = time.Now()

func doCurrentWattAndBattery(env *settings.Environment) {
	//Nur ausführen, wenn Awtrix und MqttClient initialisiert sind
	if env.Config.Awtrix == nil || env.MqttClient == nil {
		return
	}

	tempCurrentProduction_W, ok1 := env.Battery.ProductionW()
	tempCurrentSoc, ok2 := env.Battery.Soc()
	if ok1 && ok2 {
		if tempCurrentProduction_W != 0 {
			production_W_lastNonZero = time.Now()
		}
		if tempCurrentSoc != 0 {
			currentSoc_lastNonZero = time.Now()
		}

		before10Minutes := time.Now().Add(-10 * time.Minute)
		if !production_W_lastNonZero.Before(before10Minutes) {
			//Vollständigen View anzeigen, da innerhalb der letzten 10 Minuten > 0
			if tempCurrentProduction_W != currentProduction_W || tempCurrentSoc != currentSoc {
				currentProduction_W = tempCurrentProduction_W
				currentSoc = tempCurrentSoc

				var topic = fmt.Sprintf("%s/custom/ProductionW", env.Config.Awtrix.Prefix)
				var body = fmt.Sprintf(`{
					"pos" : 2,
					"color" : "FFFFFF",
					"progress" : %d,
					"icon" : "52730",
					"text" : "%d W",
					"textCase" : 2
				}`, currentSoc, currentProduction_W)
				env.MqttClient.Publish(topic, body)
				prodSocHidden = false
			}
			//fmt.Println("Vollständiger View")
		} else if !currentSoc_lastNonZero.Before(before10Minutes) {
			//Batterie View anzeigen
			if tempCurrentSoc != currentSoc {
				currentSoc = tempCurrentSoc

				var topic = fmt.Sprintf("%s/custom/ProductionW", env.Config.Awtrix.Prefix)
				var body = fmt.Sprintf(`{
					"pos" : 2,
					"color" : "FFFFFF",
					"icon" : "batteryfull",
					"text" : "%d %%",
					"textCase" : 2
				}`, currentSoc)
				env.MqttClient.Publish(topic, body)
				prodSocHidden = false
			}
			//fmt.Println("Batterie View")
		} else {
			//View ausblenden
			if !prodSocHidden {
				prodSocHidden = true
				var topic = fmt.Sprintf("%s/custom/ProductionW", env.Config.Awtrix.Prefix)
				env.MqttClient.Publish(topic, "")
			}
			//fmt.Println("Kein View")
		}
	} else {
		//TODO: Error View anzeigen
	}
}

var currentConsumption_W = -1

func doCurrentConsumption(env *settings.Environment) {
	//Nur ausführen, wenn Awtrix und MqttClient initialisiert sind
	if env.Config.Awtrix == nil || env.MqttClient == nil {
		return
	}

	if tempCurrentConsumption_W, ok := env.Battery.ConsumptionW(); ok {
		if tempCurrentConsumption_W != currentConsumption_W {
			currentConsumption_W = tempCurrentConsumption_W

			var consString string
			if currentConsumption_W >= 10000.0 {
				value := fmt.Sprintf("%.1f %s", float64(currentConsumption_W)/1000.0, "kW")
				consString = strings.Replace(value, ".", ",", -1)
			} else {
				consString = fmt.Sprintf("%d %s", currentConsumption_W, "W")
			}

			//fmt.Printf("doCurrentConsumption: %s \n", consString)

			var topic = fmt.Sprintf("%s/custom/ConsumptionW", env.Config.Awtrix.Prefix)
			var body = fmt.Sprintf(`{
				"color" : "FFFFFF",
				"icon" : "54064",
				"text" : "%s",
				"textCase" : 2
			}`, consString)
			env.MqttClient.Publish(topic, body)
		}
	} else {
		//TODO: Error View anzeigen
	}
}

var currentTotalProduction_Wh float64 = -1

func doTotalProduction(env *settings.Environment) {
	//Nur ausführen, wenn Awtrix, MqttClient und InfluxClient initialisiert sind
	if env.Config.Awtrix == nil || env.MqttClient == nil || env.InfluxClient == nil {
		return
	}

	tempTotalProduction, err := env.InfluxClient.InfluxProductionTotal()

	if err != nil {
		fmt.Printf("doTotalProduction error: %v", err)
	} else {
		if tempTotalProduction != currentTotalProduction_Wh {
			currentTotalProduction_Wh = tempTotalProduction

			var topic = fmt.Sprintf("%s/custom/TotalProductionWh", env.Config.Awtrix.Prefix)
			if currentTotalProduction_Wh < 1.0 {
				env.MqttClient.Publish(topic, "")
			} else {
				var text string
				if currentTotalProduction_Wh >= 10000.0 {
					text = fmt.Sprintf("%.1f kWh", currentTotalProduction_Wh/1000.0)
					text = strings.Replace(text, ".", ",", -1)
				} else if currentTotalProduction_Wh >= 1000.0 {
					text = fmt.Sprintf("%.2f kWh", currentTotalProduction_Wh/1000.0)
					text = strings.Replace(text, ".", ",", -1)
				} else {
					text = fmt.Sprintf("%.0f Wh", currentTotalProduction_Wh)
				}

				//fmt.Printf("doTotalProduction text: %s", text)

				var body = fmt.Sprintf(`{
					"gradient" : ["FFE91F","FF8F1F"],
					"text" : "%s",
					"textCase" : 2
				}`, text)
				env.MqttClient.Publish(topic, body)
			}
		}
	}
}
