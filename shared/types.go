package shared

import "strings"

const (
	FireEnergy   Elemental = "fire"
	FrostEnergy  Elemental = "frost"
	ArcaneEnergy Elemental = "arcane"
)

type Elemental string

func (e Elemental) String() string {
	return string(e)
}

func (e Elemental) Resource() Resource {
	switch e {
	case FireEnergy:
		return FireResource
	case FrostEnergy:
		return FrostResource
	case ArcaneEnergy:
		return ArcaneResource
	}
	return ""
}

const (
	FireResource   Resource = "manawell.io/fire"
	FrostResource  Resource = "manawell.io/frost"
	ArcaneResource Resource = "manawell.io/arcane"
)

type Resource string

func (r Resource) String() string {
	return string(r)
}

type AllocationInfo struct {
	PodUID    string   `json:"podUID"`
	PodName   string   `json:"podName"`
	Namespace string   `json:"namespace"`
	DeviceIDs []string `json:"deviceIDs"`
	Timestamp int64    `json:"timestamp"`
}

type Tier string

const (
	Common    Tier = "Common"
	Rare      Tier = "Rare"
	Epic      Tier = "Epic"
	Legendary Tier = "Legendary"
)

func (t Tier) Lower() string {
	return strings.ToLower(string(t))
}

type Requirements struct {
	Fire   int `json:"fire"`
	Frost  int `json:"frost"`
	Arcane int `json:"arcane"`
}

type MagicalItem struct {
	ID           int          `json:"id"`
	Name         string       `json:"name"`
	Tier         Tier         `json:"tier"`
	Requirements Requirements `json:"requirements"`
	Priority     int          `json:"priority"`
}

func (i MagicalItem) RequiredList() map[Elemental]int {
	m := make(map[Elemental]int)
	if i.Requirements.Fire > 0 {
		m[FireEnergy] = i.Requirements.Fire
	}
	if i.Requirements.Frost > 0 {
		m[FrostEnergy] = i.Requirements.Frost
	}
	if i.Requirements.Arcane > 0 {
		m[ArcaneEnergy] = i.Requirements.Arcane
	}
	return m
}

type EnchantmentPhase string

const (
	ScheduledAS   EnchantmentPhase = "Scheduled"
	RequeuedAS    EnchantmentPhase = "Requeued"
	PreemptedAS   EnchantmentPhase = "Preempted"
	PrioritizedAS EnchantmentPhase = "Prioritized"
	EnchantingAS  EnchantmentPhase = "Enchanting"
	CompletedAS   EnchantmentPhase = "Completed"
	FailedAS      EnchantmentPhase = "Failed"
	DeletedAS     EnchantmentPhase = "Deleted"
)

func (p EnchantmentPhase) String() string {
	return string(p)
}

func (p EnchantmentPhase) Ptr() *EnchantmentPhase {
	return &p
}

type NodeStatus struct {
	Name        string
	Available   int
	Allocated   int
	Healthy     bool
	RunningJobs int
}
