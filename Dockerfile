from voidlinux/voidlinux:latest

run xbps-install -Suyy
run xbps-install -Syy jack-devel pkg-config gcc go
env GO111MODULE on

