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

package unpack

import (
	"testing"
)

func TestSort(t *testing.T) {
	crds, _, _, err := Resources("crossplane/provider-gcp:v0.11.0")
	if err != nil {
		t.Fatal(err)
	}

	if len(crds) != 19 {
		t.Fatalf("Found %d CRDs but expected %d", len(crds), 19)
	}
}
