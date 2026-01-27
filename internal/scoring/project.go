package scoring

import (
	"github.com/arayofcode/footprint/internal/domain"
)

func (c *Calculator) ScoreOwnedProject(project domain.OwnedProject) domain.OwnedProject {
	// Weighted ownership impact based on project popularity
	project.Score = OwnershipScore * project.PopularityMultiplier(c.Clamp)
	return project
}
