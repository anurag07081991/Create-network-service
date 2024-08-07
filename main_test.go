package main

import (
	"testing"
)

func TestShortestPath(t *testing.T) {
	g := NewGraph()
	g.AddEdge(1, 2)
	g.AddEdge(2, 3)
	g.AddEdge(3, 4)

	path, found := g.ShortestPath(1, 4)
	if !found || len(path) != 4 || path[0] != 1 || path[1] != 2 || path[2] != 3 || path[3] != 4 {
		t.Fatalf("Expected path [1, 2, 3, 4], got %v", path)
	}
}
