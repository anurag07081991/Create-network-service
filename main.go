package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
)

// Graph representation
type Graph struct {
	adjacencyList map[int][]int
	mu            sync.RWMutex
}

func NewGraph() *Graph {
	return &Graph{adjacencyList: make(map[int][]int)}
}

func (g *Graph) AddEdge(u, v int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.adjacencyList[u] = append(g.adjacencyList[u], v)
	g.adjacencyList[v] = append(g.adjacencyList[v], u)
}

func (g *Graph) ShortestPath(start, end int) ([]int, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if start == end {
		return []int{start}, true
	}

	visited := make(map[int]bool)
	queue := []int{start}
	predecessors := make(map[int]int)
	visited[start] = true

	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		if node == end {
			path := []int{}
			for at := end; at != start; at = predecessors[at] {
				path = append([]int{at}, path...)
			}
			path = append([]int{start}, path...)
			return path, true
		}
		for _, neighbor := range g.adjacencyList[node] {
			if !visited[neighbor] {
				visited[neighbor] = true
				predecessors[neighbor] = node
				queue = append(queue, neighbor)
			}
		}
	}
	return nil, false
}

var graphs = make(map[string]*Graph)
var mu = sync.Mutex{}
var idCounter = 0

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/graph", createGraphHandler).Methods("POST")
	r.HandleFunc("/graph/{id}/shortest_path", shortestPathHandler).Methods("GET")
	r.HandleFunc("/graph/{id}", deleteGraphHandler).Methods("DELETE")

	log.Fatal(http.ListenAndServe(":8080", r))
}

func createGraphHandler(w http.ResponseWriter, r *http.Request) {
	var edges []struct {
		U int `json:"u"`
		V int `json:"v"`
	}
	if err := json.NewDecoder(r.Body).Decode(&edges); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	graph := NewGraph()
	for _, edge := range edges {
		graph.AddEdge(edge.U, edge.V)
	}

	mu.Lock()
	idCounter++
	id := idCounter
	graphs[fmt.Sprintf("%d", id)] = graph
	mu.Unlock()

	json.NewEncoder(w).Encode(map[string]int{"id": id})
}

func shortestPathHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var params struct {
		Start int `json:"start"`
		End   int `json:"end"`
	}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mu.Lock()
	graph, ok := graphs[id]
	mu.Unlock()
	if !ok {
		http.Error(w, "Graph not found", http.StatusNotFound)
		return
	}

	path, found := graph.ShortestPath(params.Start, params.End)
	if !found {
		http.Error(w, "Path not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"path": path,
	})
}

func deleteGraphHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	mu.Lock()
	_, exists := graphs[id]
	if exists {
		delete(graphs, id)
	}
	mu.Unlock()

	if !exists {
		http.Error(w, "Graph not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
