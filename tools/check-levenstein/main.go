package main

import (
	"encoding/csv"
	"fmt"
	"flag"
	"github.com/hbollon/go-edlib"
	"io"
	"log"
	"os"
)

func main() {

	// Handle command line arguments
	masterFile := flag.String("master", "../../data/master.csv", "path to master.csv master")
	flag.Parse()

	// Open the CSV file
	file, err := os.Open(*masterFile)
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
// 	fmt.Printf("Header: %v\n", header)

	// Read the rest of the records
	for {
// 		position := reader.InputOffset()
// 		fmt.Printf("%d\n", position)
// 		line, column := reader.FieldPos(0)
// 		fmt.Printf("%d,%d\n", line, column)
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Error reading record: %v", err)
		}
		//fmt.Printf("Record: %v\n", record)
		res := edlib.LevenshteinDistance(record[0], record[1])
		fmt.Printf("%d,%s,%s\n", res, record[0], record[1])
	}
}
