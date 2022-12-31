package common

import zone "github.com/lrstanley/bubblezone"

// Common is a struct all components should embed
type Common struct {
	Width  int
	Height int
	Zone   *zone.Manager
}

// SetSize sets the width and height of the common struct
func (c *Common) SetSize(height, width int) {
	c.Width = width
	c.Height = height
}
