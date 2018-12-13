"use strict";

const VERTEX_SHADER_CODE = `
uniform vec4 position;
uniform mat4 transform;

void main() {
    gl_PointSize = 1.0;
    gl_Position = transform * position;
}
`

const FRAGMENT_SHADER_CODE = `
#ifdef GL_FRAGMENT_PRECISION_HIGH
precision highp float;
#else
precision mediump float;
#endif

void main() {
    gl_FragColor = vec4(1.0, 1.0, 1.0, 1.0);
}
`

function glErrors(gl) {
    let jsonErrors = "WebGL errors:";

    // TODO
}


function loadShaders(gl) {
    let program = gl.createProgram();
    let vertexShader = gl.shaderSource
}

function main() {
    const canvas = document.querySelector("#myCanvas");
    const gl = canvas.getContext("webgl");

    if (gl === null) {
        alert("WebGL is not available!");
        return;
    }

    gl.clearColor(0.0, 0.0, 0.0, 1.0);
    gl.clear(gl.COLOR_BUFFER_BIT);
}

window.onload = main

