package util

import (
	"github.com/featbit/featbit-go-sdk/internal/util/log"
	"io/ioutil"
	"os"
)

func ReadFile(file string) []byte {
	f, err := os.Open(file)
	if err != nil {
		log.LogError("FB GO SDK: error loading file %s - %v", file, err)
		return []byte(nil)
	}
	defer f.Close()

	fd, err := ioutil.ReadAll(f)
	if err != nil {
		log.LogError("FB GO SDK: error loading file %s - %v", file, err)
		return []byte(nil)
	}
	return fd
}
