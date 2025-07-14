package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"
	"strings"
)

func main() {

	// Handle command line arguments
	masterFile := flag.String("master", "../../data/master.csv", "path to master.csv master")
	metricsPyFile := flag.String("code", "../../integrations-extras/redpanda/tests/common.py", "path to metrics.py master")
	flag.Parse()

	commmonPyMappings := make(map[string]bool)

	// Open metrics.py
	mp, err := os.Open(*metricsPyFile)
	if err != nil {
		log.Fatalf("Could not open master: %v", err)
	}
	defer mp.Close()

	scanner := bufio.NewScanner(mp)

	sb := strings.Builder{}
	inMap := false
	for scanner.Scan() {
		line := scanner.Text()

		if line == "INSTANCE_METRIC_GROUP_MAP = {" {
			inMap = true
			sb.WriteString("{")
			continue
		}
		if line == "}" {
			inMap = false
			sb.WriteString("}")
			continue
		}
		if inMap {
			trimmed := strings.TrimSpace(line)

			sb.WriteString(trimmed)
			continue
		}
	}
	j := sb.String()
	j = strings.Replace(j, "'", "\"", -1)
	j = strings.Replace(j, ",]", "]", -1)

	m := make(map[string][]string)
	err = json.Unmarshal([]byte(j), &m)
	if err != nil {
		log.Fatalf("Could not parse map from common.py: %v", err)
	}

	for _, list := range m {
		for _, item := range list {
			commmonPyMappings[item] = false
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

	validate := func(metric string) {
		_, ok := commmonPyMappings[metric]
		if !ok {
			log.Fatalf("Could not find dd metric %s in common.py", metric)
		}
		commmonPyMappings[metric] = true
	}

	// Read the records
	for {
		masterRecord, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Error reading masterRecord: %v", err)
		}

		//lookups := make([]string,0)
		dd := masterRecord[1]

		if masterRecord[2] == "gauge" {
			validate(dd)
		}
		if masterRecord[2] == "count" {
			validate(dd + ".count")
		}
		if masterRecord[2] == "histogram" {
			validate(dd + ".bucket")
			validate(dd + ".count")
			validate(dd + ".sum")
		}
	}

	for metric, asserted := range commmonPyMappings {
		if !asserted {
			log.Fatalf("master.csv doesn't contain a dd metric %s that is referenced in common.py", metric)
		}
	}
}
