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
	screenWidth     int
	screenHeight    int
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
	j := &JigglerConfig{
		coordX:          0,
		coordY:          0,
		intervalSeconds: 5,
		jitterSeconds:   2,
		lastInterval:    0,
		screenWidth:     1920,
		screenHeight:    1080,
	}
	j.calcNewInterval()
	go j.runJiggler()
}

func (j *JigglerConfig) runJiggler() {
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
		j.move(590, 650, 1.0) // TODO testing
		time.Sleep(time.Duration(j.lastInterval) * time.Second)
		j.calcNewInterval()
	}
}

func (j *JigglerConfig) move(targetX int, targetY int, speedFactor float64) {
	//Navigate cursor to target coordinates with organic movement patterns.
	nodes := calculatePathNodes()
	logger.Infof("[jiggler.go:move] nodes: %v", nodes)
	curvesX, curvesY := j.generatePathCoordinates(nodes, targetX, targetY)
	logger.Infof("[jiggler.go:move] \ncurvesX: %v\ncurvesY: %v", curvesX, curvesY)
	//trajectory := j.computeTrajectory(nodes, curvesX, curvesY)
	//interval := calculateMovementInterval(curvesX, curvesY, speedFactor)
	//performMovement(trajectory, interval)
}

func (j *JigglerConfig) generatePathCoordinates(nodes int, targetX int, targetY int) ([]float64, []float64) {
	coordsX, coordsY := generateBaseCoordinates(nodes, j.coordX, j.coordY, targetX, targetY)
	variance := randRange(7, 12)
	logger.Infof("[jiggler.go:generatePathCoordinates]\ncoordsX: %v\ncoordsY: %v\nvariance: %v", coordsX, coordsY, variance)
	return applyCoordinateVariance(float64(variance), nodes, coordsX, coordsY)
}

func (j *JigglerConfig) largestScreenDimension() int {
	return max(j.screenWidth, j.screenHeight)
}

func generateBaseCoordinates(nodes int, x1 int, y1 int, x2 int, y2 int) ([]float64, []float64) {
	coordsX := linspace(float64(x1), float64(x2), nodes)
	coordsY := linspace(float64(y1), float64(y2), nodes)
	return coordsX, coordsY
}

func calculatePathNodes() int {
	base := randRange(2, 7)
	ceiling := randRange(10, 15)
	return randRange(base, ceiling)
}

func randRange(min, max int) int {
	return rand.Intn(max-min) + min
}

func linspace(start, stop float64, steps int) []float64 {
	result := make([]float64, steps)
	stepSize := (stop - start) / float64(steps-1)

	for i := 0; i < steps; i++ {
		result[i] = start + float64(i)*stepSize
	}

	return result
}

func applyCoordinateVariance(variance float64, numPoints int, pathsX []float64, pathsY []float64) ([]float64, []float64) {
	if numPoints < 2 {
		val1 := []float64{pathsX[0], pathsX[len(pathsX)-1]}
		val2 := []float64{pathsY[0], pathsY[len(pathsY)-1]}
		return val1, val2
	}

	offsetsX := randomNormalSamples(0, variance, numPoints)
	offsetsY := randomNormalSamples(0, variance, numPoints)

	//No variance on start or end points
	offsetsX[0], offsetsY[0], offsetsX[numPoints-1], offsetsY[numPoints-1] = 0, 0, 0, 0

	logger.Infof("[jiggler.go:applyCoordinateVariance]\noffsetsX: %v\noffsetsY: %v", offsetsX, offsetsY)

	return addOffsets(pathsX, offsetsX), addOffsets(pathsY, offsetsY)
}

func addOffsets(coords []float64, offsets []float64) []float64 {
	offsetCoords := make([]float64, len(coords))
	for i := range coords {
		offsetCoords[i] = coords[i] + offsets[i]
	}
	return offsetCoords
}

func randomNormalSamples(mean float64, stdDev float64, nSamples int) []float64 {
	samples := make([]float64, nSamples)
	for x := range samples {
		samples[x] = rand.NormFloat64()*stdDev + mean
	}
	return samples
}

//func (j *JigglerConfig) computeTrajectory(nodes int, x []int, y []int) ([]float64, []float64) {
//	unique_mask := createBooleanArray(len(x), true)
//	for i := range x {
//		if x[i] == x[i-1] && y[i] == y[i-1] {
//			unique_mask[i] = false
//		}
//	}
//	x = x[unique_mask]
//	y = y[unique_mask]
//
//	if len(x) < 4 {
//		t_points := linspace(0, 1, j.largestScreenDimension())
//		x_interp := np.interp(t_points, linspace(0, 1, len(x)), x)
//		y_interp := np.interp(t_points, linspace(0, 1, len(y)), y)
//		return x_interp, y_interp
//	}
//
//	try:
//		curveDegree := min(3, len(x) - 1)
//		spline_params, _ := makeSplprep(x, y, curveDegree)
//		//spline_params, _ = interpolate.splprep([x, y], k=curveDegree, s=0)
//		t_points := linspace(0, 1, j.largestScreenDimension())
//		return interpolate.splev(t_points, spline_params)
//	except ValueError:
//		t_points := linspace(0, 1, j.largestScreenDimension())
//		x_interp := np.interp(t_points, linspace(0, 1, len(x)), x)
//		y_interp := np.interp(t_points, linspace(0, 1, len(y)), y)
//	return x_interp, y_interp
//}

func createBooleanArray(length int, value bool) []bool {
	// Create a slice with the given length and initialize all elements with the provided value
	array := make([]bool, length)
	for i := range array {
		array[i] = value
	}
	return array
}

// makeSplprep creates a B-spline representation for a curve
func makeSplprep(x, y []float64, degree int) (knots []float64, coefficients []float64) {
	n := len(x)
	if n != len(y) {
		panic("x and y arrays must have the same length")
	}

	// Generate uniform knots vector
	knots = make([]float64, n+degree+1)
	for i := 0; i <= n; i++ {
		knots[i] = float64(i)
	}
	for i := n + 1; i < len(knots); i++ {
		knots[i] = float64(n)
	}

	// Set up the matrix system for spline coefficients (simplified approach)
	// Here we directly take the y values as the spline coefficients for simplicity
	// This is a rough approximation and doesn't solve the system for real B-spline coefficients
	coefficients = append([]float64{}, y...)

	return knots, coefficients
}
