package watch

import (
	"context"
	"errors"
	"io"
	"os"
	"steamprofilewatcher/helpers"
	"steamprofilewatcher/steam"
	"time"
)

func Generate(c *steam.Client) ([]steam.GameStat, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()
	stats, err := c.GetRecentlyPlayedGameStats(ctx)
	if err != nil {
		return nil, err
	}
	return stats, nil
}

func Output(stats []steam.GameStat) (e error) {
	filename := helpers.StatisticFilename()
	fp, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND, 0)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		fp, err = os.Create(filename)
		if err != nil {
			return err
		}
		if err := prepareCSVHeader(fp); err != nil {
			return err
		}
	}
	defer func(file *os.File) {
		e = errors.Join(e, file.Close())
	}(fp)

	now := time.Now()
	for _, stat := range stats {
		if _, err := fp.WriteString(stat.CSVLine(now) + "\n"); err != nil {
			return err
		}
	}
	return nil
}

func prepareCSVHeader(w io.Writer) error {
	if err := helpers.WriteBOMPrefix(w); err != nil {
		return err
	}
	if _, err := w.Write([]byte(steam.GameStat{}.CSVHeader() + "\n")); err != nil {
		return err
	}
	return nil
}
