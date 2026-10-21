// SPDX-License-Identifier: GPL-2.0-or-later

package collectors

import (
	"fmt"
	"log"
)

type collectonBuilderFunc func(*CollectionConstructor) (Collector, error)
type collectorInclusionType int

const (
	required collectorInclusionType = iota
	optional
	optin
)

type CollectorRegistry struct {
	registry map[string]collectonBuilderFunc
	required []string
	optional []string
	optin    []string
}

var registry *CollectorRegistry

func GetRegistry() *CollectorRegistry {
	return registry
}

func (reg *CollectorRegistry) GetBuilderFunc(collectorName string) (collectonBuilderFunc, error) {
	builderFunc, ok := reg.registry[collectorName]
	if !ok {
		return nil, fmt.Errorf("not index in registry for collector named %s", collectorName)
	}

	return builderFunc, nil
}

func (reg *CollectorRegistry) GetRequiredNames() []string {
	return reg.required
}

func (reg *CollectorRegistry) GetOptionalNames() []string {
	return reg.optional
}

func (reg *CollectorRegistry) GetOptInNames() []string {
	return reg.optin
}

func (reg *CollectorRegistry) register(
	collectorName string,
	builderFunc collectonBuilderFunc,
	inclusionType collectorInclusionType,
) {
	reg.registry[collectorName] = builderFunc

	switch inclusionType {
	case required:
		reg.required = append(reg.required, collectorName)
	case optional:
		reg.optional = append(reg.optional, collectorName)
	case optin:
		// opt-in collectors are resolvable by name but are not part of the
		// default set of collectors that run when "all" is selected
		reg.optin = append(reg.optin, collectorName)
	default:
		log.Panic("Incorrect collector inclusion type")
	}
}

func RegisterCollector(collectorName string, builderFunc collectonBuilderFunc, inclusionType collectorInclusionType) {
	if registry == nil {
		registry = &CollectorRegistry{
			registry: make(map[string]collectonBuilderFunc, 0),
			required: make([]string, 0),
			optional: make([]string, 0),
			optin:    make([]string, 0),
		}
	}

	registry.register(collectorName, builderFunc, inclusionType)
}
