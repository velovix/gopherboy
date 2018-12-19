"use strict";

const VERTEX_SHADER_CODE = `
attribute vec3 coordinates;

// The size of the Game Boy screen
float SCREEN_WIDTH = 160.0;
float SCREEN_HEIGHT = 144.0;

void main() {
    vec3 glCoords;

    // Transform the coordinates from Game Boy space to OpenGL space
    glCoords.x = coordinates.x / SCREEN_WIDTH;
    glCoords.y = coordinates.y / SCREEN_HEIGHT;
    
    // Shift them to start at the bottom left of the screen
    glCoords.x = glCoords.x * 2.0 - 1.0;
    glCoords.y = glCoords.y * 2.0 - 1.0;

    gl_Position = vec4(glCoords, 1.0);
}
`

const FRAGMENT_SHADER_CODE = `
void main() {
    gl_FragColor = vec4(1.0, 1.0, 1.0, 1.0);
}
`

function glErrors(gl) {
    let jsonErrors = "WebGL errors: ";
    let errCount = 0;

    let err = gl.getError();
    while (err !== gl.NO_ERROR) {
        jsonErrors += err + " ";
        err = gl.getError();

        errCount++;
    }

    if (errCount > 0) {
        throw jsonErrors;
    }
}

function loadShader(gl, type, source) {
    let shader = gl.createShader(type);
    if (shader == 0) {
        throw "Failed to create shader!";
    }

    gl.shaderSource(shader, source);
    gl.compileShader(shader);

    var log = gl.getShaderInfoLog(shader);
    if (log.length > 0) {
        throw "Failed to compile shader: " + log;
    }

    glErrors(gl);

    return shader;
}


function loadShaders(gl) {
    let vertexShader = loadShader(gl, gl.VERTEX_SHADER, VERTEX_SHADER_CODE);
    let fragmentShader = loadShader(gl, gl.FRAGMENT_SHADER, FRAGMENT_SHADER_CODE);

    let program = gl.createProgram();

    gl.attachShader(program, vertexShader);
    gl.attachShader(program, fragmentShader);
    gl.linkProgram(program);

    if (!gl.getProgramParameter(program, gl.LINK_STATUS)) {
        let info = gl.getProgramInfoLog(program);
        throw "Could not compile program:" + info
    }

    glErrors(gl);

    return program;
}

function helloToJS(message) {
    alert("Got a message! " + message);
}

function main() {
    let dropZone = document.querySelector("#rom-drop-zone")
    dropZone.addEventListener("drop", function(ev) {
        ev.preventDefault();
        console.log("Gotem!", ev.dataTransfer.items);
        for (let i=0; i<ev.dataTransfer.items.length; i++) {
            console.log("gitem", ev.dataTransfer.items[i]);
            ev.dataTransfer.items[i].getAsString(function (s) {
                console.log("and check it", s)
            });
        }
    });
    dropZone.addEventListener("dragover", function(ev) {
        ev.preventDefault();
    });
    console.log("The trap has been set");

    const canvas = document.querySelector("#emulator-canvas");
    const gl = canvas.getContext("webgl");

    if (gl === null) {
        alert("WebGL is not available!");
        return;
    }

    let points = [
        5.0, 5.0, 0.0,
        155.0, 5.0, 0.0,
        5.0, 139.0, 0.0
    ]
    let pointsBuffer = gl.createBuffer();
    gl.bindBuffer(gl.ARRAY_BUFFER, pointsBuffer);
    gl.bufferData(gl.ARRAY_BUFFER, new Float32Array(points), gl.STATIC_DRAW);
    gl.bindBuffer(gl.ARRAY_BUFFER, null);

    let indices = [0, 1, 2];
    let indexBuffer = gl.createBuffer();
    gl.bindBuffer(gl.ELEMENT_ARRAY_BUFFER, indexBuffer);
    gl.bufferData(gl.ELEMENT_ARRAY_BUFFER, new Uint16Array(indices), gl.STATIC_DRAW);
    gl.bindBuffer(gl.ELEMENT_ARRAY_BUFFER, null);

    let program = loadShaders(gl);

    gl.useProgram(program);

    gl.bindBuffer(gl.ARRAY_BUFFER, pointsBuffer);
    gl.bindBuffer(gl.ELEMENT_ARRAY_BUFFER, indexBuffer);

    let coord = gl.getAttribLocation(program, "coordinates");
    gl.vertexAttribPointer(coord, 3, gl.FLOAT, false, 0, 0);
    
    gl.enableVertexAttribArray(coord);

    gl.clearColor(0.0, 0.0, 0.0, 1.0);

    gl.enable(gl.DEPTH_TEST);
    gl.clear(gl.COLOR_BUFFER_BIT);

    gl.viewport(0, 0, canvas.width, canvas.height);
    console.log(canvas.width, canvas.height);

    gl.drawElements(gl.TRIANGLES, indices.length, gl.UNSIGNED_SHORT, 0);
}

window.onload = main

