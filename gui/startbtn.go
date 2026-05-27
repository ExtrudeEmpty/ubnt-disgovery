package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var (
	btnStartNormal = color.NRGBA{R: 0x16, G: 0xa3, B: 0x4a, A: 0xff} // Grün
	btnStartHover  = color.NRGBA{R: 0x0f, G: 0x7a, B: 0x35, A: 0xff} // Dunkelgrün
	btnStopNormal  = color.NRGBA{R: 0xdc, G: 0x26, B: 0x26, A: 0xff} // Rot
	btnStopHover   = color.NRGBA{R: 0xa8, G: 0x12, B: 0x12, A: 0xff} // Dunkelrot
)

// ColorButton ist ein benutzerdefinierter Button mit Grün/Rot-Zuständen
// und dunkleren Hover-Varianten statt der generischen Theme-Farbe.
type ColorButton struct {
	widget.BaseWidget
	Text     string
	IsStop   bool
	OnTapped func()
	hovered  bool
}

func NewColorButton(text string, onTapped func()) *ColorButton {
	b := &ColorButton{Text: text, OnTapped: onTapped}
	b.ExtendBaseWidget(b)
	return b
}

func (b *ColorButton) SetButtonText(text string) {
	b.Text = text
	b.Refresh()
}

func (b *ColorButton) SetIsStop(stop bool) {
	b.IsStop = stop
	b.Refresh()
}

func (b *ColorButton) Tapped(*fyne.PointEvent) {
	if b.OnTapped != nil {
		b.OnTapped()
	}
}

func (b *ColorButton) MouseIn(*desktop.MouseEvent)    { b.hovered = true; b.Refresh() }
func (b *ColorButton) MouseOut()                      { b.hovered = false; b.Refresh() }
func (b *ColorButton) MouseMoved(*desktop.MouseEvent) {}

func (b *ColorButton) Cursor() desktop.Cursor { return desktop.PointerCursor }

func (b *ColorButton) CreateRenderer() fyne.WidgetRenderer {
	bg := canvas.NewRectangle(btnStartNormal)
	bg.CornerRadius = 4
	txt := canvas.NewText(b.Text, color.White)
	txt.Alignment = fyne.TextAlignCenter
	txt.TextSize = theme.TextSize()
	r := &colorBtnRenderer{btn: b, bg: bg, txt: txt}
	r.Refresh()
	return r
}

type colorBtnRenderer struct {
	btn *ColorButton
	bg  *canvas.Rectangle
	txt *canvas.Text
}

func (r *colorBtnRenderer) Layout(size fyne.Size) {
	r.bg.Move(fyne.NewPos(0, 0))
	r.bg.Resize(size)
	ts := fyne.MeasureText(r.txt.Text, r.txt.TextSize, r.txt.TextStyle)
	r.txt.Move(fyne.NewPos(0, (size.Height-ts.Height)/2))
	r.txt.Resize(fyne.NewSize(size.Width, ts.Height))
}

func (r *colorBtnRenderer) MinSize() fyne.Size {
	ts := fyne.MeasureText(r.txt.Text, r.txt.TextSize, r.txt.TextStyle)
	p := theme.Padding()
	return fyne.NewSize(ts.Width+p*4, ts.Height+p*2)
}

func (r *colorBtnRenderer) Refresh() {
	r.txt.Text = r.btn.Text
	r.txt.TextSize = theme.TextSize()
	if r.btn.IsStop {
		if r.btn.hovered {
			r.bg.FillColor = btnStopHover
		} else {
			r.bg.FillColor = btnStopNormal
		}
	} else {
		if r.btn.hovered {
			r.bg.FillColor = btnStartHover
		} else {
			r.bg.FillColor = btnStartNormal
		}
	}
	r.bg.Refresh()
	r.txt.Refresh()
}

func (r *colorBtnRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.bg, r.txt}
}

func (r *colorBtnRenderer) Destroy() {}
