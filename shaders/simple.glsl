#shader VERTEX
#version 330 core

layout(location = 0) in vec4 vp;
layout(location = 1) in vec4 color;

out vec4 vColor;

void main() {
    gl_Position = vp;
    vColor = color;

}

#shader FRAGMENT
#version 330 core

in vec4 vColor;
in vec4 BrightColor;
out vec4 fColor;

void main() {
    fColor = vColor;
}
