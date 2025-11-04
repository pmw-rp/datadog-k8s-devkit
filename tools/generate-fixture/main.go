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

	// Read the rest of the records
	for {
		masterRecord, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Error reading record: %v", err)
		}

		// Begin validation excludes

		// Fixing typo in previous releases (<=2.1.0)
		if masterRecord[1] == "redpanda.schema_registry_latency_seconds" {
			continue
		}

        // Fixing wrongly named metric in previous releases (<=2.1.0)
		if masterRecord[1] == "redpanda.cluster.controller_log_limit_requests_dropped" {
			continue
		}

        // Non-existent metric (still referenced in metadata.csv)
    	if masterRecord[0] == "redpanda_cluster_replicas" {
			continue
		}

		// End validation excludes

		fmt.Printf("# HELP %s %s\n", masterRecord[0], masterRecord[6])
		if masterRecord[2] == "gauge" {
			fmt.Printf("# TYPE %s gauge\n", masterRecord[0])
			fmt.Printf("%s{} 0\n", masterRecord[0])
		}
		if masterRecord[2] == "count" {
			fmt.Printf("# TYPE %s counter\n", masterRecord[0])
			fmt.Printf("%s{} 0\n", masterRecord[0])
		}
		if masterRecord[2] == "histogram" {
			fmt.Printf("# TYPE %s histogram\n", masterRecord[0])
			//redpanda_schema_registry_request_latency_seconds_bucket{instance="10.0.1.180:9644",le="0.000255"} 0
			fmt.Printf("%s_bucket{le=\"0.1\"} 0\n", masterRecord[0])
			fmt.Printf("%s_bucket{le=\"0.2\"} 1\n", masterRecord[0])
			fmt.Printf("%s_bucket{le=\"+Inf\"} 2\n", masterRecord[0])
			fmt.Printf("%s_count{} 3\n", masterRecord[0])
			fmt.Printf("%s_sum{} 20\n", masterRecord[0])
		}

	}
}
