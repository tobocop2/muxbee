package bridges

import (
	_ "embed"
	"fmt"
	"sort"

	"gopkg.in/yaml.v3"
)

//go:embed bridges.yaml
var bridgesYAML []byte

var registry map[string]BridgeInfo

func init() {
	if err := loadRegistry(); err != nil {
		panic(fmt.Sprintf("failed to load bridges registry: %v", err))
	}
}

func loadRegistry() error {
	var raw map[string]BridgeInfo
	if err := yaml.Unmarshal(bridgesYAML, &raw); err != nil {
		return err
	}

	registry = make(map[string]BridgeInfo, len(raw))
	for name, info := range raw {
		info.Name = name
		registry[name] = info
	}

	return nil
}

// Get returns bridge info by name, or nil if not found.
func Get(name string) *BridgeInfo {
	if info, ok := registry[name]; ok {
		return &info
	}
	return nil
}

// List returns all bridges sorted by name.
func List() []BridgeInfo {
	bridges := make([]BridgeInfo, 0, len(registry))
	for _, info := range registry {
		bridges = append(bridges, info)
	}
	sort.Slice(bridges, func(i, j int) bool {
		return bridges[i].Name < bridges[j].Name
	})
	return bridges
}

// Names returns all bridge names sorted alphabetically.
func Names() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Exists checks if a bridge with the given name exists.
func Exists(name string) bool {
	_, ok := registry[name]
	return ok
}

// BridgeV2Ready returns bridges that have bridgev2 support.
func BridgeV2Ready() []BridgeInfo {
	var ready []BridgeInfo
	for _, info := range List() {
		if info.BridgeV2Ready {
			ready = append(ready, info)
		}
	}
	return ready
}

// External returns bridges that require external subprocess mode.
func External() []BridgeInfo {
	var ext []BridgeInfo
	for _, info := range List() {
		if !info.BridgeV2Ready {
			ext = append(ext, info)
		}
	}
	return ext
}
