#version 100

precision mediump float;

// Input vertex attributes (from vertex shader)
varying vec2 fragTexCoord;
varying vec4 fragColor;

// Input uniform values
uniform sampler2D texture0;
uniform vec4 colDiffuse;

// NOTE: Add here your custom variables

const vec2 size = vec2(800, 450); // render size
const float quality = 2.5; // lower = smaller glow, better quality

const float samples = 10.0;
uniform float brightness;

void main() {
  vec4 sum = vec4(0);
  vec2 sizeFactor = vec2(1) / size * quality;

  // Texel color fetching from texture sampler
  vec4 source = texture2D(texture0, fragTexCoord);

  const int range = int(0.5 * (samples - 1.0)); // should be = (samples - 1)/2;

  for (int x = -range; x <= range; x++) {
    for (int y = -range; y <= range; y++) {
      sum += texture2D(texture0, fragTexCoord + vec2(x, y) * sizeFactor);
    }
  }

  // Calculate final fragment color
  vec4 outColor = clamp(brightness * ((sum / (samples * samples)) + source) * colDiffuse, vec4(0), vec4(1));
  const vec4 paletteWhite = vec4(0.9058823529411765, 0.8352941176470589, 0.7019607843137254, 1);
  gl_FragColor = outColor * paletteWhite;
}
