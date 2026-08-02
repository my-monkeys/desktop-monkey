package main

// version est injectee au build : -ldflags "-X main.version=2.5.0".
// "dev" designe un build local, que l'auto-mise a jour laisse tranquille.
var version = "dev"
