package corestate

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/diamnet/go/protocols/diamnetcore"
)

type State struct {
	Synced                       bool
	CurrentProtocolVersion       int32
	CoreSupportedProtocolVersion int32
	CoreVersion                  string
}

type Store struct {
	sync.RWMutex
	state State

	// metrics
	Metrics struct {
		CoreSynced                   prometheus.GaugeFunc
		CoreSupportedProtocolVersion prometheus.GaugeFunc
	}
}

func (c *Store) Set(resp *diamnetcore.InfoResponse) {
	c.Lock()
	defer c.Unlock()
	c.state.Synced = resp.IsSynced()
	c.state.CoreVersion = resp.Info.Build
	c.state.CurrentProtocolVersion = int32(resp.Info.Ledger.Version)
	c.state.CoreSupportedProtocolVersion = int32(resp.Info.ProtocolVersion)
}

func (c *Store) SetState(state State) {
	c.Lock()
	defer c.Unlock()
	c.state = state
}

func (c *Store) Get() State {
	c.RLock()
	defer c.RUnlock()
	return c.state
}

func (c *Store) RegisterMetrics(registry *prometheus.Registry) {
	c.Metrics.CoreSynced = prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: "aurora", Subsystem: "diamnet_core", Name: "synced",
			Help: "determines if Diamnet-Core defined by --diamnet-core-url is synced with the network",
		},
		func() float64 {
			if c.Get().Synced {
				return 1
			} else {
				return 0
			}
		},
	)
	registry.MustRegister(c.Metrics.CoreSynced)

	c.Metrics.CoreSupportedProtocolVersion = prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: "aurora", Subsystem: "diamnet_core", Name: "supported_protocol_version",
			Help: "determines the supported version of the protocol by Diamnet-Core defined by --diamnet-core-url",
		},
		func() float64 {
			return float64(c.Get().CoreSupportedProtocolVersion)
		},
	)
	registry.MustRegister(c.Metrics.CoreSupportedProtocolVersion)
}
