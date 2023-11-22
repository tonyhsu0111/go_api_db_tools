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

var currentSetting Setting

func readSettingFromNacos() Setting {
	// Nacos 服務器配置
	serverConfigs := []constant.ServerConfig{
		{
			IpAddr: "192.168.10.26",
			Port:   8848,
		},
	}
	// Nacos 客戶端配置
	clientConfig := constant.ClientConfig{
		NamespaceId: "sea_dev",
		Username:    "sea_dev",
		Password:    "1qaz2wsx",
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
		DataId: "dev_searebot",
		Group:  "sea_group",
		// DataId: "yourDataId",
		// Group:  "DEFAULT_GROUP",
	})
	if err != nil {
		log.Fatalf("Error getting config from Nacos: %v", err)
	}

	client.ListenConfig(vo.ConfigParam{
		DataId: "dev_searebot",
		Group:  "sea_group",
		// DataId: "yourDataId",
		// Group:  "DEFAULT_GROUP",
		OnChange: func(namespace, group, dataId, data string) {
			fmt.Println("配置文件發生了變化...")
			handleConfigChange(data)
		},
	})
	var setting Setting
	err = json.Unmarshal([]byte(config), &setting)
	if err != nil {
		log.Fatalf("Error unmarshalling config: %v", err)
	}
	return setting
}

func handleConfigChange(newConfig string) {
	var setting Setting
	err := json.Unmarshal([]byte(newConfig), &setting)
	if err != nil {
		log.Printf("Error unmarshalling new config: %v", err)
		return
	}
	// 更新當前設定
	currentSetting = setting
	log.Println("config data:", currentSetting)
	// 根據需要重啟服務或重新初始化資源
	// restartServices(setting)
}

func runMainLogic() {
	for {
		// 使用 currentSetting 進行業務邏輯
		// ...
		db, err := sql.Open("mysql", currentSetting.DSN)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		for {
			var count int
			err := db.QueryRow(currentSetting.StatSQL).Scan(&count)
			if err != nil {
				log.Println("Error executing stat SQL:", err)
				continue
			}

			_, err = db.Exec(currentSetting.InsertSQL, count)
			if err != nil {
				log.Println("Error inserting into stats table:", err)
				continue
			}

			log.Printf("Inserted count %d into stats table.\n", count)
			time.Sleep(time.Duration(currentSetting.Freq) * time.Second)
		}

	}
}

func main() {
	// 初始化 Nacos 連接
	setting := readSettingFromNacos()

	// 設置當前配置
	currentSetting = setting

	// 開始執行主要邏輯
	go runMainLogic()

	// 持續運行，等待配置更新
	select {}
}
