package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"

	_ "github.com/go-sql-driver/mysql"
)

type Setting struct {
	DSN       string `json:"dsn"`
	StatSQL   string `json:"stat_sql"`
	InsertSQL string `json:"insert_sql"`
	Freq      int    `json:"freq"`
}

func main() {
	setting := readSettingFromNacos()

	db, err := sql.Open("mysql", setting.DSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	for {
		var count int
		err := db.QueryRow(setting.StatSQL).Scan(&count)
		if err != nil {
			log.Println("Error executing stat SQL:", err)
			continue
		}

		_, err = db.Exec(setting.InsertSQL, count)
		if err != nil {
			log.Println("Error inserting into stats table:", err)
			continue
		}

		fmt.Printf("Inserted count %d into stats table.\n", count)

		time.Sleep(time.Duration(setting.Freq) * time.Second)
	}
}

func readSettingFromNacos() Setting {
	serverConfigs := []constant.ServerConfig{
		{
			IpAddr: "127.0.0.1",
			Port:   8848,
		},
	}

	clientConfig := constant.ClientConfig{
		NamespaceId: "public",
		TimeoutMs:   5000,
		LogLevel:    "info",
	}

	client, err := clients.CreateConfigClient(map[string]interface{}{
		"serverConfigs": serverConfigs,
		"clientConfig":  clientConfig,
	})
	if err != nil {
		log.Fatalf("Error creating Nacos client: %v", err)
	}

	config, err := client.GetConfig(vo.ConfigParam{
		DataId: "yourDataId",
		Group:  "DEFAULT_GROUP",
	})
	if err != nil {
		log.Fatalf("Error getting config from Nacos: %v", err)
	}

	var setting Setting
	err = json.Unmarshal([]byte(config), &setting)
	if err != nil {
		log.Fatalf("Error unmarshalling config: %v", err)
	}

	return setting
}
