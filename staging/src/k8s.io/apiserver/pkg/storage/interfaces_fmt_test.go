/*
Copyright 2026 The Kubernetes Authors.

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

package storage

import (
	"fmt"
	"math/rand"
	"testing"

	"k8s.io/apimachinery/pkg/types"
)

// Preconditions.Check's UID message formats string(uid): byte-identical to formatting the
// types.UID itself (a method-less named string), which fmt prints as its value.
func TestPreconditionUIDMessageSame(t *testing.T) {
	ss := []string{"", "a", "6c0f3b3e-1d2c-4b8e-9a7f-0c1d2e3f4a5b", "\"q\"", "\\", "\x00\x01\x7f", "\xff\xfe", "é", "\U0001F600", "%v"}
	r := rand.New(rand.NewSource(1))
	for i := 0; i < 100000; i++ {
		b := make([]byte, r.Intn(40))
		for j := range b {
			b[j] = byte(r.Intn(256))
		}
		ss = append(ss, string(b))
	}
	const f = "Precondition failed: UID in precondition: %v, UID in object meta: %v"
	for i, s := range ss {
		u, v := types.UID(s), types.UID(ss[(i+1)%len(ss)])
		if x, y := fmt.Sprintf(f, u, v), fmt.Sprintf(f, string(u), string(v)); x != y {
			t.Fatalf("%q: %q vs %q", s, x, y)
		}
	}
}
