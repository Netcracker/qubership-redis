package helper

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"go.uber.org/zap"
)

func GetEnvBool(key string, fallback bool) bool {
	if value, ok := os.LookupEnv(key); ok {
		bvalue, err := strconv.ParseBool(value)
		if err != nil {
			log.Error(fmt.Sprintf("Can't parse %s boolean variable", key), zap.Error(err))
			panic(err)
		}
		return bvalue
	}
	return fallback
}

func GenerateDbName() string {
	currentTime := time.Now().UTC()
	timestamp := currentTime.Format("150405.000.020106")
	return strings.ReplaceAll(timestamp, ".", "")
}
