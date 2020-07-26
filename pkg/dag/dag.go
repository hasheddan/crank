/*
Copyright 2020 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package dag

import (
	"fmt"
	"strings"

	"github.com/hasheddan/crank/apis/v1alpha1"
	"github.com/pkg/errors"
)

// Dag is a directed acyclic graph.
type Dag struct {
	nodes map[string][]string
}

// New returns a new Dag.
func New(deps map[string]v1alpha1.PackageDependencies) (*Dag, error) {
	d := &Dag{map[string][]string{}}
	for _, pkg := range deps {
		if err := d.AddNode(strings.Split(pkg.Image, ":")[0]); err != nil {
			return nil, err
		}
	}
	for _, pkg := range deps {
		if err := d.AddEdges(map[string][]string{strings.Split(pkg.Image, ":")[0]: pkg.Dependencies}); err != nil {
			return nil, err
		}
	}
	return d, nil
}

// AddNodes adds nodes to the graph.
func (d *Dag) AddNodes(names ...string) error {
	for _, n := range names {
		if err := d.AddNode(n); err != nil {
			return err
		}
	}
	return nil
}

// AddNode adds a node to the graph.
func (d *Dag) AddNode(name string) error {
	if _, ok := d.nodes[name]; ok {
		return errors.New("node already exists")
	}
	d.nodes[name] = []string{}
	return nil
}

// NodeExists checks whether a node exists.
func (d *Dag) NodeExists(name string) bool {
	if _, ok := d.nodes[name]; !ok {
		return false
	}
	return true
}

// AddEdges adds edges to the graph.
func (d *Dag) AddEdges(edges map[string][]string) error {
	for f, ne := range edges {
		for _, e := range ne {
			if err := d.AddEdge(f, e); err != nil {
				return err
			}
		}
	}
	return nil
}

// AddEdge adds an edge to the graph.
func (d *Dag) AddEdge(from, to string) error {
	if _, ok := d.nodes[from]; !ok {
		return errors.New(fmt.Sprintf("from node %s does not exist", from))
	}
	if _, ok := d.nodes[to]; !ok {
		return errors.New(fmt.Sprintf("to node %s does not exist", to))
	}
	d.nodes[from] = append(d.nodes[from], to)
	return nil
}

// Sort performs topological sort on the graph.
func (d *Dag) Sort() ([]string, error) {
	visited := map[string]bool{}
	results := make([]string, len(d.nodes))
	for n, deps := range d.nodes {
		if visited[n] == false {
			stack := map[string]bool{}
			if err := d.visit(n, deps, stack, visited, results); err != nil {
				return nil, err
			}
		}
	}
	return results, nil
}

func (d *Dag) visit(name string, deps []string, stack map[string]bool, visited map[string]bool, results []string) error {
	visited[name] = true
	stack[name] = true
	for _, dep := range deps {
		if visited[dep] == false {
			if err := d.visit(dep, d.nodes[dep], stack, visited, results); err != nil {
				return err
			}
		} else if stack[dep] == true {
			return errors.New(fmt.Sprintf("detected cycle: %s", dep))
		}
	}
	for i, r := range results {
		if r == "" {
			results[i] = name
			break
		}
	}
	stack[name] = false
	return nil
}
