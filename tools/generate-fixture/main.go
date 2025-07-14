package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

func main() {

	// Handle command line arguments
	inputFile := flag.String("input", "../../master.csv", "path to master.csv file")
	flag.Parse()

	// Open the CSV file
	file, err := os.Open(*inputFile)
	if err != nil {
		log.Fatalf("Could not open file: %v", err)
	}
	defer file.Close()

	// Create a new CSV reader
	reader := csv.NewReader(file)

	// Optionally read the header
	header, err := reader.Read()
	if err != nil {
		log.Fatalf("Could not read header: %v", err)
	}
	_ = header
	//fmt.Printf("Header: %v\n", header)

	// Read the rest of the records
	for {
		//position := reader.InputOffset()
		//fmt.Printf("%d\n", position)
		//line, column := reader.FieldPos(0)
		//fmt.Printf("%d,%d\n", line, column)
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Error reading record: %v", err)
		}
		fmt.Printf("# HELP %s %s\n", record[0], record[6])
		if record[2] == "gauge" {
			fmt.Printf("# TYPE %s gauge\n", record[0])
			fmt.Printf("%s{} 0\n", record[0])
		}
		if record[2] == "count" {
			fmt.Printf("# TYPE %s counter\n", record[0])
			fmt.Printf("%s{} 0\n", record[0])
		}
		if record[2] == "histogram" {
			fmt.Printf("# TYPE %s histogram\n", record[0])
			//redpanda_schema_registry_request_latency_seconds_bucket{instance="10.0.1.180:9644",le="0.000255"} 0
			fmt.Printf("%s_bucket{le=\"0.1\"} 0\n", record[0])
			fmt.Printf("%s_bucket{le=\"+Inf\"} 0\n", record[0])
			fmt.Printf("%s_count{} 0\n", record[0])
			fmt.Printf("%s_sum{} 0\n", record[0])
		}

	}
}
