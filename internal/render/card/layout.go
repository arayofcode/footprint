package card

type SectionLayoutInput struct {
	Rows    int
	IsEmpty bool
}

type LayoutInput struct {
	StatCount    int
	ShowAllStats bool
	Sections     []SectionLayoutInput
}

// DecideLayout determines the card's structural properties based on input content.
// Pure function: Input -> LayoutVM.
func DecideLayout(input LayoutInput) LayoutVM {
	numSections := len(input.Sections)
	isVertical := !input.ShowAllStats && input.StatCount == 2 && numSections <= 1

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

	// Calculate Section Heights
	rowHeight := 55 // Horizontal default
	if isVertical {
		rowHeight = 65
	}

	currentY := contentY
	sectionGap := 0

	if isVertical {
		for _, s := range input.Sections {
			h := 0
			if !s.IsEmpty {
				h = 40 + (s.Rows * rowHeight)
			} else {
				h = 70 // Empty text height
			}
			currentY += h
			if s.IsEmpty {
				currentY += 30
			}
		}
	} else {
		// Horizontal: Sections are side-by-side (max 2 supported in this specific layout style)
		maxSectionH := 0
		for _, s := range input.Sections {
			h := 0
			if !s.IsEmpty {
				h = 40 + (s.Rows * rowHeight)
			} else {
				h = 70
			}
			if h > maxSectionH {
				maxSectionH = h
			}
		}
		currentY += maxSectionH
	}

	totalHeight := currentY + 50

	return LayoutVM{
		Width:       width,
		Height:      totalHeight,
		IsVertical:  isVertical,
		StatSpacing: statSpacing,
		StatRows:    statRows,
		ContentY:    contentY,
		RowHeight:   rowHeight,
		SectionGap:  sectionGap,
	}
}
