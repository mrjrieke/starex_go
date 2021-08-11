#shader VERTEX
#version 330 core
// -------------------------------------
// VERTEX SHADER
// -------------------------------------

layout(location = 0) in vec3 vPos;
layout(location = 1) in vec2 vTexCoords;

out vec2 TexCoords;

void main(){
    TexCoords = vTexCoords;
	gl_Position = vec4(vPos, 1.0);
}

#shader FRAGMENT
#version 330 core
// -------------------------------------
// FRAGMENT SHADER
// -------------------------------------
out vec4 FragColor;

in vec2 TexCoords;

uniform sampler2D scene;
uniform sampler2D bloomBlur;
uniform bool bloom;
uniform float exposure;

void main()
{            
    const float gamma = 2.2;
    vec3 hdrColor = texture(scene, TexCoords).rgb;      
    vec3 bloomColor = texture(bloomBlur, TexCoords).rgb;
    if(bloom)
        hdrColor += bloomColor; // additive blending

    // tone mapping
    vec3 result = vec3(1.0) - exp(-hdrColor * exposure);
    // also gamma correct while we're at it       
    result = pow(result, vec3(1.0 / gamma));
    FragColor = vec4(result, 1.0);
}