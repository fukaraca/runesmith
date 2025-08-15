package shared

const (
	FireEnergy      Elemental = "fire"
	FrostEnergy     Elemental = "frost"
	ArcaneElemental Elemental = "arcane"
)

type Elemental string

func (e Elemental) String() string {
	return string(e)
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

type ArtifactStatus string

const (
	ScheduledAS   ArtifactStatus = "Scheduled"
	QueuedAS      ArtifactStatus = "Queued"
	PreemptedAS   ArtifactStatus = "Preempted"
	PrioritizedAS ArtifactStatus = "Prioritized"
	EnchantingAS  ArtifactStatus = "Enchanting"
	CompletedAS   ArtifactStatus = "Completed"
	DeletedAS     ArtifactStatus = "Deleted"
)

type NodeStatus struct {
	Name        string
	Available   int
	Allocated   int
	Healthy     bool
	RunningJobs int
}
