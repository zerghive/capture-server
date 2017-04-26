package util

import (
	"github.com/rcrowley/go-librato"

	"os"
	"sync"
	"time"
)

const (
	cLibratoUser  = "denis.s.mishin@gmail.com"
	cLibratoToken = "a4b1b242c0337193ac325eab53b958521ff55f942ddad4de0f17de9e9e5325b9"

	cEnableMetrics = false
)

type ISampler interface {
	Name() string
	Sample() float64 // called when we need to take measures
}

var libratoClient struct {
	lock    sync.Mutex
	metrics librato.Metrics
}

func getMetrics() (librato.Metrics, error) {
	libratoClient.lock.Lock()
	defer libratoClient.lock.Unlock()

	if libratoClient.metrics == nil {
		hostname, err := os.Hostname()
		if err != nil {
			return nil, err
		}
		libratoClient.metrics = librato.NewSimpleMetrics(cLibratoUser, cLibratoToken, hostname)
	}
	return libratoClient.metrics, nil
}

func AddGauge(metric ISampler, interval time.Duration) error {
	if !cEnableMetrics {
		return nil
	}

	m, err := getMetrics()
	if err != nil {
		return err
	}

	gauge := m.GetGauge(metric.Name())
	go SafeRun(func() {
		for {
			time.Sleep(interval)
			gauge <- int64(metric.Sample())
		}
	})

	return nil
}
