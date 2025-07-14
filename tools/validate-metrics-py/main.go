package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"io"
	"log"
	"os"
	"strings"
)

type metricMapping struct {
	dd       string
	asserted bool
}

func main() {

	// Handle command line arguments
	masterFile := flag.String("master", "../../data/master.csv", "path to master.csv")
	metricsPyFile := flag.String("code", "../../integrations-extras/redpanda/datadog_checks/redpanda/metrics.py", "path to metrics.py")
	flag.Parse()

	metricsPyMappings := make(map[string]*metricMapping)

	// Open metrics.py
	mp, err := os.Open(*metricsPyFile)
	if err != nil {
		log.Fatalf("Could not open master: %v", err)
	}
	defer mp.Close()

	scanner := bufio.NewScanner(mp)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "'")
		n := len(parts)
		if n == 5 {
			metricsPyMappings[parts[1]] = &metricMapping{dd: parts[3], asserted: false}
		}
	}

	// Check for errors during scanning
	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading master: %v\n", err)
	}

	// Open the master CSV master
	master, err := os.Open(*masterFile)
	if err != nil {
		log.Fatalf("Could not open master: %v", err)
	}
	defer master.Close()

	// Create a new CSV reader
	reader := csv.NewReader(master)

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
			log.Fatalf("Error reading masterRecord: %v", err)
		}

		masterRP := masterRecord[0]
		if masterRecord[2] == "count" && strings.HasSuffix(masterRecord[0], "_total") {
			masterRP = masterRP[0 : len(masterRP)-6]
		}
		masterDD := masterRecord[1]

		metricsPyMapping, ok := metricsPyMappings[masterRP]
		if !ok {
			log.Fatalf("Could not find metric %s in metrics.py", masterRP)
		} else {
			namespacedDDFromMetricsPy := "redpanda." + metricsPyMapping.dd
			if namespacedDDFromMetricsPy != masterDD {
				log.Fatalf("master dd metric name %s doesn't match %s in metrics.py", masterDD, metricsPyMapping.dd)
			} else {
				if metricsPyMapping.asserted {
					log.Fatalf("master dd metric name %s already asserted", masterDD)
				}
				metricsPyMapping.asserted = true
			}
		}

	}

	for rp, mapping := range metricsPyMappings {
		if !mapping.asserted {
			log.Fatalf("metrics.py contains a metricsPyMappings %s : %s not known in master.csv", rp, mapping.dd)
		}
	}
}
