package card

type CardViewModel struct {
	Width      int
	Height     int
	User       UserVM
	Stats      []StatVM
	Sections   []SectionVM
	Footer     FooterVM
	IsVertical bool
}

type UserVM struct {
	Username  string
	AvatarURL string // Data URL
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
	Title   string
	X       int
	Y       int
	Content string // Pre-rendered SVG content for the section body (rows)
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
	StatRows    int
	ContentY    int // Y position where sections start
}
