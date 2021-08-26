package gui_opengl

//https://kylewbanks.com/blog/tutorial-opengl-with-golang-part-2-hello-opengl
//https://pkg.go.dev/github.com/go-gl/gl@v-1.-1.-1-20210315015931-ae71cafe8d/v3.5-core/gl
// docs.gl
// https://www.youtube.com/watch?v=FBbPWSOQ0-w

import (
	"fmt"
	"strconv"

	//	"runtime"

	// > 4.4 panics on windows
	"github.com/engoengine/glm"
	"github.com/go-gl/gl/v4.4-core/gl"
	//"github.com/go-gl/gl/v3.2-core/gl"
)

// ----- ERROR CHECKING ------
func GlClearError() {
	for gl.GetError() != gl.NO_ERROR {
	}
}

func GlCheckError(module string) bool {
	es := map[string]string{
		"500": "GL_INVALID_ENUM",
		"501": "GL_INVALID_VALUE",
		"502": "GL_INVALID_OPERATION",
		"504": "GL_STACK_UNDERFLOW",
		"505": "GL_OUT_OF_MEMORY",
		"506": "GL_INVALID_FRAMEBUFFER_OPERATION",
		"507": "GL_CONTEXT_LOST"}

	for error := gl.GetError(); error != gl.NO_ERROR; {
		hexerror := strconv.FormatInt(int64(error), 16)
		fmt.Printf("ERROR %s: OpenGL Error (0x%s): %s\n", module, hexerror, es[hexerror])
		return false
	}
	return true
}

// ---- INITS -----------------

func createTextures(amount int32, fbo uint32) [2]uint32 {
	var tex [2]uint32
	gl.GenTextures(amount, &tex[0])
	for i := int32(0); i < amount; i += 1 {
		GlClearError()
		gl.BindTexture(gl.TEXTURE_2D, tex[i])
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGB, 1200, 800, 0, gl.RGB, gl.UNSIGNED_BYTE, nil)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
		//		gl.FramebufferTexture(gl.FRAMEBUFFER, uint32(gl.COLOR_ATTACHMENT0)+uint32(i), tex[i], 0)
		//gl.FramebufferTexture2D(gl.FRAMEBUFFER, uint32(gl.COLOR_ATTACHMENT0)+uint32(i), gl.TEXTURE_2D, tex[i], 0)
		gl.FramebufferTexture2D(fbo, uint32(gl.COLOR_ATTACHMENT0)+uint32(i), gl.TEXTURE_2D, tex[i], 0)
		GlCheckError("Generate Textures")
	}
	return tex
}

var vao uint32

func FeedVBOBuffer3D(positions []float32, colors []float32, width int32, height int32) (uint32, uint32, uint32, uint32) {
	var tex uint32
	var vbo uint32
	var col uint32
	var fbo uint32

	gl.GenTextures(1, &tex)
	GlClearError()
	gl.BindTexture(gl.TEXTURE_2D, tex)
	GlCheckError("Bind Texture")

	GlClearError()
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGB, width, height, 0, gl.RGB, gl.UNSIGNED_BYTE, nil)
	GlCheckError("Tex Image 2D")

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)

	GlClearError()
	gl.GenFramebuffers(1, &fbo)
	GlCheckError("Creating Frame Buffers")

	fbs := gl.CheckFramebufferStatus(gl.FRAMEBUFFER)
	if fbs != gl.FRAMEBUFFER_COMPLETE {
		o := ""
		switch fbs {
		case gl.FRAMEBUFFER_INCOMPLETE_ATTACHMENT:
			o = "FRAMEBUFFER INCOMPLETE ATTACHMENT"
		case gl.FRAMEBUFFER_INCOMPLETE_MISSING_ATTACHMENT:
			o = "No Images are attached to the Frame Buffer"
		case gl.FRAMEBUFFER_UNSUPPORTED:
			o = "gl.FRAMEBUFFER_UNSUPPORTED"
		default:
			o = strconv.Itoa(int(fbs))
		}
		fmt.Println("Framebuffer status:", o)
	}

	GlClearError()
	gl.BindFramebuffer(gl.FRAMEBUFFER, fbo)
	GlCheckError("Binding Frame buffers")

	GlClearError()
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, tex, 0)
	GlCheckError("Framebuffer Texture 2D")

	gl.GenVertexArrays(2, &vao)
	gl.BindVertexArray(vao)
	GlCheckError("VBO - Bind Vertex Array")

	GlClearError()
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	GlCheckError("VBO - Bind Buffer")
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(positions), gl.Ptr(positions), gl.STATIC_DRAW)
	GlCheckError("VBO - buffer data")
	// describe what the positions array actually mean
	GlClearError()
	gl.EnableVertexAttribArray(0)
	GlCheckError("VBO - Enable Vertex Array")
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, nil)
	GlCheckError("VBO - Enable Buffer")

	// color buffer
	GlClearError()
	gl.GenBuffers(1, &col)
	gl.BindBuffer(gl.ARRAY_BUFFER, col)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(colors), gl.Ptr(colors), gl.STATIC_DRAW)
	GlCheckError("Bind Color Buffer")

	GlClearError()
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 4, gl.FLOAT, false, 0, nil)
	GlCheckError("feedVBOBuffer3D")

	return tex, col, fbo, vbo

}

func FeedColorBuffer(colors []float32) uint32 {
	var vbo uint32
	GlClearError()
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(colors), gl.Ptr(colors), gl.STATIC_DRAW)
	// describe what the positions array actually mean
	gl.VertexAttribPointer(1, 4, gl.FLOAT, false, 0, nil)
	gl.EnableVertexAttribArray(1)
	GlCheckError("feedColorBuffer")
	return vbo
}

// https://github.com/JoeyDeVries/LearnOpenGL/blob/master/src/5.advanced_lighting/7.bloom/bloom.cpp
// lines 101-113
func CreateTextureBuffers(numBufs uint8, w int32, h int32) []uint32 {
	var i uint8
	texBufs := make([]uint32, numBufs)

	gl.GenTextures(int32(numBufs), &texBufs[0])
	for i = 0; i < numBufs; i++ {
		GlClearError()
		gl.BindTexture(gl.TEXTURE_2D, texBufs[i])
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA16F, w, h, 0, gl.RGBA, gl.FLOAT, nil)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
		GlCheckError("Create Texture Buffers")
		GlClearError()
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0+uint32(i), gl.TEXTURE_2D, texBufs[i], 0)
		GlCheckError("Attach texture to framebuffer")
	}
	return texBufs
}

// https://github.com/JoeyDeVries/LearnOpenGL/blob/master/src/5.advanced_lighting/7.bloom/bloom.cpp
// lines 129 - 146
func CreatePingPongFBOs(numBufs uint8, w int32, h int32) ([]uint32, []uint32) {
	var i uint8

	pingpongFBOs := make([]uint32, numBufs)
	pingpongColorBuffers := make([]uint32, numBufs)

	gl.GenFramebuffers(int32(numBufs), &pingpongFBOs[i])
	gl.GenTextures(int32(numBufs), &pingpongColorBuffers[0])

	for i = 0; i < numBufs; i++ {
		GlClearError()
		gl.BindFramebuffer(gl.FRAMEBUFFER, pingpongFBOs[i])
		gl.BindTexture(gl.TEXTURE_2D, pingpongColorBuffers[i])
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA16F, w, h, 0, gl.RGBA, gl.FLOAT, nil)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, pingpongColorBuffers[i], 0)
		GlCheckError("Create Ping Pong Buffers")

		CheckFramebufferStatus()
		GlClearError()
		GlCheckError("Attach texture to framebuffer")
	}

	return pingpongFBOs, pingpongColorBuffers
}

// https://github.com/JoeyDeVries/LearnOpenGL/blob/master/src/5.advanced_lighting/7.bloom/bloom.cpp
// lines 390 - 418
var quadVAO uint32
var quadVBO uint32

func RenderQuad() {
	if quadVAO == 0 {
		//		fmt.Println("Init Quad")

		quadVertices := []float32{-1, 1, 0, 0, 1,
			-1, -1, 0, 0, 0,
			1, 1, 0, 1, 1,
			1, -1, 0, 1, 0}

		gl.GenVertexArrays(1, &quadVAO)
		gl.BindVertexArray(quadVAO)

		gl.GenBuffers(1, &quadVBO)
		gl.BindBuffer(gl.ARRAY_BUFFER, quadVBO)
		gl.BufferData(gl.ARRAY_BUFFER, 4*len(quadVertices), gl.Ptr(quadVertices), gl.STATIC_DRAW)
		//		fmt.Println("quad size", 4*len(quadVertices))

		gl.EnableVertexAttribArray(0)
		gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 5*4, nil)
		gl.EnableVertexAttribArray(1)
		gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 5*4, gl.PtrOffset(12))
	}
	gl.BindVertexArray(quadVAO)
	gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)
	gl.BindVertexArray(0)
}

func CheckFramebufferStatus() bool {
	// check if the framebuffer is complete:
	fbs := gl.CheckFramebufferStatus(gl.FRAMEBUFFER)
	if fbs != gl.FRAMEBUFFER_COMPLETE {
		o := ""
		switch fbs {
		case gl.FRAMEBUFFER_INCOMPLETE_ATTACHMENT:
			o = "FRAMEBUFFER INCOMPLETE ATTACHMENT"
		case gl.FRAMEBUFFER_INCOMPLETE_MISSING_ATTACHMENT:
			o = "No Images are attached to the Frame Buffer"
		case gl.FRAMEBUFFER_UNSUPPORTED:
			o = "gl.FRAMEBUFFER_UNSUPPORTED"
		}
		fmt.Println("Framebuffer status:", o)
	}
	return fbs == gl.FRAMEBUFFER_COMPLETE
}

func FeedLumBuffer(lums []float64) uint32 {
	var vbo uint32
	GlClearError()
	gl.GenBuffers(1, &vbo)
	//	fmt.Println("---")
	//	fmt.Println(lums)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(lums), gl.Ptr(lums), gl.STATIC_DRAW)
	// describe what the positions array actually mean
	gl.VertexAttribPointer(2, 1, gl.FLOAT, false, 0, nil)
	gl.EnableVertexAttribArray(2)
	GlCheckError("feedLumBuffer")
	return vbo
}

func FeedIBOBuffer(indices []uint32) uint32 {
	var ibo uint32
	gl.GenBuffers(1, &ibo)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ibo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, 4*len(indices), gl.Ptr(indices), gl.STATIC_DRAW)
	// describe what the positions array actually mean
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 8, nil)
	gl.EnableVertexAttribArray(0)
	return ibo
}

func SelectIBO(ibo uint32) {
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ibo)
}

func GetUniformLoc(program uint32, varname string) int32 {
	//	fmt.Println("looking for uniform name ", varname)
	loc := gl.GetUniformLocation(program, gl.Str(varname+"\x00"))
	if loc == -1 {
		fmt.Println("Uniform not found!", varname)
	}
	return loc
}

func UniformMatrix(varname int32, data glm.Mat4) {
	GlClearError()
	gl.UniformMatrix4fv(varname, 1, false, &data[0])
	GlCheckError("UniformMatrix")
}

func UniformVector(varname int32, data glm.Vec4) {
	GlClearError()
	gl.Uniform4fv(varname, 1, &data[0])
	GlCheckError("UniformVector")
}

func UniformFloat(varname int32, data float32) {
	GlClearError()
	gl.Uniform1f(varname, data)
	GlCheckError("UniformFloat")
}

func DrawDots(data_len int32) {
	GlClearError()
	gl.BindVertexArray(vao)
	gl.DrawArrays(gl.POINTS, 0, int32(data_len))
	GlCheckError("DrawDots")
}
