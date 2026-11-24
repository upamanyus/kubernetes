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

package union

// EvaluateConditions (union.go) spells out what fmt's %v prints of a []Decision instead of
// formatting the slice with %v. This test pins its error to the original fmt.Errorf's, byte for
// byte.

import (
	"context"
	"fmt"
	"testing"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apiserver/pkg/authorization/authorizer"
)

// TestEvaluateConditionsImpossibleDecisionFmt: a leaf that evaluates its conditions map to a
// decision the map excluded (or to an unknown value) gets the original error.
func TestEvaluateConditionsImpossibleDecisionFmt(t *testing.T) {
	effects := []evalTestEffect{effectAllow, effectDeny, effectNoOpinion}
	evals := []authorizer.Decision{authorizer.DecisionAllow, authorizer.DecisionDeny,
		authorizer.DecisionNoOpinion, authorizer.Decision(7), authorizer.Decision(-3)}
	checked := 0
	for _, effect := range effects {
		for _, eval := range evals {
			leaf := &evalTestAuthz{conditionEffect: effect, evalDecision: eval}
			authz, err := New(NamedAuthorizer{AuthorizerName: "leaf", Authorizer: leaf})
			if err != nil {
				t.Fatal(err)
			}
			d := authz.ConditionsAwareAuthorize(context.Background(), nil)
			var sub authorizer.ConditionsAwareDecision
			for _, s := range d.UnionedDecisions() {
				sub = s
			}
			if !sub.IsConditionsMap() || sub.PossibleDecisions().Has(eval) {
				continue
			}
			checked++
			gotDecision, gotReason, gotErr := authz.EvaluateConditions(context.Background(), d, nil)
			want := fmt.Errorf("evaluated to decision %s, but only %v were possible", eval,
				sets.List(sub.PossibleDecisions())).Error()
			if gotErr == nil || gotErr.Error() != want || gotDecision != sub.FailureDecision() {
				t.Errorf("%s/%v: got (%v, %q, %v), want (%v, error %q)", effect, eval, gotDecision, gotReason,
					gotErr, sub.FailureDecision(), want)
			}
		}
	}
	if checked < 6 {
		t.Fatalf("only %d impossible evaluations", checked)
	}
}
