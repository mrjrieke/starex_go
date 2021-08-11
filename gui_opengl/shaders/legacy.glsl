#shader VERTEX
#version 140

in vec3 vp;
in vec4 color;

uniform mat4 uMVP;
uniform mat4 uProjection;
uniform vec4 uTest;

out vec4 vColor;

void main() {
//    gl_Position = uView;
//    gl_Position = uProjection;
//    gl_Position = uTest;
    //gl_Position = vp;
    vColor = color;
    gl_Position = uMVP * vec4(vp, 1.0);
    //gl_Position = uView * uProjection * vec4(vp, 1.0);
    //gl_Position = vec4(vp, 1.0);
    //gl_Position = uTest;
//    gl_Position = uProjection *  vec4(vp, 1.0f);
//    gl_Position = uView *  vec4(vp, 1.0f);
//	TexCoord = vec2(aTexCoord.x, aTexCoord.y)
}

#shader FRAGMENT
#version 140

in vec4 vColor;
out vec4 fColor;

void main() {
    fColor = vColor;
//    fColor = vec4(1.0, 0.0, 0.5, 1.0);
}
