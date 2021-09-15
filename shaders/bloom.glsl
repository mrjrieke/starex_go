#shader VERTEX
#version 330 core
// -------------------------------------
// VERTEX SHADER
// -------------------------------------

layout(location = 0) in vec3 vPos;
layout(location = 1) in vec2 vTexCoords;
layout(location = 2) in vec2 vOverlayTex;

out vec2 TexCoords;
out vec2 OvTexCoords;

void main(){
    TexCoords = vTexCoords;
    OvTexCoords = vOverlayTex;
	gl_Position = vec4(vPos, 1.0);
}

#shader FRAGMENT
#version 330 core
// -------------------------------------
// FRAGMENT SHADER
// -------------------------------------
out vec4 FragColor;

in vec2 TexCoords;
in vec2 OvTexCoords;

uniform sampler2D scene;
uniform sampler2D bloomBlur;
uniform sampler2D overlayTex;
uniform bool bloom;
uniform float exposure;
uniform float uSatMult;
uniform bool uOverlayEnabled;

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

void main()
{            
    const float gamma = 2.2;
    vec3 hdrColor = texture(scene, TexCoords).rgb;      

    vec3 bloomColor = texture(bloomBlur, TexCoords).rgb;
//    if(bloom)
    hdrColor += bloomColor; // additive blending


    // tone mapping
    vec3 result = vec3(1.0) - exp(-hdrColor * exposure);
    // also gamma correct while we're at it       
    result = pow(result, vec3(1.0 / gamma));

   vec3 hsv = rgb2hsv(result);
    // boost Saturation is desired - no clue why the stars are so 'faint'
    hsv.y *= uSatMult;
    // back to rgb
    result = hsv2rgb(hsv);
    

    if (uOverlayEnabled) {
        vec4 ovrly = texture(overlayTex, TexCoords).rgba;
        FragColor =  vec4((result.rgb * ovrly.rgb),1.0); 
    } else {
        FragColor =  vec4(result.rgb ,1.0); 
    }
}