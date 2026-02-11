package scoring

import (
	"github.com/arayofcode/footprint/internal/domain"
)

func (c *Calculator) EnrichOwnedProject(project domain.OwnedProject) domain.EnrichedProject {
	return domain.EnrichedProject{
		OwnedProject:  project,
		BaseScore:     OwnershipScore,
		PopularityRaw: project.PopularityMultiplier(),
	}
}
