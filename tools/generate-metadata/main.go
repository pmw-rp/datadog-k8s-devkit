package main

import (
	"encoding/csv"
	"flag"
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
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	header, err := reader.Read()
	if err != nil {
		log.Fatalf("Could not read header: %v", err)
	}
	_ = header

	newHeader := []string{"metric_name", "metric_type", "interval", "unit_name", "per_unit_name", "description", "orientation", "integration", "short_name", "curated_metric"}
	writer.Write(newHeader)

	// Read the rest of the records
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Error reading record: %v", err)
		}
		writer.Write(record[1:10])
	}
}
