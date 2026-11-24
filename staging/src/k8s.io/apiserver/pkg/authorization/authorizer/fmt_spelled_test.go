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

package authorizer

// UnconditionalPartsOrFailClosed (conditions.go) and ConditionsAwareDecisionUnion.Add
// (conditionsunion.go) spell out what fmt's %v prints of a slice ("[" + the elements' %v
// separated by spaces + "]"), instead of formatting the slice with %v. These tests pin their
// output to the original fmt calls', byte for byte.

import (
	"fmt"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/util/sets"
	utilvalidation "k8s.io/apimachinery/pkg/util/validation"
)

// spelled is the fork's loop, over any []Decision: what fmt's %v prints of the slice, without its
// brackets.
func spelled(ds []Decision) string {
	possible := ""
	for i, p := range ds {
		if i > 0 {
			possible += " "
		}
		possible += p.String()
	}
	return possible
}

// TestDecisionSliceSpelled: the loop and fmt's %v agree on every subset of the three decisions and
// two unknown values, in every order of the subset's elements sets.List gives (sorted).
func TestDecisionSliceSpelled(t *testing.T) {
	all := []Decision{DecisionDeny, DecisionAllow, DecisionNoOpinion, Decision(7), Decision(-3)}
	for mask := 0; mask < 1<<len(all); mask++ {
		s := sets.New[Decision]()
		for i, d := range all {
			if mask&(1<<i) != 0 {
				s.Insert(d)
			}
		}
		ds := sets.List(s)
		if got, want := "["+spelled(ds)+"]", fmt.Sprintf("%v", ds); got != want {
			t.Errorf("%d: spelled %q, fmt %q", mask, got, want)
		}
	}
	if got, want := "["+spelled(nil)+"]", fmt.Sprintf("%v", []Decision(nil)); got != want {
		t.Errorf("nil: spelled %q, fmt %q", got, want)
	}
}

func fmtTestConditions(id string) []Condition {
	return []Condition{GenericCondition{ID: "example.com/" + id, Condition: "c"}}
}

// TestUnconditionalPartsOrFailClosedFmt: UnconditionalPartsOrFailClosed's reason is the original
// fmt.Sprintf's, for conditions maps of every shape and unions of them.
func TestUnconditionalPartsOrFailClosedFmt(t *testing.T) {
	var base []ConditionsAwareDecision
	for mask := 0; mask < 8; mask++ {
		var deny, noOpinion, allow []Condition
		if mask&1 != 0 {
			deny = fmtTestConditions("deny")
		}
		if mask&2 != 0 {
			noOpinion = fmtTestConditions("noopinion")
		}
		if mask&4 != 0 {
			allow = fmtTestConditions("allow")
		}
		base = append(base, ConditionsAwareDecisionConditionsMap(deny, noOpinion, allow))
	}
	base = append(base, ConditionsAwareDecisionAllow("", nil), ConditionsAwareDecisionDeny("", nil),
		ConditionsAwareDecisionNoOpinion("", nil))
	ds := append([]ConditionsAwareDecision{}, base...)
	for _, a := range base {
		for _, b := range base {
			var u ConditionsAwareDecisionUnion
			u.Add("a.example.com", a)
			u.Add("b.example.com", b)
			ds = append(ds, u.ToDecision())
		}
	}
	conditional := 0
	for i, d := range ds {
		gotDecision, gotReason, gotErr := d.UnconditionalPartsOrFailClosed()
		if d.IsAllow() || d.IsDeny() || d.IsNoOpinion() {
			continue
		}
		conditional++
		wantReason := fmt.Sprintf(
			"failed closed from conditional decision (with possible outcomes %v) to %s during unconditional authorization",
			sets.List(d.PossibleDecisions()),
			d.FailureDecision().String(),
		)
		if gotDecision != d.FailureDecision() || gotReason != wantReason || gotErr != nil {
			t.Errorf("%d (%s): got (%v, %q, %v), want (%v, %q, nil)", i, d, gotDecision, gotReason, gotErr,
				d.FailureDecision(), wantReason)
		}
	}
	if conditional == 0 {
		t.Fatal("no conditional decision tested")
	}
}

// TestAddInvalidNameFmt: Add's error for an authorizerName that is not a DNS-1123 subdomain is the
// original fmt.Errorf's, for names with one message and with two (too long and not matching), and
// special characters.
func TestAddInvalidNameFmt(t *testing.T) {
	names := []string{"", "UPPER", "a_b", "-a", "a.", "a..b", "with space", "%v [x] %s", "a\nb", "\"q\"",
		"ä.example.com", strings.Repeat("a", 254), strings.Repeat("A", 300), strings.Repeat("a.", 127) + "a"}
	checked := 0
	for _, name := range names {
		errs := utilvalidation.IsDNS1123Subdomain(name)
		if len(errs) == 0 {
			continue
		}
		var u ConditionsAwareDecisionUnion
		u.Add(name, ConditionsAwareDecisionNoOpinion("", nil))
		want := fmt.Errorf("invalid authorizerName: %v", errs).Error()
		found := false
		for _, err := range u.errs {
			if err.Error() == want {
				found = true
			}
		}
		if !found {
			t.Errorf("%q: no error %q among %v", name, want, u.errs)
		}
		checked++
	}
	if checked < 10 {
		t.Fatalf("only %d invalid names", checked)
	}
	for _, errs := range [][]string{nil, {}, {"one"}, {"one", "two"}, {"a b", "[c]", "%v", ""}} {
		msgs := ""
		for i, e := range errs {
			if i > 0 {
				msgs += " "
			}
			msgs += e
		}
		if got, want := fmt.Errorf("invalid authorizerName: [%s]", msgs).Error(),
			fmt.Errorf("invalid authorizerName: %v", errs).Error(); got != want {
			t.Errorf("%q: spelled %q, fmt %q", errs, got, want)
		}
	}
}
