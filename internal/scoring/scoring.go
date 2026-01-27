package scoring

const (
	DefaultClamp   = 10.0
	MergedPRBonus  = 1.5
	OwnershipScore = 2500.0
)

type Calculator struct {
	Clamp float64
}

func NewCalculator(clamp float64) *Calculator {
	if clamp <= 0 {
		clamp = DefaultClamp
	}
	return &Calculator{Clamp: clamp}
}
