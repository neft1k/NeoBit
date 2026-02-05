package cluster

import (
	"math"
	"math/rand"
	"time"
)

func assignOnePass(points [][]float32, k int) ([]int, [][]float32) {
	if k <= 0 {
		k = 1
	}
	if k > len(points) {
		k = len(points)
	}

	centroids := make([][]float32, 0, k)
	chosen := make(map[int]struct{}, k)
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	for len(centroids) < k {
		idx := rnd.Intn(len(points))
		if _, ok := chosen[idx]; ok {
			continue
		}
		chosen[idx] = struct{}{}
		centroids = append(centroids, clone(points[idx]))
	}

	assignments := make([]int, len(points))
	for i, p := range points {
		best := 0
		bestDist := distance(p, centroids[0])
		for c := 1; c < k; c++ {
			d := distance(p, centroids[c])
			if d < bestDist {
				bestDist = d
				best = c
			}
		}
		assignments[i] = best
	}

	return assignments, centroids
}

func distance(a, b []float32) float32 {
	var sum float32
	for i := range a {
		d := a[i] - b[i]
		sum += d * d
	}
	return float32(math.Sqrt(float64(sum)))
}

func clone(v []float32) []float32 {
	out := make([]float32, len(v))
	copy(out, v)
	return out
}
