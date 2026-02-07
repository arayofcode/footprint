package card

type SectionPlacement int

const (
	StackVertical SectionPlacement = iota
	StackHorizontal
)

type LayoutMode int

const (
	LayoutAuto LayoutMode = iota
	LayoutVertical
	LayoutHorizontal
)

type SectionLayoutInput struct {
	Rows      int
	IsEmpty   bool
	Placement SectionPlacement
	Column    int // 0-indexed column
}

type LayoutInput struct {
	StatCount    int
	ShowAllStats bool
	Mode         LayoutMode
	Sections     []SectionLayoutInput
}

const (
	SectionHeaderHeight  = 40
	SectionRowHeight     = 55
	EmptyStateHeight     = 70
	EmptyStatePadding    = 30
	StatBoxHeight        = 70
	StatBoxSpacing       = 90
	VerticalStatSpacing  = 80
	HeaderMargin         = 90
	ContentMargin        = 20
	FooterHeight         = 50
	LandscapeWidth       = 800
	VerticalWidth        = 500
	SectionWidth         = 340
	VerticalSectionWidth = 420
)

// DecideLayout determines the card's structural properties based on input content.
// Pure function: Input -> LayoutVM.
func DecideLayout(input LayoutInput) LayoutVM {
	numSections := len(input.Sections)

	// Resolve Layout Mode
	isVertical := false
	switch input.Mode {
	case LayoutVertical:
		isVertical = true
	case LayoutHorizontal:
		isVertical = false
	case LayoutAuto:
		// Vertical if specifically 2 stats and 0 or 1 sections (Legacy heuristic)
		isVertical = !input.ShowAllStats && input.StatCount == 2 && numSections <= 1
	}

	width := LandscapeWidth
	if isVertical {
		width = VerticalWidth
	}

	statSpacing := StatBoxSpacing
	if isVertical {
		statSpacing = VerticalStatSpacing
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
		gridHeight = ((statRows - 1) * statSpacing) + StatBoxHeight
	}

	contentY := HeaderMargin + gridHeight + ContentMargin

	rowHeight := SectionRowHeight
	if isVertical {
		rowHeight = 65 // Specific to vertical layout
	}

	currentY := contentY
	var sectionLayouts []SectionLayoutVM

	// Track max height of horizontal rows to correctly increment Y
	maxRowH := 0

	for _, s := range input.Sections {
		h := 0
		if !s.IsEmpty {
			h = SectionHeaderHeight + (s.Rows * rowHeight)
		} else {
			h = EmptyStateHeight
		}

		xPos := 40 // Default X
		yPos := currentY

		// If the entire card is in vertical mode, we ignore horizontal placement intent
		if isVertical {
			currentY += h
			if s.IsEmpty {
				currentY += EmptyStatePadding
			}
		} else if s.Placement == StackHorizontal {
			// Explicit column positioning
			xPos = 40 + (s.Column * 380)

			// Horizontal sections in the same row share the same Y
			if h > maxRowH {
				maxRowH = h
			}
		} else {
			// StackVertical in a non-vertical card (full width stack)
			currentY += h
			if s.IsEmpty {
				currentY += EmptyStatePadding
			}
		}

		sectionLayouts = append(sectionLayouts, SectionLayoutVM{X: xPos, Y: yPos})
	}

	// Add max height from horizontal sections if any
	currentY += maxRowH

	totalHeight := currentY + FooterHeight

	return LayoutVM{
		Width:       width,
		Height:      totalHeight,
		IsVertical:  isVertical,
		StatSpacing: statSpacing,
		StatRows:    statRows,
		ContentY:    contentY,
		RowHeight:   rowHeight,
		Sections:    sectionLayouts,
	}
}
