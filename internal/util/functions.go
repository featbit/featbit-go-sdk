package util

import (
	"fmt"
	"github.com/featbit/featbit-go-sdk/internal/util/log"
	"io/ioutil"
	"math"
	"math/rand"
	. "net/url"
	"os"
	"strconv"
	"strings"
	"time"
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

var alphabetsMap map[string]string = map[string]string{
	"0": "Q",
	"1": "B",
	"2": "W",
	"3": "S",
	"4": "P",
	"5": "H",
	"6": "D",
	"7": "X",
	"8": "Z",
	"9": "U",
}

func BuildToken(envSecret string) string {

	substring := func(source string, start int, end int) string {
		var r = []rune(source)
		length := len(r)

		if start < 0 || end > length || start > end {
			return ""
		}

		if start == 0 && end == length {
			return source
		}

		var substring = ""
		for i := start; i < end; i++ {
			substring += string(r[i])
		}

		return substring
	}

	encodeNumber := func(number int64, length int) string {
		str := "000000000000" + strconv.FormatInt(number, 10)
		numberWithLeadingZeros := substring(str, len(str)-length, len(str))
		strList := strings.Split(numberWithLeadingZeros, "")
		var encodeStr string
		for _, v := range strList {
			encodeStr = encodeStr + alphabetsMap[v]
		}
		return encodeStr

	}

	text := strings.TrimRight(envSecret, "=")
	now := time.Now().UnixNano() / 1e6

	timestampCode := encodeNumber(now, len(strconv.FormatInt(int64(now), 10)))
	start := math.Max(math.Floor(rand.Float64()*float64(len(text))), 2)

	part1 := encodeNumber(int64(start), 3)
	part2 := encodeNumber(int64(len(timestampCode)), 2)
	part3 := substring(text, 0, int(start))
	part4 := timestampCode
	part5 := substring(text, int(start), len(text))
	result := fmt.Sprintf("%s%s%s%s%s", part1, part2, part3, part4, part5)
	return result
}
