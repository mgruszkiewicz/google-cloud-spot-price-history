package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
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

type MachineType struct {
	Family      string
	MachineType string
	CpuCores    int
	MemoryGB    float64
}

func main() {
	database_path := flag.String("dbpath", "db.sqlite3", "Desired location of sqlite3 database")
	data_path := flag.String("data", "data/", "Location of pricing.yml history files")
	batch_size := flag.Int("batch", 2000, "Batch size for database inserts")
	flag.Parse()
	db, err := sql.Open("sqlite3", *database_path)
	if err != nil {
		log.Fatalf("failed opening connection to sqlite: %v", err)
	}
	defer db.Close()

	// Optimize SQLite settings for bulk inserts
	if _, err := db.Exec("PRAGMA synchronous = OFF"); err != nil {
		log.Printf("Failed to set synchronous pragma: %v", err)
	}
	if _, err := db.Exec("PRAGMA journal_mode = MEMORY"); err != nil {
		log.Printf("Failed to set journal_mode pragma: %v", err)
	}

	initDatabase(context.Background(), db)

	files, err := os.ReadDir(*data_path)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		start := time.Now()
		fmt.Printf("Processing file %s\n", file.Name())

		if err := processFile(*data_path, file.Name(), db, *batch_size); err != nil {
			log.Printf("Error processing file %s: %v", file.Name(), err)
			continue
		}

		fmt.Printf("Completed %s in %v\n", file.Name(), time.Since(start))
	}

}

func initDatabase(ctx context.Context, client *sql.DB) {
	statement, err := client.Prepare(`CREATE TABLE IF NOT EXISTS pricing_history (
		id INTEGER PRIMARY KEY, 
		machine_type varchar(64), 
		region_name varchar(64), 
		hour_price REAL, 
		spot_hour_price REAL, 
		updated_ts INTEGER, 
		updated varchar(64),
		UNIQUE(machine_type, region_name, updated_ts)
	)`)

	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}
	defer statement.Close()
	if _, err := statement.Exec(); err != nil {
		log.Fatalf("Failed to execute create table: %v", err)
	}

	statement, err = client.Prepare(`CREATE TABLE IF NOT EXISTS machine_type (
		id INTEGER PRIMARY KEY,
		family varchar(64),  
		machine_type varchar(64), 
		cpu_cores INTEGER, 
		memory_gb REAL, 
		UNIQUE(family, machine_type, cpu_cores, memory)
	)`)

	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}
	defer statement.Close()
	if _, err := statement.Exec(); err != nil {
		log.Fatalf("Failed to execute create table: %v", err)
	}

	// Create index for better query performance
	if _, err := client.Exec("CREATE INDEX IF NOT EXISTS idx_machine_region ON pricing_history(machine_type, region_name)"); err != nil {
		log.Printf("Failed to create index: %v", err)
	}

}

func processFile(dataPath, fileName string, db *sql.DB, batchSize int) error {
	fileData, err := os.ReadFile(fmt.Sprintf("%s/%s", dataPath, fileName))
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var data map[string]interface{}
	if err := yaml.Unmarshal(fileData, &data); err != nil {
		return fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	timestamp, timestampOk := getTimestamp(data)
	if !timestampOk {
		return fmt.Errorf("no valid timestamp found")
	}

	// Extract and validate structure once
	compute, ok := data["compute"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid compute structure")
	}

	instances, ok := compute["instance"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid instance structure")
	}

	// Collect all records first
	var records []PricingHistory
	var machine_types []MachineType
	updated := convertTimestampToDate(timestamp).UTC()

	for machineTypeName, instanceData := range instances {

		instanceMap, ok := instanceData.(map[string]interface{})
		if !ok {
			continue
		}
		fmt.Printf("%s have %d cpus and %d memory\n", machineTypeName, instanceMap["cpu"], instanceMap["ram"])

		costData, ok := instanceMap["cost"].(map[string]interface{})
		if !ok {
			continue
		}

		for regionName, regionCostData := range costData {
			regionMap, ok := regionCostData.(map[string]interface{})
			if !ok {
				continue
			}

			hourSpotPrice, spotOk := regionMap["hour_spot"].(float64)
			hourPrice, priceOk := regionMap["hour"].(float64)

			if spotOk && priceOk {
				machine_types = append(machine_types, MachineType{
					Family:      strings.SplitAfter(machineTypeName, "-")[0],
					MachineType: machineTypeName,
					CpuCores:    instanceMap["cpu"].(int),
					MemoryGB:    instanceMap["ram"].(float64),
				})
				records = append(records, PricingHistory{
					MachineType:   machineTypeName,
					RegionName:    regionName,
					HourSpotPrice: hourSpotPrice,
					HourPrice:     hourPrice,
					UpdatedTS:     timestamp,
					Updated:       updated,
				})
			}
		}
	}

	fmt.Printf("Found %d records to insert\n", len(records))

	// Insert in batches with transactions
	return insertRecordsInBatches(db, records, batchSize)
}

func insertRecordsInBatches(db *sql.DB, records []PricingHistory, batchSize int) error {
	for i := 0; i < len(records); i += batchSize {
		end := i + batchSize
		if end > len(records) {
			end = len(records)
		}

		// Begin transaction for this batch
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		stmt, err := tx.Prepare("INSERT OR IGNORE INTO pricing_history (machine_type, region_name, hour_price, spot_hour_price, updated_ts, updated) VALUES (?, ?, ?, ?, ?, ?)")
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to prepare statement: %w", err)
		}

		for j := i; j < end; j++ {
			record := records[j]
			if _, err := stmt.Exec(
				record.MachineType,
				record.RegionName,
				record.HourPrice,
				record.HourSpotPrice,
				record.UpdatedTS,
				record.Updated,
			); err != nil {
				stmt.Close()
				tx.Rollback()
				return fmt.Errorf("failed to insert record: %w", err)
			}
		}

		stmt.Close()
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		fmt.Printf("Inserted batch of %d records\n", end-i)
	}

	return nil
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
