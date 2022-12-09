package util

import (
	"io"
	"os"
	"path"
	"strings"
)

func GetSize(content io.Seeker) (int64, error) {
	size, err := content.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, err
	}
	_, err = content.Seek(0, io.SeekStart)
	if err != nil {
		return 0, err
	}
	return size, nil
}

func GetUrl(dstPath, root string) string {
	url := strings.TrimPrefix(dstPath, root)
	url = path.Join("/files", url)
	return url
}

func GenDir(absPath string, dirName string) (string, error) {
	dirPath := path.Join(absPath, dirName)
	if _, err := os.Stat(dirPath); err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir(dirPath, 0755)
			if err != nil {
				Zlog.Errorf("mkdir %s error: %v", dirPath, err)
				return "", err
			}
		}
	}
	return dirPath, nil
}
