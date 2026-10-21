// SPDX-License-Identifier: GPL-2.0-or-later

package runner

import (
	"slices"
	"testing"
)

func TestOptInCollectorsAreNotRunByDefault(t *testing.T) {
	if len(OptInCollectorNames) == 0 {
		t.Fatal("expected at least one opt-in collector to be registered")
	}

	for _, name := range OptInCollectorNames {
		if slices.Contains(GetCollectorsToRun([]string{All}), name) {
			t.Errorf("opt-in collector %s should not be included when running all collectors", name)
		}

		if !slices.Contains(GetCollectorsToRun([]string{name}), name) {
			t.Errorf("opt-in collector %s should be included when explicitly selected", name)
		}
	}
}

func TestRequiredCollectorsAreAlwaysRun(t *testing.T) {
	for _, name := range RequiredCollectorNames {
		if !slices.Contains(GetCollectorsToRun([]string{}), name) {
			t.Errorf("required collector %s missing from selection", name)
		}
	}
}
