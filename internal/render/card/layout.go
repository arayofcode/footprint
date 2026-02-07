package card

type LayoutInput struct {
	StatCount       int
	SectionCount    int
	ShowAllStats    bool
	MinimalSections bool
}

// DecideLayout determines the card's structural properties based on input content.
// Pure function: Input -> LayoutVM.
func DecideLayout(input LayoutInput) LayoutVM {
	// Logic from original renderer:
	// isVertical := !showAllStats && len(activeStats) == 2 && numSections <= 1
	isVertical := !input.ShowAllStats && input.StatCount == 2 && input.SectionCount <= 1

	width := 800
	if isVertical {
		width = 500
	}

	statSpacing := 90
	if isVertical {
		statSpacing = 80
	}

	statRows := 0
	if input.StatCount > 0 {
		if isVertical {
			statRows = input.StatCount
		} else if !input.ShowAllStats && input.StatCount == 4 {
			statRows = 2
		} else {
			statRows = (input.StatCount + 2) / 3
		}
	}

	// Calculate grid height to find where content sections start
	gridHeight := 0
	if input.StatCount > 0 {
		gridHeight = ((statRows - 1) * statSpacing) + 70
	}

	// initial Y for sections
	contentY := 90 + gridHeight + 20

	return LayoutVM{
		Width:       width,
		Height:      0, // Dynamic, calculated later based on section content
		IsVertical:  isVertical,
		StatSpacing: statSpacing,
		StatRows:    statRows,
		ContentY:    contentY,
	}
}
