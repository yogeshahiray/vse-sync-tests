// SPDX-License-Identifier: GPL-2.0-or-later

package collectors

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/openshift-kni/vse-sync-tests/collection_tools/pkg/collectors/contexts"
	"github.com/openshift-kni/vse-sync-tests/collection_tools/pkg/collectors/devices"
)

const (
	DPLLCollectorName = "DPLL"
)

// Returns a new DPLLCollector from the CollectionConstuctor Factory
func NewDPLLCollector(constructor *CollectionConstructor) (Collector, error) {
	ctx, err := contexts.GetPTPDaemonContext(constructor.Clientset, constructor.PTPNodeName)
	if err != nil {
		return &DPLLNetlinkCollector{}, fmt.Errorf("failed to create DPLLCollector: %w", err)
	}

	dpllFSExists, err := devices.IsDPLLFileSystemPresent(ctx, constructor.PTPInterface)
	log.Debug("DPLL FS exists: ", dpllFSExists)

	if dpllFSExists && err == nil {
		return NewDPLLFilesystemCollector(constructor)
	} else {
		return NewDPLLNetlinkCollector(constructor)
	}
}

func init() {
	// The DPLL collector is opt-in: the linuxptp-daemon already logs DPLL
	// monitoring data which the test suite analyses, so it is not run by
	// default. Select it explicitly with --collector DPLL when pin level
	// data (e.g. SMA1 measurements) is required.
	RegisterCollector(DPLLCollectorName, NewDPLLCollector, optin)
}
