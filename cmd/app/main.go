package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/yaml.v3"
)

type PricingHistory struct {
	MachineType   string
	RegionName    string
	HourSpotPrice float64
	HourPrice     float64
	UpdatedTS     int
	Updated       time.Time
}

func main() {

	database_path := flag.String("dbpath", "db.sqlite3", "Desired location of sqlite3 database")
	data_path := flag.String("data", "data/", "Location of pricing.yml history files")
	flag.Parse()
	db, err := sql.Open("sqlite3", *database_path)
	if err != nil {
		log.Fatalf("failed opening connection to sqlite: %v", err)
	}
	defer db.Close()

	initDatabase(context.Background(), db)

	files, err := os.ReadDir(*data_path)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		fmt.Printf("Processing file %s\n", file.Name())
		file_open, _ := os.ReadFile(fmt.Sprintf("%s/%s", *data_path, file.Name()))
		file_content := string(file_open)

		var data map[string]interface{}

		err := yaml.Unmarshal([]byte(file_content), &data)
		if err != nil {
			log.Fatal(err)
		}
		timestamp, timestamp_ok := getTimestamp(data)

		if timestamp_ok {
			instances := data["compute"].(map[string]interface{})["instance"].(map[string]interface{})
			for machineTypeName, instanceData := range instances {
				costData := instanceData.(map[string]interface{})["cost"].(map[string]interface{})
				for regionName, regionCostData := range costData {
					hour_spot_price := regionCostData.(map[string]interface{})["hour_spot"]
					hour_price := regionCostData.(map[string]interface{})["hour"]
					if hour_spot_price != nil && hour_price != nil {
						updated := convertTimestampToDate(timestamp).UTC()
						priceHistoryStruct := PricingHistory{
							MachineType:   machineTypeName,
							RegionName:    regionName,
							HourSpotPrice: hour_spot_price.(float64),
							HourPrice:     hour_price.(float64),
							UpdatedTS:     timestamp,
							Updated:       updated,
						}
						if _, err := createPriceHistoryEntry(context.Background(), db, &priceHistoryStruct); err != nil {
							log.Printf("failed to create price history entry: %v", err)
						}
					}
				}
			}
		}
	}
}

func initDatabase(ctx context.Context, client *sql.DB) {
	statement, err := client.Prepare("CREATE TABLE IF NOT EXISTS pricing_history (id INTEGER PRIMARY KEY, machine_type varchar(64), region_name varchar(64), hour_price REAL, spot_hour_price REAL, updated_ts INTEGER, updated varchar(64))")

	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}
	statement.Exec()

}
func createPriceHistoryEntry(ctx context.Context, client *sql.DB, priceHistory *PricingHistory) (bool, error) {
	statement, err := client.Prepare("INSERT OR IGNORE INTO pricing_history (machine_type, region_name, hour_price, spot_hour_price, updated_ts, updated) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatalf("failed to insert new record to pricing_history table: %v", err)
		return false, err
	}
	statement.Exec(priceHistory.MachineType, priceHistory.RegionName, priceHistory.HourPrice, priceHistory.HourSpotPrice, priceHistory.UpdatedTS, priceHistory.Updated)
	return true, nil
}

func getTimestamp(data map[string]interface{}) (int, bool) {
	if about, ok := data["about"]; ok {
		if aboutMap, ok := about.(map[string]interface{}); ok {
			if timestamp, ok := aboutMap["timestamp"]; ok {
				return timestamp.(int), true
			}
		}
	}
	return 0, false
}

// int timestamp as input, returns time.Time
func convertTimestampToDate(timestamp int) time.Time {
	return time.Unix(int64(timestamp), 0)
}
