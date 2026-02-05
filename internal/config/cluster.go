package config

import "time"

type ClusterConfig struct {
	Algorithm      string
	K              int
	BatchSize      int
	Interval       time.Duration
	MaxIterations  int
	MiniBatchSize  int
	MiniBatchIters int
}

func DefaultClusterConfig() ClusterConfig {
	return ClusterConfig{
		Algorithm:      "kmeans",
		K:              10,
		BatchSize:      1000,
		Interval:       5 * time.Second,
		MaxIterations:  20,
		MiniBatchSize:  256,
		MiniBatchIters: 50,
	}
}
