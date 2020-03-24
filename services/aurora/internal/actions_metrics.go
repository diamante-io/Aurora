package aurora

import (
	metrics "github.com/rcrowley/go-metrics"
	"github.com/hcnet/go/services/aurora/internal/actions"
	"github.com/hcnet/go/support/render/hal"
)

// Interface verification
var _ actions.JSONer = (*MetricsAction)(nil)

// MetricsAction collects and renders a snapshot from the metrics system that
// will inlude any previously registered metrics.
type MetricsAction struct {
	Action
	Snapshot map[string]interface{}
}

// JSON is a method for actions.JSON
func (action *MetricsAction) JSON() error {
	action.LoadSnapshot()
	action.Snapshot["_links"] = map[string]interface{}{
		"self": hal.NewLink("/metrics"),
	}

	hal.Render(action.W, action.Snapshot)
	return action.Err
}

// LoadSnapshot populates action.Snapshot
//
// Original code copied from github.com/rcrowley/go-metrics MarshalJSON
func (action *MetricsAction) LoadSnapshot() {
	action.Snapshot = make(map[string]interface{})

	action.App.metrics.Each(func(name string, i interface{}) {
		values := make(map[string]interface{})
		switch metric := i.(type) {
		case metrics.Counter:
			values["count"] = metric.Count()
		case metrics.Gauge:
			values["value"] = metric.Value()
		case metrics.GaugeFloat64:
			values["value"] = metric.Value()
		case metrics.Healthcheck:
			values["error"] = nil
			metric.Check()
			if err := metric.Error(); nil != err {
				values["error"] = metric.Error().Error()
			}
		case metrics.Histogram:
			h := metric.Snapshot()
			ps := h.Percentiles([]float64{0.5, 0.75, 0.95, 0.99, 0.999})
			values["count"] = h.Count()
			values["min"] = h.Min()
			values["max"] = h.Max()
			values["mean"] = h.Mean()
			values["stddev"] = h.StdDev()
			values["median"] = ps[0]
			values["75%"] = ps[1]
			values["95%"] = ps[2]
			values["99%"] = ps[3]
			values["99.9%"] = ps[4]
		case metrics.Meter:
			m := metric.Snapshot()
			values["count"] = m.Count()
			values["1m.rate"] = m.Rate1()
			values["5m.rate"] = m.Rate5()
			values["15m.rate"] = m.Rate15()
			values["mean.rate"] = m.RateMean()
		case metrics.Timer:
			t := metric.Snapshot()
			ps := t.Percentiles([]float64{0.5, 0.75, 0.95, 0.99, 0.999})
			values["count"] = t.Count()
			values["min"] = t.Min()
			values["max"] = t.Max()
			values["mean"] = t.Mean()
			values["stddev"] = t.StdDev()
			values["median"] = ps[0]
			values["75%"] = ps[1]
			values["95%"] = ps[2]
			values["99%"] = ps[3]
			values["99.9%"] = ps[4]
			values["1m.rate"] = t.Rate1()
			values["5m.rate"] = t.Rate5()
			values["15m.rate"] = t.Rate15()
			values["mean.rate"] = t.RateMean()
		}
		action.Snapshot[name] = values
	})
}
