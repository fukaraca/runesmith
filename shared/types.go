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
