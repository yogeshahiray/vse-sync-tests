// SPDX-License-Identifier: GPL-2.0-or-later

package devices

import (
	"errors"
	"testing"

	"github.com/openshift-kni/vse-sync-tests/collection_tools/pkg/utils"
)

// zl3073xPinsJSON is a pin dump captured from a system where the DPLL of
// the primary interface is a zl3073x device: the GNSS 1PPS reference input
// reports the E825 style board label "GNSS_1PPS_IN" with package label
// "REF4P" and the zl3073x driver reports its state as "selectable".
const zl3073xPinsJSON = `{"pin":[
 {"id":20,"board-label":"GNSS_10M_IN","package-label":"REF2N","type":"gnss","clock-id":4394414709669525026,
  "parent-device":[{"parent-id":9,"direction":"input","prio":14,"state":"selectable","phase-offset":0}]},
 {"id":22,"board-label":"1EPPS_IN","package-label":"REF3N","type":"ext","clock-id":4394414709669525026,
  "parent-device":[{"parent-id":9,"direction":"input","prio":2,"state":"selectable","phase-offset":0}]},
 {"id":23,"board-label":"GNSS_1PPS_IN","package-label":"REF4P","type":"gnss","clock-id":4394414709669525026,
  "parent-device":[{"parent-id":9,"direction":"input","prio":0,"state":"selectable","phase-offset":-5607000},
                   {"parent-id":10,"direction":"input","prio":0,"state":"selectable","phase-offset":-5607000}]},
 {"id":27,"package-label":"OUT5","type":"gnss","clock-id":4394414709669525026,
  "parent-device":[{"parent-id":9,"direction":"output","state":"connected"}]}
]}`

// icePinsJSON is a pin dump in the E810 style: the GNSS 1PPS reference
// input reports the board label "GNSS-1PPS" with state "connected".
const icePinsJSON = `{"pin":[
 {"id":6,"board-label":"GNSS-1PPS","type":"gnss","clock-id":5799633565433967608,
  "parent-device":[{"parent-id":0,"direction":"input","prio":0,"state":"connected","phase-offset":406616064733390},
                   {"parent-id":1,"direction":"input","prio":0,"state":"connected","phase-offset":-1870360}]},
 {"id":7,"board-label":"SMA1","type":"ext","clock-id":5799633565433967608,
  "parent-device":[{"parent-id":0,"direction":"input","prio":1,"state":"connected","phase-offset":0}]}
]}`

func TestSelectPinZl3073xGNSSReference(t *testing.T) {
	// The sysfs pin directory of this interface has no SMA1 pin
	id, pinType, err := selectPin([]byte(zl3073xPinsJSON), 4394414709669525026, "1PPS SDP0 SDP1 SDP2 SDP3")
	if err != nil {
		t.Fatalf("expected the GNSS_1PPS_IN reference to be selected, got %v", err)
	}
	if id != 23 || pinType != OnePPSLabel {
		t.Fatalf("expected pin 23 of type %q, got %d of type %q", OnePPSLabel, id, pinType)
	}
}

func TestSelectPinIceGNSSReference(t *testing.T) {
	// SMA1 exists in sysfs but the 1PPS reference pin is preferred
	id, pinType, err := selectPin([]byte(icePinsJSON), 5799633565433967608, "SMA1 SMA2")
	if err != nil {
		t.Fatalf("expected the GNSS-1PPS reference to be selected, got %v", err)
	}
	if id != 6 || pinType != OnePPSLabel {
		t.Fatalf("expected pin 6 of type %q, got %d of type %q", OnePPSLabel, id, pinType)
	}
}

func TestSelectPinPackageLabelFallback(t *testing.T) {
	// A pin with an unknown board label but the GNSS package label prefix
	pins := `{"pin":[
	 {"id":9,"board-label":"PLATFORM_SPECIFIC_NAME","package-label":"REF4N","type":"gnss","clock-id":42,
	  "parent-device":[{"parent-id":0,"direction":"input","prio":0,"state":"connected","phase-offset":0}]}
	]}`

	id, pinType, err := selectPin([]byte(pins), 42, "")
	if err != nil {
		t.Fatalf("expected the REF4 package pin to be selected, got %v", err)
	}
	if id != 9 || pinType != OnePPSLabel {
		t.Fatalf("expected pin 9 of type %q, got %d of type %q", OnePPSLabel, id, pinType)
	}
}

func TestSelectPinSMAGatedBySysfs(t *testing.T) {
	// The netlink pin table contains an SMA1 pin...
	pins := `{"pin":[
	 {"id":7,"board-label":"SMA1","type":"ext","clock-id":42,
	  "parent-device":[{"parent-id":0,"direction":"input","prio":1,"state":"connected","phase-offset":0}]}
	]}`

	// ...but without SMA1 in the sysfs pin directory it must not be used
	_, _, err := selectPin([]byte(pins), 42, "1PPS SDP0")
	if !isRequirementsNotMet(err) {
		t.Fatalf("expected no suitable pin without sysfs SMA1, got %v", err)
	}

	// ...with SMA1 in sysfs it is selected
	id, pinType, err := selectPin([]byte(pins), 42, "SMA1 SMA2")
	if err != nil {
		t.Fatalf("expected the SMA1 pin to be selected, got %v", err)
	}
	if id != 7 || pinType != SMA1Label {
		t.Fatalf("expected pin 7 of type %q, got %d of type %q", SMA1Label, id, pinType)
	}
}

func TestSelectPinWrongClockID(t *testing.T) {
	_, _, err := selectPin([]byte(zl3073xPinsJSON), 12345, "1PPS")
	if !isRequirementsNotMet(err) {
		t.Fatalf("expected no suitable pin for unknown clock ID, got %v", err)
	}
}

func TestHasSysfsSMAPinName(t *testing.T) {
	for _, listing := range []string{"", "1PPS SDP0 SDP1 SDP2 SDP3 TIME_SYNC", "1PPS", "SDP0 SMA2"} {
		if hasSysfsSMAPinName(listing) {
			t.Fatalf("expected no SMA1 in %q", listing)
		}
	}

	for _, listing := range []string{"SMA1", "SMA1 SMA2", "INT_PPS SMA1"} {
		if !hasSysfsSMAPinName(listing) {
			t.Fatalf("expected SMA1 in %q", listing)
		}
	}
}

func isRequirementsNotMet(err error) bool {
	var reqNotMet *utils.RequirementsNotMetError
	return errors.As(err, &reqNotMet)
}
