package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"gopkg.in/yaml.v3"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
)

type Config struct {
	Url          string   `yaml:"url"`
	Regex        string   `yaml:"regex"`
	Excludes     []string `yaml:"excludes"`
	excludesMap  map[string]bool
	Replacements map[string]string `yaml:"replacements"`
}

func responseToLines(resp *http.Response) ([]string, error) {
	defer resp.Body.Close()

	var lines []string
	scanner := bufio.NewScanner(resp.Body)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

func getPublicMetrics(config Config) map[string]bool {
	result := make(map[string]bool)
	resp, err := http.Get(config.Url)
	if err != nil {
		log.Fatalf("Unable to GET public metrics from %s", config.Url)
	}
	lines, err := responseToLines(resp)
	if err != nil {
		log.Fatal("Unable to parse public metrics response into lines", err)
	}
	metricRegex := regexp.MustCompile(config.Regex)

	for _, line := range lines {
		metrics := metricRegex.FindAllStringSubmatch(line, -1)
		for _, metric := range metrics {
			_, excluded := config.excludesMap[metric[1]]
			if !excluded {
				_, replaced := config.Replacements[metric[1]]
				if replaced {
					result[config.Replacements[metric[1]]] = false
				} else {
					result[metric[1]] = false
				}
			}
		}
	}
	return result
}

func main() {

	// Handle command line arguments
	masterFile := flag.String("master", "../../data/master.csv", "path to master.csv master")
	configFile := flag.String("config", "config.yaml", "Path to config file")
	flag.Parse()

	config := Config{}
	configFileContents, err := os.ReadFile(*configFile)
	if err != nil {
		log.Fatalf("Unable to read config file: %v", err)
	}
	err = yaml.Unmarshal(configFileContents, &config)
	if err != nil {
		log.Fatalf("Unable to parse config file: %v", err)
	}
	config.excludesMap = make(map[string]bool)
	for _, metric := range config.Excludes {
		config.excludesMap[metric] = true
	}

	publicMetrics := getPublicMetrics(config)

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
		_, ok := publicMetrics[metric]
		if !ok {
			log.Fatalf("Could not find master rp metric %s in docs", metric)
		}
		publicMetrics[metric] = true
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

		validate(masterRecord[0])
	}

	for metric, asserted := range publicMetrics {
		if !asserted {
			log.Printf("master.csv doesn't contain a rp metric %s that is referenced in docs", metric)
		}
	}
}
