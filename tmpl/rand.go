package tmpl

import (
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	maxLengthEmailUserName = 10
	maxLengthEmailDomain   = 5

	maxLengthString = 10

	letterBytes = "abcdefghijklmnopqrstuvwxyz0123456789"
)

func randStringBytes(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func RandomEmail() string {
	return randStringBytes(maxLengthEmailUserName) + "@" + randStringBytes(maxLengthEmailDomain) + ".com"
}

func RandomString() string {
	return randStringBytes(maxLengthString)
}

func RandomTime() string {
	randTime := time.Duration(rand.Int31n(86400)) * time.Second
	return time.Now().UTC().Add(randTime).Format("15:04:05.000000")
}

func RandomDate() string {
	randDate := time.Duration(rand.Int31n(365)) * time.Hour * 24
	return time.Now().UTC().Add(randDate).Format("2006-01-02")
}

func RandomDateTime() string {
	randTime := time.Duration(rand.Int31n(86400)) * time.Second
	randDate := time.Duration(rand.Int31n(365)) * time.Hour * 24
	return time.Now().UTC().Add(randTime).Add(randDate).Format("2006-01-02 15:04:05.000000")
}

func RandomInteger() string {
	return strconv.Itoa(int(rand.Int63()))
}

func RandomFloat() string {
	return strconv.FormatFloat(rand.Float64(), 'f', 10, 64)
}

func RandomRequest() string {
	newUUID, err := uuid.NewRandom()
	if err != nil {
		return ""
	}
	return strings.ReplaceAll(newUUID.String(), "-", "")
}
