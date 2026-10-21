// SPDX-License-Identifier: GPL-2.0-or-later

package validations

import (
	"strings"

	"github.com/openshift-kni/vse-sync-tests/collection_tools/pkg/collectors/devices"
)

const (
	deviceFirmwareID          = TGMEnvVerPath + "/nic-firmware/"
	deviceFirmwareDescription = "Verify NIC firmware version"
)

var (
	MinFirmwareVersion = "4.20"

	// E825/E830 device IDs that should skip firmware version check
	// These cards may have older firmware versions that still work correctly
	e825e830DeviceIDs = map[string]bool{
		"0x12d3": true, // E825-C
		"0x12d8": true, // E825
		"0x12dc": true, // E830 (Timed I/O)
		"0x12dd": true, // E830
	}

	// TEMPORARY local bypass (do not upstream): gnrd-sno E810 runs NVM 4.12.
	// Remove this once the NIC firmware is updated to >= 4.20.
	e810BypassDeviceIDs = map[string]bool{
		"0x579e": true, // E810-XXV
	}
)

func NewDeviceFirmware(ptpDevInfo *devices.PTPDeviceInfo) *VersionCheck {
	parts := strings.Split(ptpDevInfo.FirmwareVersion, " ")

	// Skip version check for E825/E830 cards
	minVersion := MinFirmwareVersion
	if e825e830DeviceIDs[ptpDevInfo.DeviceID] || e810BypassDeviceIDs[ptpDevInfo.DeviceID] {
		minVersion = "0.0" // Allow any version for exempted cards
	}

	return &VersionCheck{
		id:           deviceFirmwareID,
		Version:      ptpDevInfo.FirmwareVersion,
		checkVersion: parts[0],
		MinVersion:   minVersion,
		description:  deviceFirmwareDescription,
		order:        deviceFirmwareOrdering,
	}
}
