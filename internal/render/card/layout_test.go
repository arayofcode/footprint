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
				StatCount:       2,
				HasOwnedSection: true,
				ShowAllStats:    false,
			},
			wantVert: true,
			wantW:    500,
		},
		{
			name: "Full stats -> Horizontal",
			input: LayoutInput{
				StatCount:       6,
				HasOwnedSection: true,
				ShowAllStats:    true,
			},
			wantVert: false,
			wantW:    800,
		},
		{
			name: "Minimal stats + 2 sections -> Horizontal",
			input: LayoutInput{
				StatCount:          2,
				HasOwnedSection:    true,
				HasExternalSection: true,
				ShowAllStats:       false,
			},
			wantVert: false,
			wantW:    800,
		},
		{
			name: "4 stats minimal -> Horizontal 2x2",
			input: LayoutInput{
				StatCount:       4,
				HasOwnedSection: true,
				ShowAllStats:    false,
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
