package synccmd

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/indaco/tempo/internal/utils"
)

func getLastRunTimestamp(cacheFile string) int64 {
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		// If file doesn't exist, return 0 (meaning everything should be processed)
		return 0
	}

	ts, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return 0
	}
	return ts
}

func saveLastRunTimestamp(cacheFile string) error {
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	return utils.WriteStringToFile(cacheFile, timestamp)
}

// getFileLastModifiedTime retrieves the last modified timestamp of a given file.
func getFileLastModifiedTime(filePath string) (int64, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}
	return info.ModTime().Unix(), nil
}
