package main

import (
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

func main() {
	flag.Parse()
	if err := handle(); err != nil {
		log.Fatal(err)
	}
}

func handle() error {
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

func output(stats []steam.GameStat) (e error) {
	filename := filepath.Join(exeDirOrEmpty(), "steam_profile_watch.csv")
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
