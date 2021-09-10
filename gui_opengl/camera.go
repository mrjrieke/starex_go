package gui_opengl

import (
	"fmt"
	"math"

	//	"example.com/helmut/starex_vis_opengl/opengl"

	"github.com/engoengine/glm"
)

const ()

type Perspective struct {
	AspectRatio  float32
	SceneNear    float32
	SceneFar     float32
	ViewAngleRad float32
}

func (p *Perspective) SetViewAngleDeg(focallen float32) {
	p.ViewAngleRad = glm.DegToRad(focallen)
}

func (p *Perspective) SetViewAngleRad(focallen float32) {
	p.ViewAngleRad = focallen
}

func (p *Perspective) GetProjectionMatrix() glm.Mat4 {
	return glm.Perspective(p.ViewAngleRad, p.AspectRatio, p.SceneNear, p.SceneFar)
}

type Camera struct {
	// orthogonal
	Pos glm.Vec3
	X   float32
	Y   float32
	Z   float32
	// radial
	Dist float32
	A    float32
	B    float32
	// projection
	Target glm.Vec3
}

func (cam *Camera) SetPosition(x float32, y float32, z float32) {
	fmt.Println("Camera.setPosition() not implemented yet!")
}

func (cam *Camera) SetPositionVec(pos glm.Vec3) {
	fmt.Println("Camera.setPositionVec() not implemented yet!")
}

func (cam *Camera) SetPositionRadial(dist float32, a float32, b float32) {
	cam.Dist = dist
	cam.A = a
	cam.B = b
	cam.Y = dist * float32(math.Sin(float64(b)))
	distXZ := float32(math.Sqrt(float64(cam.Dist*cam.Dist - cam.Y*cam.Y)))
	cam.X = distXZ * float32(math.Sin(float64(a)))
	cam.Z = distXZ * float32(math.Cos(float64(a)))
	cam.Pos = glm.Vec3{cam.X, cam.Y, cam.Z}
}

func (cam *Camera) GetViewMatrix() glm.Mat4 {
	camDirection := glm.NormalizeVec3(cam.Pos.Sub(&cam.Target))
	up := glm.Vec3{0.0, 1, 0}
	camRight := glm.NormalizeVec3(up.Cross(&camDirection))
	camUp := camDirection.Cross(&camRight)

	return glm.LookAtV(&cam.Pos, &cam.Target, &camUp)
}

func GetMVPMatrix(cam Camera, p Perspective) glm.Mat4 {
	// MVP = ModelViewProdection
	viewMatrix := cam.GetViewMatrix()
	projMatrix := p.GetProjectionMatrix()
	x := projMatrix.Mul4(&viewMatrix)
	return x

}
