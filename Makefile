all: gopherboy_server

gopherboy_server: static/emulator.wasm .FORCE
	go build -o gopherboy_server github.com/velovix/gopherboy/server

static/emulator.wasm: .FORCE
	GOOS=js GOARCH=wasm go build -o static/emulator.wasm github.com/velovix/gopherboy/cmd/gopherboy_wasm

run: gopherboy_server
	./gopherboy_server

clean:
	rm -f gopherboy static/emulator.wasm

.FORCE:

