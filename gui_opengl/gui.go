package gui_opengl

import (
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"image"
	"image/png"

	"github.com/go-gl/gl/v4.4-core/gl"
	"github.com/go-gl/glfw/v3.0/glfw"
	"github.com/engoengine/glm"

//	"github.com/shibukawa/nanogui-go"

	"github.com/Jest0r/starex_go/galaxy"

)

const (
	ScreenWidth      = 1200
	ScreenHeight     = 800
	FullRotationTime = 15.0
	CamX             = -20000
	CamY             = -20000
	CamZ             = 10000
	CamViewAngle     = 35.0 // NOT focal length
	CamAngleA        = 0
	CamAngleB        = 30
	MinCamDist       = 0.05

	InitialZoom     = 1.4
	ZoomIncrement   = 0.01
	RotateIncrement = 0.01

	SceneNear = 0.1
	SceneFar  = 100

	// turn off if 0
	FrameRateLimit = 0

	Bloom = true
)

type Window struct {
	Window     *glfw.Window
	Title      string
	Width      int
	Height     int
	Fullscreen bool
}

func (w *Window) Init() {
	// initialize the library
	if err := glfw.Init(); !err {
		panic(err)
	}
//	nanogui.Init()
	// create a window mode and it's OpenGL Context
	w.Window = InitScreen(w.Width, w.Height, w.Title, w.Fullscreen)

	// Init OpenGL
	if err := gl.Init(); err != nil {
		panic("Init error! - " + err.Error())
	}
	//	gl.GetString(name uint32)
	version := gl.GoStr(gl.GetString(gl.VERSION))
	log.Println("OpenGL version", version)

	// Enable Texture
	gl.Enable(gl.TEXTURE_2D)

	// Enable Blending
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE)
}

type Gui struct {
	Win   Window
	Cam   Camera
	Scene Scene
	Persp Perspective
	//NanoGUI nanogui.Screen
	// --- Shaders
	Shader           ShaderData // Shader for simple display
	BloomStep1Shader ShaderData // Step 1 for enabled lighting effect
	BlurShader       ShaderData // Step 2 for enabled lighting effect, called multiple times
	BloomShader      ShaderData // Step 3 and final step for enabled lighting effects
	// --- Shader options
	uBrightThreshold float32
	uExposure        float32
	uBloomBlur       int32
	uWeights         [][]float32
	uActiveWeight    int32
	blurSteps        int
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
	hdrFBO               uint32
	//TexPtrs        [2]uint32
	BloomActive bool
}

func (g *Gui) setCallbacks() {
	g.Win.Window.SetSizeCallback(g.windowSizeCallback)
	g.Win.Window.SetKeyCallback(g.keyCallback)
	g.Win.Window.SetMouseButtonCallback(g.mouseCallback)
	g.Win.Window.SetPositionCallback(g.posCallback)
	g.Win.Window.SetScrollCallback(g.scrollCallback)
}

func (g *Gui) cleanup() {
	fmt.Println("Cleaning up...")
	defer glfw.Terminate()
}

func (g *Gui) Pause() {
	fmt.Println("----- PAUSE -----")
	for g.pause {
		time.Sleep(100 * time.Millisecond)
		glfw.PollEvents()
	}

}

// ----------- Callbacks --------------
func (g *Gui) windowSizeCallback(window *glfw.Window, width int, height int) {
	g.Win.Width = width
	g.Win.Height = height
	fmt.Println("Adjusting window")

	gl.Viewport(0, 0, int32(width), int32(height))
	g.Persp.AspectRatio = float32(g.Win.Width / g.Win.Height)
}

func (g *Gui) mouseCallback(window *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {
	if action == glfw.Press {
		switch button {
		case glfw.MouseButton1:
			g.Mouse.pressButton(1)
		case glfw.MouseButton2:
			g.Mouse.pressButton(2)
		}
	} else if action == glfw.Release {
		switch button {
		case glfw.MouseButton1:
			g.Mouse.releaseButton(1)
		case glfw.MouseButton2:
			g.Mouse.releaseButton(2)
		}
	}
}

func (g *Gui) posCallback(window *glfw.Window, xpos int, ypos int) {
	g.Mouse.move(xpos, ypos)
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
		switch key {
		case glfw.KeyF11:
			g.toggleFullscreen()
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
		// --- brightness threshold
		case glfw.KeyR:
			g.uBrightThreshold += 0.1
		case glfw.KeyF:
			if g.uBrightThreshold > 0.1 {
				g.uBrightThreshold -= 0.1
			}
			fmt.Println(g.uBrightThreshold)
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
		case glfw.KeyV:
			g.uBloomBlur = 1 - g.uBloomBlur
		case glfw.KeyC:
			g.uActiveWeight = (g.uActiveWeight + 1) % int32(len(g.uWeights))
			fmt.Println("Active Weight", g.uActiveWeight)
		case glfw.KeyX:
			g.blurSteps = (g.blurSteps + 2) % 10
			fmt.Println("Blur Steps:", g.blurSteps)
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
	fmt.Println("Image save complete.")
}

// ----------- Toggles --------------

func (g *Gui) toggleFullscreen() {
	// Toggle fullscreen
	g.Win.Fullscreen = !g.Win.Fullscreen
	g.Win.Width, g.Win.Height = g.Win.Window.GetSize()
	// Close the current window.
	g.Win.Window.Destroy()
	g.Win.Init()
	// Enable Texture
	gl.Enable(gl.TEXTURE_2D)
	// Enable Blending
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE)
	// Shaders
	g.Shader.CreateShaderProg()
	g.BloomStep1Shader.CreateShaderProg()
	g.BlurShader.CreateShaderProg()
	g.BloomShader.CreateShaderProg()
	//	g.Shader.FeedBuffers()
	FeedColorBuffer(g.Scene.Colors)
	FeedLumBuffer(g.Scene.Lums)
	g.Shader.GetUniformLoc("uMVP")
	g.setCallbacks()
}

func (g *Gui) togglePause() {
	g.pause = !g.pause
	if g.pause {
		g.Pause()
	}
}
/*
func (g *Gui) LoadGalaxyFromFile(filename string) {
	// loading galaxy
	fmt.Print("Loading Data...")
	// reading json file into internal structure
	g.Galaxy.Import(filename)
//	g.Galaxy.Import("saves/galaxy2")
	fmt.Printf("done.\nPreparing data...")
	// loading into internal format
	// feeding graphics card with internal format
	g.PrepareScene()

}
*/

func (g *Gui) Init() {
	// Init some vars
	// 		window stuff
	g.Win.Title = "Starex Starfield Visualizer (openGL)"
	g.Win.Height = int(ScreenHeight)
	g.Win.Width = int(ScreenWidth)
	g.Win.Init()
	// Set Callbacks for Key input and Size change
	g.setCallbacks()
	g.BloomActive = Bloom

//	g.NanoGUI.Initialize(@g.Win.Window, true)
//	nanogui.Init()
//	nanogui.

	// get glgs version
	glgsver := strings.Split(gl.GoStr(gl.GetString(gl.SHADING_LANGUAGE_VERSION)), ".")

	// Loading Shaders
	if gv, _ := strconv.Atoi(glgsver[0]); gv < 3 {
		fmt.Println("Old glgs version - using legacy shader")
		g.Shader.Init("shaders/legacy.glsl")
	} else {
		//g.Shader.UseShader("shaders/bloom.glsl")
		g.Shader.Init("shaders/experimental.glsl")
	}
	// get Uniform loc
	g.Shader.GetUniformLoc("uMVP")

	// specific shader just for bloom effect
	g.BloomStep1Shader.Init("shaders/bloom_step1.glsl")
	g.BloomStep1Shader.GetUniformLoc("uMVP")

	g.BlurShader.Init("shaders/blur.glsl")
	g.BlurShader.GetUniformLoc("uImage")
	// init different blurring weights.
	g.uWeights = append(g.uWeights, []float32{0.1216216, 0.054054, 0.016216})
	g.uWeights = append(g.uWeights, []float32{0.1216216, 0.054054, 0.016216, 0.008})
	g.uWeights = append(g.uWeights, []float32{0.1216216, 0.054054, 0.016216, 0.012, 0.008, 0.005})
	g.uActiveWeight = 0

	g.BloomShader.Init("shaders/bloom.glsl")
	g.BloomShader.GetUniformLoc("uImage")
	g.BloomShader.Use()
	g.BloomShader.SetInt("bloom", 1)
	g.BloomShader.SetInt("scene", 0)

	g.uBrightThreshold = 1.0
	g.uExposure = 1.0
	g.uBloomBlur = 1

	g.blurSteps = 6
}


func (g *Gui) PrepareScene(){
	start := time.Now()
	g.Scene.LoadData(g.Galaxy, float32(g.Galaxy.Radius))
	// ------------------------------------------------
	// for bloom:
	// https://github.com/JoeyDeVries/LearnOpenGL/blob/master/src/5.advanced_lighting/7.bloom/bloom.cpp
	// lines 97-
	// HDR Framebuffer
	//	var hdrFBO uint32
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

	// ------------------------------------------------
	// Here the buffer magic starts

	//---- not sure if that correct or if that should be done via the ColorBuffers
	GlClearError()
	gl.GenBuffers(1, &g.colorBuf)
	gl.BindBuffer(gl.ARRAY_BUFFER, g.colorBuf)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(g.Scene.Colors), gl.Ptr(g.Scene.Colors), gl.STATIC_DRAW)
	// describe what the positions array actually mean
	gl.VertexAttribPointer(1, 4, gl.FLOAT, false, 0, nil)
	gl.EnableVertexAttribArray(1)
	GlCheckError("Feed Colors into VBO")

	g.texBuf, g.fbo = FeedVBOBuffer3D(g.Scene.Points, int32(g.Win.Width), int32(g.Win.Height))

	// ------------------------------------------------

	fmt.Printf("...done. (%d systems, %d ms)\n", g.Galaxy.SysCount, time.Since(start)/1000000)

	//		camera stuff
	// this is clunky. Maybe Cam and Persp should be combineed in an 'MVP' object or so,
	// and Cam and Persp init should be done in one.
	camDist := float32(InitialZoom / math.Tan(float64(glm.DegToRad(CamViewAngle))))
	fmt.Println(camDist)
	// temp  - has to be changed to camDist
	g.Cam.SetPositionRadial(camDist, glm.DegToRad(CamAngleA), glm.DegToRad(CamAngleB))
	g.Cam.Target = glm.Vec3{0.0, 0.0, 0.0}

	// perspective stuff
	g.Persp = Perspective{float32(g.Win.Width / g.Win.Height), SceneNear, SceneFar, 0}
	g.Persp.SetViewAngleDeg(CamViewAngle)

	// set the viewport
	gl.Viewport(0, 0, int32(g.Win.Width), int32(g.Win.Height))
}

// ----------- MAINLOOP --------------

func (g *Gui) Mainloop() {
	//		angle stuff
	var rotPerFrameA float32 = 0
	// timing stuff
	lastTime := time.Now()
	var nbFrames int = 0
	g.DegPerSecond = float32(-2 * math.Pi / FullRotationTime)
	// 		framerate stuff
	var tick time.Duration
	var frameRateLimit int = FrameRateLimit
	if frameRateLimit > 0 {
		tick = time.Duration(1000000/frameRateLimit) * time.Microsecond
	}
	for !g.Win.Window.ShouldClose() {

		curTime := time.Now()
		// if one second passed, print frame draw time
		if time.Since(lastTime) > time.Second {
			// print frame rate and other info
			fmt.Printf("%d Stars. %.3f ms/frame (desired %d) - %d fps. - bloom active: %v\n", g.Galaxy.SysCount, 1000/float32(nbFrames), tick, nbFrames, g.BloomActive)
			nbFrames = 0
			lastTime = curTime
		}

		if !g.BloomActive {
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
			g.BloomStep1Shader.SetFloat("uBrightThreshold", g.uBrightThreshold)
			g.BloomShader.SetFloat("exposure", g.uExposure)
			g.BloomShader.SetInt("bloomBlur", g.uBloomBlur)
			// ------------- BLOOM SHADER STUFF ------------
			// render to given framebuffer
			gl.BindFramebuffer(gl.FRAMEBUFFER, g.hdrFBO)
			attachments := [2]uint32{gl.COLOR_ATTACHMENT0, gl.COLOR_ATTACHMENT1}
			gl.DrawBuffers(2, &attachments[0])

			// Clear Screen
			GlClearError()
			gl.ClearColor(0.0, 0.0, 0.0, 1.0)
			gl.Clear(gl.COLOR_BUFFER_BIT)
			GlCheckError("Clearing screen")

			g.BloomStep1Shader.Use()
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
			gl.ActiveTexture(gl.TEXTURE0)
			gl.BindTexture(gl.TEXTURE_2D, g.pingpongColorBuffers[1-horizontal])

			RenderQuad()

		}
		// Swap front and back buffers
		g.Win.Window.SwapBuffers()
		nbFrames += 1
		// Poll for and process events
		glfw.PollEvents()

		// limit Frame Rate
		if frameRateLimit > 0 {
			time.Sleep(time.Duration(tick - time.Since(curTime)))
		}

		// steady movement rate over time
		rotPerFrameA = g.DegPerSecond * float32(time.Since(curTime).Seconds())
		g.Cam.SetPositionRadial(g.Cam.Dist, g.Cam.A+rotPerFrameA, g.Cam.B)
	}

}

/*
// ------------- MAIN -------------
func main() {
	// strictly neccessary, otherwise everything breaks. Not sure why exactly. I believe gl.init has to be in the same thread...
	runtime.LockOSThread()
	var game Game
	//	Define Cleanup procedure
	defer game.cleanup()

	game.init()

	game.mainloop()
}
*/
