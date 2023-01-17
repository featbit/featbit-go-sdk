package util

import (
	"crypto/md5"
	"encoding/binary"
	"math"
	"strconv"
	"strings"
)

func PercentageOfKey(key string) float64 {
	ascii := strings.ReplaceAll(strconv.QuoteToASCII(key), "\"", "")
	bytes := md5.Sum([]byte(ascii))
	num := float64(int32(binary.LittleEndian.Uint32(bytes[:])))
	return math.Abs(num / math.MinInt32)
}

func IfKeyBelongsPercentage(key string, percentageRange []float64) bool {
	if percentageRange[0] == 0 && percentageRange[1] == 1 {
		return true
	}
	percentage := PercentageOfKey(key)
	if percentage >= percentageRange[0] && percentage <= percentageRange[1] {
		return true
	}
	return false
}
