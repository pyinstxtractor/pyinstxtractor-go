# Pyinstxtractor-Go🌐

PyInstaller Extractor developed in Golang.

Runs both on desktop and on the web. The browser version is powered by GopherJS to compile Go to Javascript.

Site hosted on Netlify.

Try it out at https://pyinstxtractor-web.netlify.app/

[![Netlify Status](https://api.netlify.com/api/v1/badges/63aa28b4-8134-44d9-a934-7e2833b79557/deploy-status)](https://app.netlify.com/sites/pyinstxtractor-web/deploys)

## Known Limitations

The tool (both desktop & web) works best with Python 3.x based PyInstaller executables. Python 2.x based executables are still supported but the PYZ archive wqonrt be extracted.

## See also

- [pyinstxtractor](https://github.com/extremecoders-re/pyinstxtractor): The original tool developed in Python.
- [pyinstxtractor-ng](https://github.com/pyinstxtractor/pyinstxtractor-ng): Same as pyinsxtractor but this one doesn't require Python to run and can extract all supported pyinstaller versions.


## Compiling for Deskop

```
go build
```

## Compiling for Web

GopherJS requires Go 1.18.x. For more details check https://github.com/gopherjs/gopherjs#installation-and-usage

```
go install github.com/gopherjs/gopherjs@v1.18.0-beta1

gopherjs build --minify --tags=gopherjs -o public/js/pyinstxtractor-go.js
```
