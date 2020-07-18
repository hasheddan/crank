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
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestSort(t *testing.T) {
	dag := New()
	if err := dag.AddNodes("A", "B", "C"); err != nil {
		t.Fatalf("cannot add node: %s", err)
	}
	fmt.Println(dag.nodes)
	if err := dag.AddEdges(map[string][]string{"A": {"B", "C"}, "B": {"C"}}); err != nil {
		t.Fatalf("cannot add edges: %s", err)
	}
	res, err := dag.Sort()
	if err != nil {
		t.Fatalf("got error %s", err)
	}
	fmt.Println(res)
	if !cmp.Equal(res, []string{"C", "B", "A"}) {
		t.Fatalf("wrong order: %v", res)
	}
}
