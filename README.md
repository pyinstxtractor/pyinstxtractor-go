# Pyinstxtractor Go

PyInstaller Extractor developed in Golang.

## Compiling for Deskop

```
go build
```

## Compiling for Web

GopherJS requires Go 1.18 or newer. For more details check https://github.com/gopherjs/gopherjs#installation-and-usage

```
go install github.com/gopherjs/gopherjs@v1.18.0-beta1

gopherjs build --minify --tags=gopherjs -o public/js/pyinstxtractor-go.js
```
