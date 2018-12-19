all: gopherboy

gopherboy: static/emulator.wasm *.go
	go build

static/emulator.wasm: emulator/*.go
	GOOS=js GOARCH=wasm go build -o static/emulator.wasm github.com/velovix/gopherboy/emulator

run: gopherboy
	./gopherboy

clean:
	rm -f gopherboy static/emulator.wasm
