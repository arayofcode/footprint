package card

type LayoutInput struct {
	StatCount          int
	ShowAllStats       bool
	OwnedRows          int
	ExternalRows       int
	HasOwnedSection    bool
	HasExternalSection bool
	MinimalSections    bool
}

// DecideLayout determines the card's structural properties based on input content.
// Pure function: Input -> LayoutVM.
func DecideLayout(input LayoutInput) LayoutVM {
	// Logic from original renderer:
	// isVertical := !showAllStats && len(activeStats) == 2 && numSections <= 1
	numSections := 0
	if input.HasOwnedSection {
		numSections++
	}
	if input.HasExternalSection {
		numSections++
	}
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

	if input.HasOwnedSection {
		h := 0
		if input.OwnedRows > 0 {
			h = 40 + (input.OwnedRows * rowHeight)
		} else {
			h = 70 // Empty text height
		}

		if isVertical {
			currentY += h
			if input.OwnedRows == 0 {
				// Slight adjustment from original logic: if len(topOwned) == 0 { currentY += 30 }
				// 70 + 30 = 100? No, original was:
				// if len(topOwned) == 0 && !minimalSections { string... }
				// if isVertical { currentY += ...; if len==0 { currentY += 30 }}
				// My h=70 covers the base height, let's add the gap.
				currentY += 30
			}
		} else {
			// Horizontal: We track max height of the row of sections
			// But for horizontal, "currentY" is the start of the section row.
			// We need to add the max section height to currentY at the end.
		}
	}

	if input.HasExternalSection {
		h := 40 + (input.ExternalRows * rowHeight)
		if isVertical {
			currentY += h
		} else {
			// Horizontal calculation
			ownedH := 0
			if input.HasOwnedSection {
				if input.OwnedRows > 0 {
					ownedH = 40 + (input.OwnedRows * rowHeight)
				} else {
					ownedH = 70
				}
			}
			maxH := ownedH
			if h > maxH {
				maxH = h
			}
			// If we haven't added height yet (because isVertical=false), add it now
			// Note: this logic assumes Owned and External are the ONLY sections and are side-by-side in horizontal.
			currentY += maxH
		}
	} else if !isVertical && input.HasOwnedSection {
		// Only owned section in horizontal
		ownedH := 0
		if input.OwnedRows > 0 {
			ownedH = 40 + (input.OwnedRows * rowHeight)
		} else {
			ownedH = 70
		}
		currentY += ownedH
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
