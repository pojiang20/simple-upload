package util

import (
	"io"
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
