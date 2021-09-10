package gui_opengl


type Mouse struct {
	B1Pressed bool
	B2Pressed bool
	BPressedX int	
	BPressedY int
	X int 
	Y int
	Pan bool
}

func (m *Mouse) pressButton(button int) {
	if button == 1 {
		m.B1Pressed = true
	} else {
		m.B2Pressed = true
	}
	m.BPressedX = m.X
	m.BPressedY = m.Y
}

func (m *Mouse) releaseButton(button int) {
	if button == 1 {
		m.B1Pressed = false
	} else {
		m.B2Pressed = false
	}
	m.BPressedX = -1
	m.BPressedY = -1
}

func (m *Mouse) move(x int, y int) {
	m.X = x
	m.Y = y
}
