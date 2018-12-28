"use strict";

const KEY_CODE_TO_BUTTON_CODE = {
    87: 1,
    81: 2,
    90: 3,
    88: 4,
    40: 5,
    38: 6,
    37: 7,
    39: 8
};

function main() {
    let emulatorWorker = new Worker("/static/emulator_worker.js");

    let romSelector = document.getElementById("rom-selector");
    romSelector.addEventListener("change", function(ev) {
        let files = ev.target.files;
        if (files.length == 0) {
            console.log("js: No files received!");
            return
        }

        let fileReader = new FileReader();
        fileReader.onload = function(ev) {
            let array = new Uint8Array(ev.target.result);

            emulatorWorker.postMessage(["CartridgeData", array]);
            console.log("js: Sent ROM data to emulator");
        }
        fileReader.readAsArrayBuffer(files[0]);
    })

    let display = document.getElementById("frame-display");
    let displayContext = display.getContext("2d");

    emulatorWorker.onmessage = function(ev) {
        switch (ev.data[0]) {
        case "NewFrame":
            let frame = new ImageData(ev.data[1], 160, 144);
            createImageBitmap(frame, 0, 0, 160, 144, {
                resizeWidth: 160*3,
                resizeHeight: 144*3,
                resizeQuality: "pixelated",
            }).then(function(response) {
                displayContext.drawImage(response, 0, 0);
            });
            break;
        }
    }

    document.onkeydown = function(ev) {
        if (ev.keyCode in KEY_CODE_TO_BUTTON_CODE) {
            emulatorWorker.postMessage([
                "ButtonPressed",
                KEY_CODE_TO_BUTTON_CODE[ev.keyCode]
            ]);
        }
    }
    
    document.onkeyup = function(ev) {
        if (ev.keyCode in KEY_CODE_TO_BUTTON_CODE) {
            emulatorWorker.postMessage([
                "ButtonReleased",
                KEY_CODE_TO_BUTTON_CODE[ev.keyCode]
            ]);
        }
    }
}

window.onload = main

