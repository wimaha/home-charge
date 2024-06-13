package settings

import (
	"fmt"
	"log"
	"os"

	"github.com/wimaha/home-charge/battery"
	"github.com/wimaha/home-charge/database"
	"github.com/wimaha/home-charge/mqtt"
	"github.com/wimaha/home-charge/wallbox"
	yaml "gopkg.in/yaml.v3"
)

type Conf struct {
	Sonnenbatterie struct {
		ApiToken string `yaml:"apiToken"`
		Host     string `yaml:"host"`
	} `yaml:"sonnenbatterie"`
	Wallbox *struct {
		Type string `yaml:"type"`
		Host string `yaml:"host"`
	} `yaml:"wallbox,omitempty"`
	Mqtt *struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		ClientId string `yaml:"clientId"`
	} `yaml:"mqtt,omitempty"`
	Awtrix *struct {
		Prefix string `yaml:"prefix"`
	} `yaml:"awtrix,omitempty"`
	InfluxDB *struct {
		Host         string `yaml:"host"`
		Port         string `yaml:"port"`
		Token        string `yaml:"token"`
		Organisation string `yaml:"organisation"`
		Querys       struct {
			ProductionTotal string `yaml:"productionTotal"`
		} `yaml:"querys"`
	} `yaml:"influxdb,omitempty"`
	Test *struct {
		Test string `yaml:"test"`
	} `yaml:"test,omitempty"`
}

func (c *Conf) GetConf() *Conf {
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

func (c *Conf) CheckConf(print bool) bool {
	ret := true

	if c.Sonnenbatterie.ApiToken != "" && c.Sonnenbatterie.Host != "" {
		log.Println("Sonnenbatterie ✅")
	} else {
		log.Println("Sonnenbatterie ❌")
		ret = false
	}
	//log.Printf("	ApiToken: %s", c.Sonnenbatterie.ApiToken)
	//log.Printf("	Host: %s", c.Sonnenbatterie.Host)

	if c.Wallbox != nil {
		if c.Wallbox.Type == "mennekes" && c.Wallbox.Host != "" {
			log.Println("Wallbox ✅")
		} else {
			log.Println("Wallbox ❌")
			ret = false
		}
		//log.Printf("	Type: %s", c.Wallbox.Type)
		//log.Printf("	Host: %s", c.Wallbox.Host)
	}

	if c.Mqtt != nil {
		if c.Mqtt.Port != "" && c.Mqtt.Host != "" {
			log.Println("Mqtt ✅")
		} else {
			log.Println("Mqtt ❌")
			ret = false
		}
	}

	if c.Awtrix != nil {
		if c.Awtrix.Prefix != "" {
			log.Println("Awtrix ✅")
		} else {
			log.Println("Awtrix ❌")
			ret = false
		}
	}

	if c.InfluxDB != nil {
		if c.InfluxDB.Port != "" && c.InfluxDB.Host != "" && c.InfluxDB.Token != "" && c.InfluxDB.Organisation != "" && c.InfluxDB.Querys.ProductionTotal != "" {
			log.Println("InfluxDB ✅")
		} else {
			log.Println("InfluxDB ❌")
			ret = false
		}
	}

	return ret
}

type Environment struct {
	Config          *Conf
	Battery         *battery.Sonnenbatterie
	MqttClient      *mqtt.MqttClient
	InfluxClient    *database.InfluxClient
	WallboxInstance *wallbox.Mennekes
}
