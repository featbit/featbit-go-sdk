package util

import (
	"github.com/featbit/featbit-go-sdk/internal/util/log"
	"io/ioutil"
	. "net/url"
	"os"
	"strings"
	"unicode"
)

func ReadFile(file string) []byte {
	f, err := os.Open(file)
	if err != nil {
		log.LogError("FB GO SDK: error loading file %s - %v", file, err)
		return []byte(nil)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.LogError("FB GO SDK: error closing file %s - %v", f.Name(), err)
		}
	}(f)

	fd, err := ioutil.ReadAll(f)
	if err != nil {
		log.LogError("FB GO SDK: error loading file %s - %v", file, err)
		return []byte(nil)
	}
	return fd
}

func IsEnvSecretValid(envSecret string) bool {
	es := strings.Trim(envSecret, " ")
	if es == "" {
		return false
	}
	for _, r := range es {
		if !unicode.Is(unicode.ASCII_Hex_Digit, r) {
			return false
		}
	}
	return true
}

func IsUrl(url string) bool {
	if _, err := ParseRequestURI(url); err != nil {
		return false
	}
	return true
}
