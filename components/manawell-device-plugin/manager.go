package main

import (
	"errors"
	"fmt"
	"sync"

	"github.com/fukaraca/runesmith/shared"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

type ManaGer struct {
	mutex       sync.RWMutex
	maxMana     int
	energyType  shared.Elemental
	freeIDs     []string
	allDevices  []*pluginapi.Device // in our case we won't encounter an unhealthy device, so no need to keep track of it
	allocations map[string]AllocationInfo
}

type AllocationInfo struct {
	PodUID    string   `json:"podUID"`
	PodName   string   `json:"podName"`
	Namespace string   `json:"namespace"`
	DeviceIDs []string `json:"deviceIDs"`
	Timestamp int64    `json:"timestamp"`
}

func NewManaGer(cfg ManaConfig) *ManaGer {
	freeIDs := make([]string, cfg.MaxMana)
	allDevices := make([]*pluginapi.Device, cfg.MaxMana)
	for i := 1; i <= cfg.MaxMana; i++ { // TODO what to do on a restart ...
		id := fmt.Sprintf("%s-%03d", cfg.EnergyType, i)
		freeIDs[i-1] = id
		allDevices[i] = &pluginapi.Device{
			ID:     id,
			Health: pluginapi.Healthy,
		}
	}

	return &ManaGer{
		maxMana:     cfg.MaxMana,
		energyType:  cfg.EnergyType,
		freeIDs:     freeIDs,
		allDevices:  allDevices,
		allocations: make(map[string]AllocationInfo),
	}
}

func (m *ManaGer) GetAvailableMana() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.freeIDs)
}

func (m *ManaGer) GetAllocatedMana() int {
	return m.maxMana - m.GetAvailableMana()
}

func (m *ManaGer) GetAllDevices() []*pluginapi.Device {
	return m.allDevices
}

// AllocateDevices is called on  Allocate() by kubelet. We still don't know which pod took which devices
func (m *ManaGer) AllocateDevices(count int) ([]string, error) {
	if count <= 0 {
		return nil, errors.New("count must be > 0")
	}
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if len(m.freeIDs) < count {
		return nil, errors.New("insufficient mana")
	}

	allocatedIDs := make([]string, count)
	for i := 0; i < count; i++ {
		allocatedIDs[i] = m.freeIDs[len(m.freeIDs)-1]
		m.freeIDs = m.freeIDs[:len(m.freeIDs)-1]
	}

	return allocatedIDs, nil
}

// MapAllocations maps devices after pods report themselves
func (m *ManaGer) MapAllocations(podID, podName, namespace string, deviceIDs []string, timestamp int64) {
	if len(deviceIDs) == 0 {
		return
	}
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if v, ok := m.allocations[podID]; ok { // maybe same pod reported again, retry or double container scenarios.
		delete(m.allocations, podID)
		m.freeIDs = append(m.freeIDs, v.DeviceIDs...)
	}

	m.allocations[podID] = AllocationInfo{
		PodUID:    podID,
		PodName:   podName,
		Namespace: namespace,
		DeviceIDs: deviceIDs,
		Timestamp: timestamp,
	}
}

func (m *ManaGer) ReleaseDevices(podUID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	allocation, exists := m.allocations[podUID]
	if !exists {
		return fmt.Errorf("no allocation found for pod UID: %s", podUID)
	}

	m.freeIDs = append(m.freeIDs, allocation.DeviceIDs...)
	delete(m.allocations, podUID)

	return nil
}

func (m *ManaGer) GetAllocation(podUID string) (AllocationInfo, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	allocation, exists := m.allocations[podUID]
	return allocation, exists
}

func (m *ManaGer) GetAllAllocations() map[string]AllocationInfo {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	result := make(map[string]AllocationInfo)
	for k, v := range m.allocations {
		result[k] = v // slice modification may effect underlying array
	}
	return result
}
