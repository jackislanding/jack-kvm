package kvm

import (
	"math"
	"math/rand"
	"time"
)

type JigglerConfig struct {
	coordX          int
	coordY          int
	intervalSeconds float64
	jitterSeconds   float64
	lastInterval    float64
}

func (j *JigglerConfig) calcNewInterval() {
	jitter := (rand.Float64() * j.jitterSeconds * 2) - j.jitterSeconds
	logger.Infof("jiggler jitter: %v", jitter)
	j.lastInterval = math.Max(0, j.intervalSeconds+jitter)
	logger.Infof("jiggler new interval: %v", j.lastInterval)
}

var lastUserInput = time.Now()

var jigglerEnabled = false

func rpcSetJigglerState(enabled bool) {
	jigglerEnabled = enabled
}
func rpcGetJigglerState() bool {
	return jigglerEnabled
}

func init() {
	ensureConfigLoaded()
	jc := &JigglerConfig{
		coordX:          0,
		coordY:          0,
		intervalSeconds: 20,
		jitterSeconds:   2,
		lastInterval:    0,
	}
	jc.calcNewInterval()
	go runJiggler(jc)
}

func runJiggler(j *JigglerConfig) {
	for {
		if jigglerEnabled {
			if time.Since(lastUserInput) > time.Duration(j.lastInterval)*time.Second {
				//TODO: change to rel mouse
				err := rpcAbsMouseReport(1, 1, 0)
				if err != nil {
					logger.Warnf("Failed to jiggle mouse: %v", err)
				}
				err = rpcAbsMouseReport(0, 0, 0)
				if err != nil {
					logger.Warnf("Failed to reset mouse position: %v", err)
				}
			}
		}
		time.Sleep(time.Duration(j.lastInterval) * time.Second)
		j.calcNewInterval()
	}
}
