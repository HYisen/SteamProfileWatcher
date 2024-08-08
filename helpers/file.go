package helpers

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func WriteBOMPrefix(w io.Writer) error {
	// Add BOM for Excel kanji print.
	_, err := w.Write([]byte("\uFEFF"))
	return err
}

func ExeDirOrEmpty() string {
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

func StatisticFilename() string {
	return filepath.Join(ExeDirOrEmpty(), "steam_profile_watch.csv")
}
