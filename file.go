package expire

import (
	"os"
	"path/filepath"
)

func exists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

func findFileUp(fileName string) string {
	filePath := fileName

	for {
		if exists(filePath) {
			return filePath
		}

		abs, err := filepath.Abs(filePath)
		if err != nil {
			return ""
		}

		if abs == "/"+fileName {
			return ""
		}

		filePath = filepath.Join("..", filePath)
	}
	return ""

}

func getExpirationsFilePath(config GlobalConfig) string {
	return findFileUp(config.getFileName())
}
