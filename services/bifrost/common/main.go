package common

import (
	"github.com/diamnet/go/support/log"
)

const DiamNetAmountPrecision = 7

func CreateLogger(serviceName string) *log.Entry {
	return log.DefaultLogger.WithField("service", serviceName)
}
