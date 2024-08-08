package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"steamprofilewatcher/helpers"
	"steamprofilewatcher/mode/report"
	"steamprofilewatcher/mode/watch"
	"steamprofilewatcher/steam"
)

var (
	steamKey = flag.String("steamKey", "TOP_SECRET", "the Steam API key from https://steamcommunity.com/dev/apikey")
	steamID  = flag.Int64("steamID", 11223344556677880, "the Steam Account ID from your Steam profile page")
)

var parseMode = flag.Bool("parseMode", true, "whether to parse rather than generate log csv")

func main() {
	flag.Parse()
	if *parseMode {
		if err := parse(); err != nil {
			log.Fatal(err)
		}
	} else {
		if err := generate(); err != nil {
			log.Fatal(err)
		}
	}
}

func parse() error {
	fp, err := os.Open(helpers.StatisticFilename())
	if err != nil {
		return fmt.Errorf("open stat file: %w", err)
	}
	scanner := bufio.NewScanner(fp)
	data, err := report.CSVScan(scanner)
	path := filepath.Join(helpers.ExeDirOrEmpty(), "report.csv")
	if err := report.Output(path, report.Build(data)); err != nil {
		return err
	}
	return nil
}

func generate() error {
	c, err := steam.NewClient(*steamKey, *steamID)
	if err != nil {
		return err
	}
	data, err := watch.Generate(c)
	if err != nil {
		return err
	}
	return watch.Output(data)
}
