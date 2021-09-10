#shader VERTEX
#version 330 core
// -------------------------------------
// VERTEX SHADER
// -------------------------------------

layout(location = 0) in vec3 vp;
layout(location = 1) in vec4 color;
// using the alpha channel for luminosity now. It seems like I am running out of space with vlum
//layout(location = 2) in float vlum;

uniform mat4 uMVP;
uniform float uDist;
uniform sampler2D uRenderedTexture;

//out vec4 vColor;
out float sDist;
out vec4 vColor;
out vec4 vFakeColor;

// All components are in the range [0…1], including hue.
vec3 rgb2hsv(vec3 c)
{
    vec4 K = vec4(0.0, -1.0 / 3.0, 2.0 / 3.0, -1.0);
    vec4 p = mix(vec4(c.bg, K.wz), vec4(c.gb, K.xy), step(c.b, c.g));
    vec4 q = mix(vec4(p.xyw, c.r), vec4(c.r, p.yzx), step(p.x, c.r));

    float d = q.x - min(q.w, q.y);
    float e = 1.0e-10;
    return vec3(abs(q.z + (q.w - q.y) / (6.0 * d + e)), d / (q.x + e), q.x);
}

// All components are in the range [0…1], including hue.
vec3 hsv2rgb(vec3 c)
{
    vec4 K = vec4(1.0, 2.0 / 3.0, 1.0 / 3.0, 3.0);
    vec3 p = abs(fract(c.xxx + K.xyz) * 6.0 - K.www);
    return c.z * mix(K.xxx, clamp(p - K.xxx, 0.0, 1.0), c.y);
}

void main() {
    vColor = color;
    gl_Position = uMVP * vec4(vp, 1.0);
    // calculate a distance fade.
    // the constants are determined purely by experiment.
    sDist = 1.2-(gl_Position[3]/4) ;
    // rgb to hsv 
    vec3 hsv = rgb2hsv(vec3(vColor));
    // V of hsv is luminosity of the star, and fading slightly over distance
    hsv.z = vColor[3] * sDist;
    // back to rgb
//    hsv.y *= 2.0;
    vec3 rgb = hsv2rgb(hsv);
    // alpha channel is fixed
    vColor[3] = 0.75;
    vColor=vec4(rgb.xyz, vColor[3]);
    // temp, until I know why nothing's displayed
//    vColor =vec4(1.0,1.0,1.0,1.0);
//    fColor = texture(uRenderedTexture,vColor);
    if (hsv.z > 0.75)
        vFakeColor = vec4(1.0, 0.0, 0.8, 1.0);
    else
        vFakeColor = vec4(0.0, 0.0, 0.0, 1.0);
     
}

#shader FRAGMENT
#version 330 core
// -------------------------------------
// FRAGMENT SHADER
// -------------------------------------
//layout(location = 0) out vec3 color

in vec4 sPosition;
in vec4 vColor;
in vec4 vFakeColor;
//in float sDist;

in vec2 TexCoords;

uniform float weight[5] = float[] (0.227027, 0.194596, 0.1216216, 0.054054, 0.016216);

out vec4 fColor;

void main() {
    fColor = vColor;
    //fColor = vec4(1.0,1.0,1.0,1.0);
    
}