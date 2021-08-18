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

uniform sampler2D uImage;
uniform int uHorizontal;

uniform float uWeight[10];
uniform int uWeightLen;

// --- best for small
//uniform float uWeight[3] = float[] (0.1216216, 0.054054, 0.016216);
// --- also nice
//uniform float uWeight[4] = float[] (0.1216216, 0.054054, 0.016216, 0.008);
// --- sparkling stars +.+
//uniform float uWeight[6] = float[] (0.1216216, 0.054054, 0.016216, 0.012, 0.008 , 0.005);
// ------
//uniform float uWeight[4] = float[] (0.194596, 0.1216216, 0.054054, 0.026216);
//uniform float uWeight[5] = float[] (0.227027, 0.194596, 0.1216216, 0.054054, 0.016216);
//uniform float uWeight[4] = float[] (0.327027, 0.1516216, 0.034054, 0.010216);
//uniform float uWeight[4] = float[] (0.327027, 0.1216216, 0.054054, 0.016216);
//uniform float uOffset[4] = float[] (0.0, 1.5, 3.0, 4.5);
//uniform float uWeight[7] = float[] (0.6, 0.4, 0.227027, 0.194596, 0.1216216, 0.054054, 0.016216);

void main()
{             
     int passes = uWeightLen;
     //int passes = uWeight.length();
     vec2 tex_offset = 1.0 / textureSize(uImage, 0); // gets size of single texel
     vec3 result = texture(uImage, TexCoords).rgb * uWeight[0];
     if(uHorizontal == 1)
     {
         for(int i = 1; i < passes; ++i)
         {
            result += texture(uImage, TexCoords + vec2(tex_offset.x * i, 0.0)).rgb * uWeight[i];
            result += texture(uImage, TexCoords - vec2(tex_offset.x * i, 0.0)).rgb * uWeight[i];
         }
     }
     else
     {
         for(int i = 1; i < passes; ++i)
         {
             result += texture(uImage, TexCoords + vec2(0.0, tex_offset.y * i)).rgb * uWeight[i];
             result += texture(uImage, TexCoords - vec2(0.0, tex_offset.y * i)).rgb * uWeight[i];
         }
     }
     FragColor = vec4(result, 1.0);
}

