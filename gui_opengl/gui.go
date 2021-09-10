package gui_opengl

import (
	"fmt"
	//	"io"
	"bytes"
	"io/ioutil"
	"log"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"

	"time"

	"image"
	"image/png"

	"github.com/go-gl/gl/v4.4-core/gl"

	"github.com/go-gl/glfw/v3.2/glfw"

	"github.com/engoengine/glm"
	"github.com/inkyblackness/imgui-go/v4"

	"github.com/Jest0r/starex_go/galaxy"
	"github.com/Jest0r/starex_go/renderers"
)

const (
//	ScreenWidth      = 1700
//	ScreenHeight     = 1000
	ScreenWidth      = 1200
	ScreenHeight     = 800
	FullRotationTime = 15.0
	CamX             = -20000
	CamY             = -20000
	CamZ             = 10000
	CamViewAngle     = 30.0 // NOT focal length
	//CamViewAngle     = 35.0 // NOT focal length
	CamAngleA  = 0
	CamAngleB  = 30
	MinCamDist = 0.05

	InitialZoom          = 1
	ZoomIncrement        = 0.01
	RotateIncrement      = 0.01
	MousePanSensitivityX = 0.005
	MousePanSensitivityY = -0.005

	SceneNear = 0.1
	SceneFar  = 100

	SaturationMult = 2.0
	SaturationMod  = 0.2

	// turn off if 0
	// vsync of my monitor is 60, so this will limit the CPU used
	FrameRateLimit = 100

	Bloom = true

	MaxLogScrollback = 50
)

type Gui struct {
	Win   Window
	Cam   Camera
	Scene Scene
	Persp Perspective

	// --- ImGUI stuff
	ImGUIContext  imgui.Context
	ImGUIIO       imgui.IO
	ImGUIRenderer *renderers.OpenGL3

	// --- Shaders
	Shader           Shader // Shader for simple (non blurred) display
	BloomStep1Shader Shader // Step 1 for enabled lighting effect
	BlurShader       Shader // Step 2 for enabled lighting effect, called multiple times
	BloomShader      Shader // Step 3 and final step for enabled lighting effects

	// --- Shader options
	uBrightThreshold float32
	uExposure        float32
	uBloomBlur       int32
	uWeights         [][]float32
	uActiveWeight    int32
	blurSteps        int
	uSaturationMult  float32

	showDisplayTitle bool

	// --- Display stuff
	autoRotate bool

	// -----------
	Galaxy               *galaxy.Galaxy
	pause                bool
	DegPerSecond         float32
	Mouse                Mouse
	texBuf               uint32
	colorBuf             uint32
	colorBuffers         []uint32
	pingpongFBOs         []uint32
	pingpongColorBuffers []uint32
	fbo                  uint32
	vbo                  uint32
	hdrFBO               uint32
	BloomActive          bool

	// logging
	rawLog      *bytes.Buffer
	stringLog   []string
	InfoLogger  *log.Logger
	WarnLogger  *log.Logger
	ErrorLogger *log.Logger

	// ---- mainloop stuff
	rotPerFrameA float32
	lastTime     time.Time
	nbFrames     int
	// 		framerate stuff
	tick           time.Duration
	frameRateLimit int
	msPerFrame     float32
	Fps            int

	// ---- game flags
	displayGalaxy bool

	GameExitRequested bool
}

func (g *Gui) setCallbacks() {
	g.Win.Window.SetSizeCallback(g.windowSizeCallback)
	g.Win.Window.SetKeyCallback(g.keyCallback)
	g.Win.Window.SetMouseButtonCallback(g.mouseCallback)
	g.Win.Window.SetPosCallback(g.posCallback)
	g.Win.Window.SetCursorPosCallback(g.mousePosCallback)
	g.Win.Window.SetScrollCallback(g.scrollCallback)
}

func (g *Gui) Cleanup() {
	g.InfoLogger.Println("Cleaning up...")
	defer g.ImGUIContext.Destroy()
	defer g.ImGUIRenderer.Dispose()
	defer g.Win.Cleanup()
}

func (g *Gui) Pause() {
	g.WarnLogger.Println("----- PAUSE -----")
	for g.pause {
		time.Sleep(100 * time.Millisecond)
		glfw.PollEvents()
	}

}

// ----------- Callbacks --------------
func (g *Gui) windowSizeCallback(window *glfw.Window, width int, height int) {
	g.Win.Width = width
	g.Win.Height = height
	g.InfoLogger.Println("Adjusting window")

	gl.Viewport(0, 0, int32(width), int32(height))
	g.Persp.AspectRatio = float32(g.Win.Width) / float32(g.Win.Height)
	g.PrepareScene()
}

func (g *Gui) mouseCallback(window *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {
	if action == glfw.Press {
		switch button {
		case glfw.MouseButton1:
			g.Mouse.pressButton(1)
		case glfw.MouseButton2:
			g.Mouse.pressButton(2)
			g.autoRotate = false
			g.Mouse.Pan = true
		}
	} else if action == glfw.Release {
		switch button {
		case glfw.MouseButton1:
			g.Mouse.releaseButton(1)
		case glfw.MouseButton2:
			g.Mouse.releaseButton(2)
			g.Mouse.Pan = false
			g.autoRotate = true
		}
	}
}

func (g *Gui) mousePosCallback(window *glfw.Window, xpos float64, ypos float64) {
	if g.Mouse.Pan {
		xdelta := (float64(g.Mouse.X) - xpos) * MousePanSensitivityX
		ydelta := (float64(g.Mouse.Y) - ypos) * MousePanSensitivityY
		g.Cam.SetPositionRadial(g.Cam.Dist, g.Cam.A+float32(xdelta), g.Cam.B+float32(ydelta))
	}
	g.Mouse.move(int(xpos), int(ypos))

}

// called when window is moved
func (g *Gui) posCallback(window *glfw.Window, xpos int, ypos int) {
}

func (g *Gui) scrollCallback(window *glfw.Window, xpos float64, ypos float64) {
	zi := float32(ZoomIncrement)
	g.Cam.Dist += zi * float32(ypos) * 5
	if g.Cam.Dist <= MinCamDist {
		g.Cam.Dist = MinCamDist
	}
	g.DegPerSecond += zi * float32(xpos) * 5
}

func (g *Gui) keyCallback(window *glfw.Window, key glfw.Key, scancode int, keyAction glfw.Action, mods glfw.ModifierKey) {
	if keyAction == glfw.Press {
		// if in title screen, exit.
		if g.showDisplayTitle {
			g.showDisplayTitle = false
		}
		switch key {
		case glfw.KeyF11:
			if mods == glfw.ModShift {
				g.Win.toggleFullscreen(true)
			} else {
				g.Win.toggleFullscreen(false)
			}
//			g.toggleFullscreen()
		case glfw.KeyEscape:
			g.Win.Window.SetShouldClose(true)
		case glfw.KeySpace:
			g.togglePause()
		case glfw.KeyF1:
			g.SaveImage("galaxy.png", g.Win.Width, g.Win.Height)
			g.SaveBuffer(gl.COLOR_ATTACHMENT0, "col_a0.png", g.Win.Width, g.Win.Height)
			g.SaveBuffer(gl.COLOR_ATTACHMENT1, "col_a1.png", g.Win.Width, g.Win.Height)
		// --- zooming
		case glfw.KeyW:
			g.Cam.B += ZoomIncrement
		case glfw.KeyS:
			g.Cam.B -= ZoomIncrement
		// --- saturation
		case glfw.KeyY:
			if g.uSaturationMult < 3.0 {
				g.uSaturationMult += SaturationMod
			}
		case glfw.KeyH:
			if g.uSaturationMult > SaturationMod {
				g.uSaturationMult -= SaturationMod
			}
		// --- brightness threshold
		case glfw.KeyR:
			g.uBrightThreshold += 0.1
		case glfw.KeyF:
			if g.uBrightThreshold > 0.1 {
				g.uBrightThreshold -= 0.1
			}
		// --- exposure
		case glfw.KeyT:
			g.uExposure += 0.1
		case glfw.KeyG:
			if g.uExposure > 0.1 {
				g.uExposure -= 0.1
			}
		// --- bloom / blur options
		case glfw.KeyB:
			g.BloomActive = !g.BloomActive
			//		case glfw.KeyV:

			//			g.uBloomBlur = 1 - g.uBloomBlur
		case glfw.KeyC:
			g.uActiveWeight = (g.uActiveWeight + 1) % int32(len(g.uWeights))
			g.InfoLogger.Println("Active Weight", g.uActiveWeight)
		case glfw.KeyX:
			g.blurSteps = (g.blurSteps + 2) % 10
			g.InfoLogger.Println("Blur Steps:", g.blurSteps)
		// --- zooming and rotating
		case glfw.KeyUp:
			g.Cam.Dist -= ZoomIncrement
			if g.Cam.Dist <= MinCamDist {
				g.Cam.Dist = MinCamDist
			}
		case glfw.KeyDown:
			g.Cam.Dist += ZoomIncrement
		case glfw.KeyLeft:
			g.DegPerSecond -= ZoomIncrement
		case glfw.KeyRight:
			g.DegPerSecond += ZoomIncrement

		}
	} else if keyAction == glfw.Repeat {
		switch key {
		case glfw.KeyW:
			g.Cam.B += ZoomIncrement
		case glfw.KeyS:
			g.Cam.B -= ZoomIncrement
		case glfw.KeyUp:
			g.Cam.Dist -= ZoomIncrement
			if g.Cam.Dist <= MinCamDist {
				g.Cam.Dist = MinCamDist
			}
		case glfw.KeyDown:
			g.Cam.Dist += ZoomIncrement
		case glfw.KeyLeft:
			g.DegPerSecond -= ZoomIncrement
		case glfw.KeyRight:
			g.DegPerSecond += ZoomIncrement
		}
	}
}

// ----------- File Management --------------

func (g *Gui) SaveImage(filename string, width int, height int) {
	// ReadPixels has to happen in main thread
	im := image.NewNRGBA(image.Rect(0, 0, width, height))
	gl.ReadPixels(0, 0, int32(width), int32(height), gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(im.Pix))
	// saving in new thread to keep frame rate drop as brief as possible
	go g.threaded_save(filename, im, width, height)
}

func (g *Gui) SaveBuffer(bufname uint32, filename string, width int, height int) {
	// ReadPixels has to happen in main thread
	gl.ReadBuffer(bufname)
	im := image.NewNRGBA(image.Rect(0, 0, width, height))
	gl.ReadPixels(0, 0, int32(width), int32(height), gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(im.Pix))
	// saving in new thread to keep frame rate drop as brief as possible
	go g.threaded_save(filename, im, width, height)
}

func (g *Gui) threaded_save(filename string, im *image.NRGBA, width int, height int) {
	// GL images are flipped horizontally. Flipping it back
	flippedim := image.NewNRGBA(image.Rect(0, 0, width, height))
	for row := 0; row < height; row += 1 {
		for col := 0; col < width*4; col += 1 {
			flippedim.Pix[row*width*4+col] = im.Pix[(height-row-1)*width*4+col]
			// alpha value of every pixel to 1
			if (row*width*4+col)%4 == 3 {
				flippedim.Pix[row*width*4+col] = 255
			}
		}
	}
	// crating file...
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	// encode image to png and save
	if err := png.Encode(f, flippedim); err != nil {
		f.Close()
		log.Fatal(err)
	}
	// ... and close the file
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
	g.InfoLogger.Println("Image save complete.")
}

// ----------- Toggles --------------

func (g *Gui) togglePause() {
	g.pause = !g.pause
	if g.pause {
		g.Pause()
	}
}

func (g *Gui) displayLog() {
	var textCol imgui.Vec4
	imgui.SetNextWindowPos(imgui.Vec2{X: 10, Y: float32(g.Win.Height) - 150})
	imgui.SetNextWindowSize(imgui.Vec2{X: float32(g.Win.Width) / 2, Y: 150})
	imgui.Begin("Log")

	// adding raw logs to string log, and truncate raw log
	if g.rawLog.Len() > 0 {
		g.stringLog = append(g.stringLog, strings.Split(g.rawLog.String(), "\n")...)
		g.rawLog.Truncate(0)

	}

	// cap string log to max scrollback
	if len(g.stringLog) > MaxLogScrollback {
		g.stringLog = g.stringLog[len(g.stringLog)-MaxLogScrollback:]
	}

	// display lines
	for _, line := range g.stringLog {
		if len(line) > 0 {
			switch line[0] {
			case 'W':
				textCol = imgui.Vec4{X: 0.8, Y: 0.8, Z: 0, W: 1}
			case 'E':
				textCol = imgui.Vec4{X: 1, Y: 0.2, Z: 0, W: 1}
			case 'I':
				textCol = imgui.Vec4{X: 0.8, Y: 0.8, Z: 0.8, W: 1}
			default:
				textCol = imgui.Vec4{X: 1, Y: 1, Z: 1, W: 1}

			}
			imgui.PushStyleColor(imgui.StyleColorText, textCol)
			imgui.Text(line)
			imgui.PopStyleColor()

		}
	}
	imgui.SetScrollHereY(1)
	imgui.End()
}

// ---------- Game States -------------
func (g *Gui) DisplayTitle() {
	g.showDisplayTitle = true
}

func (g *Gui) displayTitle() {
	ttext, err := ioutil.ReadFile("resource/logo.txt")
	if err != nil {
		g.ErrorLogger.Println("Can't open Logo File")
	}

	imgui.SetNextWindowPos(imgui.Vec2{X: float32(g.Win.Width)/2 - 200, Y: float32(g.Win.Height)/2 - 100})
	imgui.SetNextWindowSize(imgui.Vec2{X: 400, Y: 200})
	imgui.BeginV("Stats", nil, imgui.WindowFlagsNoResize|imgui.WindowFlagsNoMove|imgui.WindowFlagsNoTitleBar)
	imgui.Text("                   Welcome to\n\n")
	imgui.Text(string(ttext))
	imgui.End()

	// should close by any keystroke
}

func (g *Gui) DisplayMainMenu() {
	g.InfoLogger.Println("Displaying Main Menu")

}
func (g *Gui) CreateGalaxy() {
	g.InfoLogger.Println("Creating Galaxy")

}

func (g *Gui) LoadGalaxy() {
	g.InfoLogger.Println("LoadingGalaxy")

}
func (g *Gui) DisplayGalaxy() {
	g.displayGalaxy = true

}

func (g *Gui) Init() {
	// Init some vars
	// 		window stuff
	g.Win.Title = "Starex Starfield Visualizer (openGL)"
	g.Win.Height = int(ScreenHeight)
	g.Win.Width = int(ScreenWidth)

	g.BloomActive = Bloom

	// neccessary, otherwise everything breaks.
	runtime.LockOSThread()

	// IMGUI stuff
	g.ImGUIContext = *imgui.CreateContext(nil)
	g.ImGUIIO = imgui.CurrentIO()

	g.autoRotate = true

	g.Win.Init()
	// Init OpenGL
	err := gl.Init()
	if err != nil {
		panic("Init error! - " + err.Error())
	}

	// logging to buffer, so it can be used for display
	g.stringLog = []string{}
	g.rawLog = new(bytes.Buffer)
	log.SetOutput(g.rawLog)
	g.InfoLogger = log.New(g.rawLog, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	g.WarnLogger = log.New(g.rawLog, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	g.ErrorLogger = log.New(g.rawLog, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	GlClearError()

	g.ImGUIRenderer, err = renderers.NewOpenGL3(g.ImGUIIO)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(-1)
	}
	GlCheckError("Init OpenGL")

	g.Win.PrepImGUI(&g.ImGUIIO)
	g.setCallbacks()

	// Enable Texture
	gl.Enable(gl.TEXTURE_2D)

	// Enable Blending
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE)

	// get opengl and glsl version
	openglVer := gl.GoStr(gl.GetString(gl.VERSION))
	g.InfoLogger.Println("OpenGL version", openglVer)
	glgsVer := strings.Split(gl.GoStr(gl.GetString(gl.SHADING_LANGUAGE_VERSION)), ".")
	g.InfoLogger.Println("GLSL version:", gl.GoStr(gl.GetString(gl.SHADING_LANGUAGE_VERSION)))

	// Loading Shaders
	if gv, _ := strconv.Atoi(glgsVer[0]); gv < 3 {
		g.WarnLogger.Println("Old glgs version - using legacy shader")
		g.Shader.Init("shaders/legacy.glsl")
	} else {
		g.Shader.Init("shaders/experimental.glsl")
	}

	// get Uniform loc
	g.Shader.GetUniformLoc("uMVP")

	g.uSaturationMult = SaturationMult
	// specific shader just for bloom effect
	g.BloomStep1Shader.Init("shaders/bloom_step1.glsl")
	//	g.BloomStep1Shader.SetFloat("uSaturationMult", g.uSaturationMult)

	g.BlurShader.Init("shaders/blur.glsl")
	// init different blurring weights.
	g.uWeights = append(g.uWeights, []float32{0.1216216, 0.054054, 0.016216})
	g.uWeights = append(g.uWeights, []float32{0.1216216, 0.054054, 0.016216, 0.008})
	//g.uWeights = append(g.uWeights, []float32{0.1216216, 0.054054, 0.016216, 0.012, 0.008, 0.005})
	g.uWeights = append(g.uWeights, []float32{0.1216216, 0.054054, 0.016216, 0.012, 0.008, 0.005, 0, 0})
	g.uWeights = append(g.uWeights, []float32{0.1216216, 0.054054, 0.016216, 0.012, 0.008, 0, 0, 0, 0.005})
	g.uActiveWeight = 0

	g.BloomShader.Init("shaders/bloom.glsl")
	g.uBrightThreshold = 0.0
	g.uExposure = 1.0
	g.uBloomBlur = 1

	g.blurSteps = 6

	//--- used by mainloop
	//		angle stuff
	// timing stuff
	g.lastTime = time.Now()
	g.DegPerSecond = float32(-2 * math.Pi / FullRotationTime)
	// 		framerate stuff
	g.frameRateLimit = FrameRateLimit
	if g.frameRateLimit > 0 {
		g.tick = time.Duration(1000000/g.frameRateLimit) * time.Microsecond
	}

	g.displayGalaxy = true
}

func (g *Gui) PrepareScene() {
	start := time.Now()

	g.BloomShader.Use()
	g.BloomShader.GetUniformLoc("uImage")
	g.BloomShader.GetUniformLoc("uSatMult")
	g.BloomShader.SetInt("bloom", 1)
	g.BloomShader.SetInt("scene", 0)

	g.BloomStep1Shader.Use()
	g.BloomStep1Shader.GetUniformLoc("uMVP")
	g.BloomStep1Shader.GetUniformLoc("uSatMult")

	g.BlurShader.Use()
	g.BlurShader.GetUniformLoc("uImage")

	fmt.Println(g.BloomStep1Shader.Uniforms)
	fmt.Println(g.BloomShader.Uniforms)

	// load data into displayable scene
	g.Scene.LoadData(g.Galaxy, float32(g.Galaxy.Radius))

	// ------------------------------------------------
	// for bloom:
	// https://github.com/JoeyDeVries/LearnOpenGL/blob/master/src/5.advanced_lighting/7.bloom/bloom.cpp
	// lines 97-
	// HDR Framebuffer

	GlClearError()
	gl.GenFramebuffers(1, &g.hdrFBO)
	gl.BindFramebuffer(gl.FRAMEBUFFER, g.hdrFBO)

	// Create 2 floating point color buffers (1 for normal rendering, other for brightness threshold values)
	g.colorBuffers = CreateTextureBuffers(2, int32(g.Win.Width), int32(g.Win.Height))

	// tell openGL which color Attachment wee'll use for rendering
	attachments := [2]uint32{gl.COLOR_ATTACHMENT0, gl.COLOR_ATTACHMENT1}
	gl.DrawBuffers(2, &attachments[0])

	// check if the framebuffer is complete:
	CheckFramebufferStatus()
	// unbind framebuffer
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	g.pingpongFBOs, g.pingpongColorBuffers = CreatePingPongFBOs(2, int32(g.Win.Width), int32(g.Win.Height))
	GlCheckError("Somethin went wrong")

	// feeding points and colors into the buffers
	// this should be in here or in a separate *Gui function in the future
	g.texBuf, g.colorBuf, g.fbo, g.vbo = FeedSceneToBuffers(g.Scene.Points, g.Scene.Colors, int32(g.Win.Width), int32(g.Win.Height))

	log.Printf("...done. (%d systems, %d ms)\n", g.Galaxy.SysCount, time.Since(start)/1000000)

	//		camera stuff
	// this is clunky. Maybe Cam and Persp should be combineed in an 'MVP' object or so,
	// and Cam and Persp init should be done in one.
	camDist := float32(InitialZoom / math.Tan(float64(glm.DegToRad(CamViewAngle))))
	// temp  - has to be changed to camDist
	g.Cam.SetPositionRadial(camDist, glm.DegToRad(CamAngleA), glm.DegToRad(CamAngleB))
	g.Cam.Target = glm.Vec3{0.0, 0.0, 0.0}

	// perspective stuff
	g.Persp = Perspective{float32(g.Win.Width) / float32(g.Win.Height), SceneNear, SceneFar, 0}
	//g.Persp = Perspective{float32(1), SceneNear, SceneFar, 0}
	g.Persp.SetViewAngleDeg(CamViewAngle)

	// set the viewport
	gl.Viewport(0, 0, int32(g.Win.Width), int32(g.Win.Height))
}

// ----------- MAINLOOP --------------
// Called once per Gameloop
func (g *Gui) Mainloop() {
	// imgui stuff
	g.Win.NewFrame()
	imgui.NewFrame()

	// create stats window and add text to it
	if g.showDisplayTitle {
		g.displayTitle()
	}

	imgui.SetNextWindowPos(imgui.Vec2{X: 0, Y: 0})
	imgui.SetNextWindowSize(imgui.Vec2{X: float32(g.Win.Width), Y: 100})
	imgui.BeginV("Stats", nil, imgui.WindowFlagsNoResize)
	imgui.Text(fmt.Sprintf("FPS: %d (%.3f ms/frame)", g.Fps, g.msPerFrame))
	imgui.PushStyleColor(imgui.StyleColorText, imgui.Vec4{0.3, 0.1, 0.9, 1.0})
	imgui.Text(fmt.Sprintf("%d Stars", g.Galaxy.SysCount))
	imgui.PopStyleColor()
	imgui.Text(fmt.Sprintf("Bloom active: %v\n", g.BloomActive))
	imgui.Text(fmt.Sprintf("Pan active: %v\n", g.Mouse.Pan))
	imgui.End()
	// --------------
	/*imgui.SetNextWindowBgAlpha(0.5)
	for i := 0; i < 20; i++ {
		imgui.Text("Random debug log entries go here...")
	}
	*/
	curTime := time.Now()
	// if one second passed, print frame draw time
	if time.Since(g.lastTime) > time.Second {

		// print frame rate and other info
		g.Fps = g.nbFrames
		g.msPerFrame = 1000 / float32(g.nbFrames)

		g.nbFrames = 0
		g.lastTime = curTime
	}
	// Clear Screen
	GlClearError()
	gl.ClearColor(0.0, 0.0, 0.0, 1.0)
	gl.Clear(gl.COLOR_BUFFER_BIT)
	GlCheckError("Clearing screen")
	if g.displayGalaxy {

		if !g.BloomActive {
			g.BloomShader.Use()
			gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

			gl.ClearColor(0.0, 0.0, 0.0, 1.0)
			gl.Clear(gl.COLOR_BUFFER_BIT)
			GlCheckError("Clearing screen")

			g.Shader.Use()
			// transformation matrix
			mvpMatrix := GetMVPMatrix(g.Cam, g.Persp)
			// Apply MVP (Model,View,Pre)
			UniformMatrix(g.Shader.Uniforms["uMVP"], mvpMatrix)

			// Clear Screen
			GlClearError()

			// draw the stars
			DrawDots(g.Galaxy.SysCount)

		} else {

			//g.BloomStep1Shader.Use()
			g.BloomShader.Use()
			g.BloomShader.SetFloat("exposure", g.uExposure)
			g.BloomShader.SetInt("bloomBlur", g.uBloomBlur)
			g.BloomShader.SetFloat("uSatMult", g.uSaturationMult)
			// ------------- BLOOM SHADER STUFF ------------
			// render to given framebuffer
			gl.BindFramebuffer(gl.FRAMEBUFFER, g.hdrFBO)
			attachments := [2]uint32{gl.COLOR_ATTACHMENT0, gl.COLOR_ATTACHMENT1}
			gl.DrawBuffers(2, &attachments[0])

			// clear the screen
			gl.ClearColor(0.0, 0.0, 0.0, 1.0)
			gl.Clear(gl.COLOR_BUFFER_BIT)
			g.BloomStep1Shader.Use()
			g.BloomStep1Shader.SetFloat("uBrightThreshold", g.uBrightThreshold)
			//g.BloomStep1Shader.SetFloat("uSatMult", g.uSaturationMult)
			// transformation matrix
			mvpMatrix := GetMVPMatrix(g.Cam, g.Persp)
			// Apply MVP (Model,View,Pre)
			UniformMatrix(g.BloomStep1Shader.Uniforms["uMVP"], mvpMatrix)
			DrawDots(g.Galaxy.SysCount)

			// clear the screen
			gl.ClearColor(0.0, 0.0, 0.0, 1.0)
			gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
			gl.Clear(gl.COLOR_BUFFER_BIT)

			// new shader
			g.BlurShader.Use()

			// set the current blur weights
			g.BlurShader.SetInt("uWeightLen", int32(len(g.uWeights[g.uActiveWeight])))
			g.BlurShader.SetFloatV("uWeight", g.uWeights[g.uActiveWeight])

			// ----------------- Flip flop blur
			var horizontal int32
			var buf uint32
			_ = buf
			horizontal = 1
			first_iteration := true

			// cleaning the pingpongFBOs
			gl.BindFramebuffer(gl.FRAMEBUFFER, g.pingpongFBOs[0])
			gl.Clear(gl.COLOR_BUFFER_BIT)
			gl.BindFramebuffer(gl.FRAMEBUFFER, g.pingpongFBOs[1])
			gl.Clear(gl.COLOR_BUFFER_BIT)

			// do the ping pong rendering - 5xhorizontal + 5xvertical
			for i := 0; i < g.blurSteps; i++ {
				gl.BindFramebuffer(gl.FRAMEBUFFER, g.pingpongFBOs[horizontal])
				g.BlurShader.SetInt("uHorizontal", horizontal)
				// bind texture to other framebuffer, or to scene if first run
				if first_iteration {
					buf = g.colorBuffers[1]
					first_iteration = false
				} else {
					buf = g.pingpongColorBuffers[1-horizontal]
				}
				gl.BindTexture(gl.TEXTURE_2D, buf)
				RenderQuad()
				horizontal = 1 - horizontal
			}

			// switch to screen and clear
			gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
			gl.ClearColor(0.0, 0.0, 0.0, 1.0)
			gl.Clear(gl.COLOR_BUFFER_BIT)

			g.BloomShader.Use()
			// first the 'normal' galaxy

			gl.ActiveTexture(gl.TEXTURE1)
			gl.BindTexture(gl.TEXTURE_2D, g.colorBuffers[0])
			// then the bloom on top
			// sequence doesn't matter - I checked both ways :p
			gl.ActiveTexture(gl.TEXTURE0)
			gl.BindTexture(gl.TEXTURE_2D, g.pingpongColorBuffers[1-horizontal])

			RenderQuad()

		}
	}
	// Display logging window
	g.displayLog()

	// Swap front and back buffers
	imgui.Render()
	g.ImGUIRenderer.Render(g.Win.DisplaySize(), g.Win.FramebufferSize(), imgui.RenderedDrawData())

	g.Win.Window.SwapBuffers()
	g.nbFrames += 1
	// Poll for and process events
	glfw.PollEvents()

	// limit Frame Rate
	if g.frameRateLimit > 0 {
		time.Sleep(time.Duration(g.tick - time.Since(curTime)))
	}

	// steady movement rate over time
	if g.autoRotate {
		g.rotPerFrameA = g.DegPerSecond * float32(time.Since(curTime).Seconds())
		g.Cam.SetPositionRadial(g.Cam.Dist, g.Cam.A+g.rotPerFrameA, g.Cam.B)
	}

	// abstracting exit request from window internals
	g.GameExitRequested = g.Win.Window.ShouldClose()

}
