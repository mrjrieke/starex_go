package gui_opengl

import (
	"bufio"
	"fmt"

	"os"
	"strings"

	"github.com/go-gl/gl/v4.4-core/gl"
)

const (
	// shader index
	ShaderTypeNone     = -1
	ShaderTypeVertex   = 0
	ShaderTypeFragment = 1
)

type Shader struct {
	ShaderName   string
	Program      uint32
	Uniforms     map[string]int32
	ShaderSource [2]string
	TexPtrs      [2]uint32
}

func (sd *Shader) GetUniformLoc(uname string) {
	unif := GetUniformLoc(sd.Program, uname)
	if sd.Uniforms == nil {
		sd.Uniforms = make(map[string]int32)
	}
	sd.Uniforms[uname] = unif
}

func (sd *Shader) CreateUniformLoc(uname string) int32 {
	var found bool
	var uid int32
	found = false
	for name, id := range sd.Uniforms {
		if name == uname {
			found = true
			uid = id
		}
	}
	if !found {
		sd.GetUniformLoc(uname)
		uid = sd.Uniforms[uname]
	}
	return uid
}

func (sd *Shader) SetFloatV(uname string, val []float32) {
	uid := sd.CreateUniformLoc(uname)
	vallen := int32(len(val))
	gl.Uniform1fv(uid, vallen, &val[0])
}

func (sd *Shader) SetInt(uname string, val int32) {
	uid := sd.CreateUniformLoc(uname)
	gl.Uniform1i(uid, val)
}

func (sd *Shader) SetFloat(uname string, val float32) {
	uid := sd.CreateUniformLoc(uname)
	gl.Uniform1f(uid, val)
	var f float32 
	gl.GetUniformfv(sd.Program, uid, &f)
//	fmt.Println ("Uniform for ", uname, ": ", f)
}

func (sd *Shader) Init(filename string) {
	sd.ShaderName = filename
	var shaderType = ShaderTypeNone
	// read shader file
	sf, err := os.Open(filename)
	if err != nil {
		panic("Failed to read shader file")
	}
	scanner := bufio.NewScanner(sf)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		//  lines starting with "#shader" determine the following content
		if strings.Contains(scanner.Text(), "#shader") {
			if strings.Contains(scanner.Text(), "FRAGMENT") {
				shaderType = ShaderTypeFragment
				continue
			} else if strings.Contains(scanner.Text(), "VERTEX") {
				shaderType = ShaderTypeVertex
				continue
			}
		}
		sd.ShaderSource[shaderType] += scanner.Text() + "\n"
	}
	sd.ShaderSource[ShaderTypeVertex] += "\x00"
	sd.ShaderSource[ShaderTypeFragment] += "\x00"
	sf.Close()

	sd.CreateShaderProg()
}

func (sd *Shader) CreateShaderProg() {
	sd.Program = gl.CreateProgram()

	GlClearError()

	vs := sd.compileShader(gl.VERTEX_SHADER, sd.ShaderSource[ShaderTypeVertex])
	fs := sd.compileShader(gl.FRAGMENT_SHADER, sd.ShaderSource[ShaderTypeFragment])
	gl.AttachShader(sd.Program, vs)
	gl.AttachShader(sd.Program, fs)

	gl.LinkProgram(sd.Program)
	gl.ValidateProgram(sd.Program)
	GlCheckError("createShader")
}

func (sd *Shader) Use() {
	GlClearError()
	gl.UseProgram(sd.Program)
	GlCheckError(fmt.Sprintf("UseProgram %s", sd.ShaderName))
}

func (sd *Shader) compileShader(shaderType uint32, source string) uint32 {
	id := gl.CreateShader(shaderType)
	csources, free := gl.Strs(source)
	_ = free

	GlClearError()
	gl.ShaderSource(id, 1, csources, nil)
	gl.CompileShader(id)

	var result int32
	gl.GetShaderiv(id, gl.COMPILE_STATUS, &result)
	GlCheckError("compileShader")
	if result == gl.FALSE {
		var length int32
		gl.GetShaderiv(id, gl.INFO_LOG_LENGTH, &length)
		log := strings.Repeat("\x00", int(length+1))
		gl.GetShaderInfoLog(id, length, nil, gl.Str(log))
		fmt.Println("Compile error! Message:", log)
		gl.DeleteShader(id)
		return 0
	}
	return id
}
