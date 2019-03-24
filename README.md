# Gopherboy

Gopherboy is a Game Boy emulator written in Go. It plays a majority of tested
games properly and has support for sound.

## Screenshots

![Screenshots][screenshots]

## Performance

With the SDL back ends, it runs at roughly 6x speed on my machine (i7-8750H @
2.20 GHz). With the in-progress WebAssembly back end, it runs at roughly 0.5x
speed. There is still significant room for optimization.

[screenshots]: https://i.imgur.com/UlDcNVC.png
