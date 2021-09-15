package gui_opengl

import (
	"fmt"
	"image"
	"image/png"
	"log"
	"math"
	"os"

	"github.com/go-gl/glfw/v3.2/glfw"

	"github.com/inkyblackness/imgui-go/v4"
)

const (
	mouseButtonPrimary   = 0
	mouseButtonSecondary = 1
	mouseButtonTertiary  = 2
	mouseButtonCount     = 3

	// active Monitor for real fullscreen. 0 is always the primary monitor
	activeMonitor = 1
)

type Window struct {
	Window     *glfw.Window
	Title      string
	Width      int
	Height     int
	WinWidth   int
	WinHeight  int
	WinXPos    int
	WinYPos    int
	Fullscreen bool
	Monitors    []*glfw.Monitor
	Monitor    *glfw.Monitor
	ActiveMonitor int
	Vidmode    *glfw.VidMode

	// part of the conversion to imgui
	imguiIO          *imgui.IO
	time             float64
	mouseJustPressed [3]bool
}

func (w *Window) Init() {
	// initialize the library
	err := glfw.Init()
	if err != nil {
		fmt.Printf("FATAL: Failed to initialize glfw")
		panic(err)
	}

	// Set active monitor for real fullscreen display
	w.ActiveMonitor = activeMonitor

	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 4)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, 1)
	// create a window mode and it's OpenGL Context
	w.InitScreen()

}

func (w *Window) InitScreen() {
	var err error
	w.Monitors = glfw.GetMonitors()
	if w.ActiveMonitor >= len(w.Monitors) {
		fmt.Println("Desired Monitor not found. Using default")
		w.ActiveMonitor = 0
	}
	w.Monitor = w.Monitors[w.ActiveMonitor]
	w.Vidmode = w.Monitor.GetVideoMode()
	if w.Fullscreen {
		w.Window, err = glfw.CreateWindow(w.Vidmode.Width, w.Vidmode.Width, w.Title, w.Monitor, nil)
	} else {
		w.Window, err = glfw.CreateWindow(w.Width, w.Height, w.Title, nil, nil)
	}
	if err != nil {
		log.Println("ERROR: glfw Window cannot be created.")
		glfw.Terminate()
		panic(err)
	}

	// Set windows icon - array with multiple sizes can be used
	f, err := os.Open("resource/starex_logo.png")
	if err != nil {
		fmt.Println("WARN: Can't read icon - disregarded.")
	} else {
		img, _ := png.Decode(f)
		w.Window.SetIcon([]image.Image{img})
	}
	defer f.Close()

	// make the window's context current
	w.Window.MakeContextCurrent()
	glfw.SwapInterval(1)
}

func (w *Window) toggleFullscreen(realFullscreen bool) {
	// toggle fullscreen indicator
	w.Fullscreen = !w.Fullscreen
	// if Fullscreen (either "real" or "fullscreen window")
	if w.Fullscreen {
		monResHeight := w.Monitor.GetVideoMode().Height
		monResWidth := w.Monitor.GetVideoMode().Width
		w.WinWidth, w.WinHeight = w.Window.GetSize()
		w.WinXPos, w.WinYPos = w.Window.GetPos()
		w.Width = monResWidth
		w.Height = monResHeight

		if realFullscreen {
			fmt.Println("Trying to set Fullscreen")
			w.Window.SetMonitor(w.Monitor, 0, 0, monResWidth, monResHeight, glfw.DontCare)
		} else {
			fmt.Println("Trying to enter Windowed Full Screen")
			w.Window.SetMonitor(nil, 0, 0, monResWidth, monResHeight, glfw.DontCare)
		}
	} else {
		fmt.Println("Trying to set to Windowed")
		w.Window.SetMonitor(nil, w.WinXPos, w.WinYPos, w.WinWidth, w.WinHeight, glfw.DontCare)
	}

}

func (w *Window) PrepImGUI(io *imgui.IO) {
	w.imguiIO = io

	// prepare keymapping
	w.imguiIO.KeyMap(imgui.KeyTab, int(glfw.KeyTab))
	w.imguiIO.KeyMap(imgui.KeyLeftArrow, int(glfw.KeyLeft))
	w.imguiIO.KeyMap(imgui.KeyRightArrow, int(glfw.KeyRight))
	w.imguiIO.KeyMap(imgui.KeyUpArrow, int(glfw.KeyUp))
	w.imguiIO.KeyMap(imgui.KeyDownArrow, int(glfw.KeyDown))
	w.imguiIO.KeyMap(imgui.KeyPageUp, int(glfw.KeyPageUp))
	w.imguiIO.KeyMap(imgui.KeyPageDown, int(glfw.KeyPageDown))
	w.imguiIO.KeyMap(imgui.KeyHome, int(glfw.KeyHome))
	w.imguiIO.KeyMap(imgui.KeyEnd, int(glfw.KeyEnd))
	w.imguiIO.KeyMap(imgui.KeyInsert, int(glfw.KeyInsert))
	w.imguiIO.KeyMap(imgui.KeyDelete, int(glfw.KeyDelete))
	w.imguiIO.KeyMap(imgui.KeyBackspace, int(glfw.KeyBackspace))
	w.imguiIO.KeyMap(imgui.KeySpace, int(glfw.KeySpace))
	w.imguiIO.KeyMap(imgui.KeyEnter, int(glfw.KeyEnter))
	w.imguiIO.KeyMap(imgui.KeyEscape, int(glfw.KeyEscape))
	w.imguiIO.KeyMap(imgui.KeyA, int(glfw.KeyA))
	w.imguiIO.KeyMap(imgui.KeyC, int(glfw.KeyC))
	w.imguiIO.KeyMap(imgui.KeyV, int(glfw.KeyV))
	w.imguiIO.KeyMap(imgui.KeyX, int(glfw.KeyX))
	w.imguiIO.KeyMap(imgui.KeyY, int(glfw.KeyY))
	w.imguiIO.KeyMap(imgui.KeyZ, int(glfw.KeyZ))

}

// NewFrame marks the begin of a render pass. It forwards all current state to imgui IO.
func (w *Window) NewFrame() {
	// Setup time step
	currentTime := glfw.GetTime()
	if w.time > 0 {
		w.imguiIO.SetDeltaTime(float32(currentTime - w.time))
	}
	w.time = currentTime

	// Setup display size (every frame to accommodate for window resizing)
	w.imguiIO.SetDisplaySize(imgui.Vec2{X: float32(w.Width), Y: float32(w.Height)})


	// Setup inputs
	if w.Window.GetAttrib(glfw.Focused) != 0 {
		x, y := w.Window.GetCursorPos()
		w.imguiIO.SetMousePosition(imgui.Vec2{X: float32(x), Y: float32(y)})
	} else {
		w.imguiIO.SetMousePosition(imgui.Vec2{X: -math.MaxFloat32, Y: -math.MaxFloat32})
	}

	for i := 0; i < len(w.mouseJustPressed); i++ {
		down := w.mouseJustPressed[i] || (w.Window.GetMouseButton(glfwButtonIDByIndex[i]) == glfw.Press)
		w.imguiIO.SetMouseButtonDown(i, down)
		w.mouseJustPressed[i] = false
	}
}

// DisplaySize returns the dimension of the display.
func (w *Window) DisplaySize() [2]float32 {
	return [2]float32{float32(w.Width), float32(w.Height)}
}

// FramebufferSize returns the dimension of the framebuffer.
func (w *Window) FramebufferSize() [2]float32 {
	x, y := w.Window.GetFramebufferSize()
	return [2]float32{float32(x), float32(y)}
}

func (w *Window) Cleanup() {
	defer glfw.Terminate()
}

var glfwButtonIndexByID = map[glfw.MouseButton]int{
	glfw.MouseButton1: mouseButtonPrimary,
	glfw.MouseButton2: mouseButtonSecondary,
	glfw.MouseButton3: mouseButtonTertiary,
}

var glfwButtonIDByIndex = map[int]glfw.MouseButton{
	mouseButtonPrimary:   glfw.MouseButton1,
	mouseButtonSecondary: glfw.MouseButton2,
	mouseButtonTertiary:  glfw.MouseButton3,
}
