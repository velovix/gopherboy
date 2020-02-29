# Gopherboy

Gopherboy is a Game Boy emulator written in Go. It plays a majority of tested
games properly and has support for sound.

## Screenshots

![Screenshots][screenshots]

## Performance

With the SDL back ends, it runs at roughly 1200% native speed on my machine
(i7-8750H @ 2.20 GHz). With the in-progress WebAssembly back end, it runs at
roughly 75% native speed.

[screenshots]: https://i.imgur.com/UlDcNVC.png

