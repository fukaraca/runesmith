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
