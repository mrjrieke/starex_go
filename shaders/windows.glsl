#shader VERTEX
#version 330 core

layout(location = 0) in vec3 vp;
layout(location = 1) in vec4 color;
layout(location = 2) in float lum;

uniform mat4 uMVP;

out vec4 vColor;
out vec4 BrightColor;
out float dist;

void main() {
    vColor = color;
    gl_Position = uMVP * vec4(vp, 1.0);
    dist = vp[2]
    /*
    if(float(lum) > 0.4)
        BrightColor = vec4(vec3(color),1.0);
    else
        BrightColor = vec4(0.0, 0.0, 0.0, 0.0);
        */
}

#shader FRAGMENT
#version 330 core

in vec4 gl_PointCoord;
in vec4 vColor;
in vec4 BrightColor;
in float dist;

uniform float weight[5] = float[] (0.227027, 0.194596, 0.1216216, 0.054054, 0.016216);

out vec4 fColor;

void main() {
    fColor = vColor;
    /*
    if (vColor[3] > 0.7)
    {
        fColor = vec4(vec3(vColor), 1.0);
    }
    */
}