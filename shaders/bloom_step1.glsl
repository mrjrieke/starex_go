#shader VERTEX
#version 330 core
// -------------------------------------
// VERTEX SHADER
// -------------------------------------

layout(location = 0) in vec3 iPos;
layout(location = 1) in vec4 iColor;
layout(location = 2) in vec2 iTexCoords;
// using the alpha channel for luminosity now. It seems like I am running out of space with vlum


uniform mat4 uMVP;
uniform float uDist;
uniform sampler2D uRenderedTexture;
uniform float uBrightThreshold;

out VS_OUT {
    vec4 vColor;
    vec4 vBrightColor;
    vec2 vTexCoords;
} vs_out;


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
    vs_out.vColor = iColor;
    vs_out.vTexCoords = iTexCoords;

    float orig_bright = iColor[3];
    gl_Position = uMVP * vec4(iPos, 1.0);
    // calculate a distance fade.
    // the constants are determined purely by experiment.
    float sDist = 1.2-(gl_Position[3]/4) ;
    // rgb to hsv 
    vec3 hsv = rgb2hsv(vec3(vs_out.vColor));
    // V of hsv is luminosity of the star, and fading slightly over distance
    hsv.z = vs_out.vColor[3] * sDist;
    // back to rgb
    vec3 rgb = hsv2rgb(hsv);
    // alpha channel is fixed
    vs_out.vColor=vec4(rgb.xyz, vs_out.vColor[3]);
    //vs_out.vColor=vec4(rgb.xyz, orig_bright);

   // temp, until I know why nothing's displayed
    if (hsv.z > uBrightThreshold)
        vs_out.vBrightColor = vec4(rgb.xyz, orig_bright);
    //    vs_out.vBrightColor = vec4(rgb.xyz, 1.0);
    else
        //vs_out.vBrightColor = vec4(0.0, 0.0, 0.0, orig_bright);
        vs_out.vBrightColor = vec4(0.0, 0.0, 0.0, 1.0);
     
}

#shader FRAGMENT
#version 330 core
// -------------------------------------
// FRAGMENT SHADER
// -------------------------------------


//in vec4 sPosition;
in VS_OUT {
    vec4 vColor;
    vec4 vBrightColor;
    vec2 vTexCoords;
} vs_in;

uniform sampler2D diffuseTexture;

uniform float weight[5] = float[] (0.227027, 0.194596, 0.1216216, 0.054054, 0.016216);

//out vec4 FragColor;
layout (location = 0) out vec4 FragColor;
layout (location = 1) out vec4 BrightColor;

void main() {
    FragColor = vs_in.vColor;
    BrightColor = vs_in.vBrightColor;
}