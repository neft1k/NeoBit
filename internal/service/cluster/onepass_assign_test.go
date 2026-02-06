package cluster

import "testing"

func TestAssignOnePass(t *testing.T) {
	points := [][]float32{
		{0, 0},
		{1, 1},
		{10, 10},
		{11, 11},
	}
	assignments, centroids := assignOnePass(points, 2)
	if len(assignments) != len(points) {
		t.Fatalf("expected %d assignments, got %d", len(points), len(assignments))
	}
	if len(centroids) != 2 {
		t.Fatalf("expected 2 centroids, got %d", len(centroids))
	}
	if len(centroids[0]) != len(points[0]) {
		t.Fatalf("centroid dimension mismatch")
	}
}
