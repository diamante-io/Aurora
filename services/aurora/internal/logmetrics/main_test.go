package logmetrics

import (
	"bytes"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestLogPackageMetrics(t *testing.T) {
	output := new(bytes.Buffer)
	l, m := New()
	l.Logger.Formatter.(*logrus.TextFormatter).DisableColors = true
	l.Logger.Level = logrus.DebugLevel
	l.Logger.Out = output

	for _, meter := range *m {
		assert.Equal(t, int64(0), meter.Count())
	}

	l.Debug("foo")
	l.Info("foo")
	l.Warn("foo")
	l.Error("foo")
	assert.Panics(t, func() {
		l.Panic("foo")
	})

	for _, meter := range *m {
		assert.Equal(t, int64(1), meter.Count())
	}
}
