package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"steamprofilewatcher/steam"
	"time"
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
	fp, err := os.Open(statisticFilename())
	if err != nil {
		return fmt.Errorf("open stat file: %w", err)
	}
	scanner := bufio.NewScanner(fp)
	idToName := make(map[string]string)
	var doneHeader bool
	for scanner.Scan() {
		if !doneHeader {
			doneHeader = true
			continue
		}
		line := scanner.Text()
		timestamp, stat, err := steam.ParseCSVLine(line)
		if err != nil {
			return fmt.Errorf("parse stat line: %w", err)
		}

		if name, ok := idToName[stat.ID]; ok {
			if name != stat.Name {
				log.Printf("shifted name on %v from %v to %d", stat.ID, name, stat.Name)
			}
		}
		fmt.Println(timestamp)
		fmt.Println(stat)
	}
	if scanner.Err() != nil {
		return fmt.Errorf("read stat file: %w", scanner.Err())
	}
	return nil
}

func generate() error {
	c, err := steam.NewClient(*steamKey, *steamID)
	if err != nil {
		return err
	}
	ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()
	stats, err := c.GetRecentlyPlayedGameStats(ctx)
	if err != nil {
		return err
	}
	return output(stats)
}

func exeDirOrEmpty() string {
	executable, err := os.Executable()
	fmt.Println(executable)
	if err != nil {
		return ""
	}
	absolute, err := filepath.EvalSymlinks(executable)
	fmt.Println(absolute)
	if err != nil {
		return ""
	}
	return filepath.Dir(absolute)
}

func statisticFilename() string {
	return filepath.Join(exeDirOrEmpty(), "steam_profile_watch.csv")
}

func output(stats []steam.GameStat) (e error) {
	filename := statisticFilename()
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND, 0)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		file, err = os.Create(filename)
		if err != nil {
			return err
		}
		if err := prepareCSVHeader(file); err != nil {
			return err
		}
	}
	defer func(file *os.File) {
		e = errors.Join(e, file.Close())
	}(file)

	now := time.Now()
	for _, stat := range stats {
		if _, err := file.WriteString(stat.CSVLine(now) + "\n"); err != nil {
			return err
		}
	}
	return nil
}

func prepareCSVHeader(w io.Writer) error {
	// Add BOM for Excel kanji print.
	if _, err := w.Write([]byte("\uFEFF")); err != nil {
		return err
	}
	if _, err := w.Write([]byte(steam.GameStat{}.CSVHeader() + "\n")); err != nil {
		return err
	}
	return nil
}
