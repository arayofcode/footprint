package card

import (
	"github.com/arayofcode/footprint/internal/domain"
)

type RowKind int

const (
	RowOwnedProject RowKind = iota
	RowExternalContribution
)

type CardViewModel struct {
	Width      int
	Height     int
	User       UserVM
	Stats      []StatVM
	Sections   []SectionVM
	Footer     FooterVM
	Layout     LayoutVM // Embedding full layout details for renderer access
	IsVertical bool     // Kept for top-level convenience (or can use Layout.IsVertical)
}

type UserVM struct {
	Username  string
	AvatarKey domain.AssetKey
}

type StatVM struct {
	Label string
	Value string
	Icon  string
	Raw   int
	X     int
	Y     int
	Color string
}

type SectionVM struct {
	Title        string
	EmptyMessage string // Message to show if Rows is empty. If "", section is hidden or just empty space? (Logic moved to VM/Layout)
	X            int
	Y            int
	Rows         []SectionRowVM
}

type SectionRowVM struct {
	Kind      RowKind
	Title     string
	Subtitle  string
	Link      string
	AvatarKey domain.AssetKey
	Badges    []BadgeVM
}

type BadgeVM struct {
	Icon  string
	Count string
	Link  string
}

type FooterVM struct {
	Y           int
	GeneratedAt string
	Attribution string
}

type LayoutVM struct {
	Width       int
	Height      int
	IsVertical  bool
	StatSpacing int
	RowHeight   int
	SectionGap  int
	ContentY    int
	StatRows    int
}
