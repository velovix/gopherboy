// Import the Go wasm helper script
self.importScripts('/static/wasm_exec.js');

// Start running the emulator binary
const go = new Go();
WebAssembly.instantiateStreaming(
  fetch('/static/emulator.wasm'),
  go.importObject,
).then(result => {
  go.run(result.instance);
});
