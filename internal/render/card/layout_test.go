package card

import (
	"testing"
)

func TestDecideLayout(t *testing.T) {
	tests := []struct {
		name     string
		input    LayoutInput
		wantVert bool
		wantW    int
	}{
		{
			name: "Minimal stats + no sections -> Vertical",
			input: LayoutInput{
				StatCount:    2,
				ShowAllStats: false,
				Mode:         LayoutAuto,
				Sections: []SectionLayoutInput{
					{Rows: 1, IsEmpty: false, Placement: StackVertical, Column: 0},
				},
			},
			wantVert: true,
			wantW:    500,
		},
		{
			name: "Full stats + horizontal intent -> Horizontal",
			input: LayoutInput{
				StatCount:    6,
				ShowAllStats: true,
				Mode:         LayoutAuto,
				Sections: []SectionLayoutInput{
					{Rows: 3, IsEmpty: false, Placement: StackHorizontal, Column: 0},
				},
			},
			wantVert: false,
			wantW:    800,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DecideLayout(tt.input)
			if got.IsVertical != tt.wantVert {
				t.Errorf("DecideLayout() IsVertical = %v, want %v", got.IsVertical, tt.wantVert)
			}
			if got.Width != tt.wantW {
				t.Errorf("DecideLayout() Width = %v, want %v", got.Width, tt.wantW)
			}
		})
	}
}

func TestDecideLayout_RespectsPlacement(t *testing.T) {
	// Vertical placement: sections should stack Y
	vInput := LayoutInput{
		Mode: LayoutVertical,
		Sections: []SectionLayoutInput{
			{Rows: 1, IsEmpty: false, Placement: StackVertical, Column: 0},
			{Rows: 1, IsEmpty: false, Placement: StackVertical, Column: 0},
		},
	}
	vLayout := DecideLayout(vInput)
	if vLayout.Sections[0].Y == vLayout.Sections[1].Y {
		t.Error("expected vertical sections to have different Y")
	}
	if vLayout.Sections[0].X != vLayout.Sections[1].X {
		t.Error("expected vertical sections to have same X")
	}

	// Horizontal placement: sections should have same Y, different X
	hInput := LayoutInput{
		Mode: LayoutHorizontal,
		Sections: []SectionLayoutInput{
			{Rows: 1, IsEmpty: false, Placement: StackHorizontal, Column: 0},
			{Rows: 1, IsEmpty: false, Placement: StackHorizontal, Column: 1},
		},
	}
	hLayout := DecideLayout(hInput)
	if hLayout.Sections[0].Y != hLayout.Sections[1].Y {
		t.Error("expected horizontal sections to have same Y")
	}
	if hLayout.Sections[0].X == hLayout.Sections[1].X {
		t.Error("expected horizontal sections to have different X")
	}
}

func TestDecideLayout_HorizontalSections_DoNotOverlap(t *testing.T) {
	input := LayoutInput{
		StatCount:    6,
		ShowAllStats: true,
		Mode:         LayoutHorizontal,
		Sections: []SectionLayoutInput{
			{
				Rows:      3,
				IsEmpty:   false,
				Placement: StackHorizontal,
				Column:    0,
			},
			{
				Rows:      3,
				IsEmpty:   false,
				Placement: StackHorizontal,
				Column:    1,
			},
		},
	}

	layout := DecideLayout(input)

	if len(layout.Sections) != 2 {
		t.Fatalf("expected 2 section layouts, got %d", len(layout.Sections))
	}

	left := layout.Sections[0]
	right := layout.Sections[1]

	// Horizontal sections must align vertically
	if left.Y != right.Y {
		t.Errorf(
			"expected horizontal sections to share Y; got %d and %d",
			left.Y, right.Y,
		)
	}

	// Must be side-by-side
	if left.X == right.X {
		t.Fatalf("horizontal sections overlap: both have X=%d", left.X)
	}

	// Ordering must respect columns
	if right.X <= left.X {
		t.Errorf(
			"expected column 1 to be right of column 0; got %d <= %d",
			right.X, left.X,
		)
	}

	// Hard non-overlap guarantee (geometry contract)
	if right.X-left.X < SectionWidth {
		t.Errorf(
			"horizontal sections overlap in width: delta=%d, expected >= %d",
			right.X-left.X,
			SectionWidth,
		)
	}
}
