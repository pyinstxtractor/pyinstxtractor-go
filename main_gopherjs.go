//go:build gopherjs
// +build gopherjs

package main

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"pyinstxtractor-go/marshal"

	"github.com/go-restruct/restruct"
	"github.com/gopherjs/gopherjs/js"
)

type PyInstArchive struct {
	inFilePath              string
	outZip                  *zip.Writer
	fPtr                    io.ReadSeeker
	fileSize                int64
	cookiePosition          int64
	pyInstVersion           int64
	pythonMajorVersion      int
	pythonMinorVersion      int
	overlaySize             int64
	overlayPosition         int64
	tableOfContentsSize     int64
	tableOfContentsPosition int64
	tableOfContents         []CTOCEntry
	pycMagic                [4]byte
	gotPycMagic             bool
	barePycsList            []*barePyc
}

type barePyc struct {
	filepath string
	contents []byte
}

var logFunc *js.Object

func appendLog(logLine string) {
	logFunc.Invoke(logLine)
}

func (p *PyInstArchive) Open() bool {
	return true
}

func (p *PyInstArchive) Close() {
}

func (p *PyInstArchive) CheckFile() bool {
	appendLog(fmt.Sprintf("[+] Processing %s\n", p.inFilePath))

	var searchChunkSize int64 = 8192
	endPosition := p.fileSize
	p.cookiePosition = -1

	if endPosition < int64(len(PYINST_MAGIC)) {
		appendLog("[!] Error : File is too short or truncated\n")
		return false
	}

	var startPosition, chunkSize int64
	for {
		if endPosition >= searchChunkSize {
			startPosition = endPosition - searchChunkSize
		} else {
			startPosition = 0
		}
		chunkSize = endPosition - startPosition
		if chunkSize < int64(len(PYINST_MAGIC)) {
			break
		}

		if _, err := p.fPtr.Seek(startPosition, io.SeekStart); err != nil {
			appendLog("[!] File seek failed\n")
			return false
		}
		var data []byte = make([]byte, searchChunkSize)
		p.fPtr.Read(data)

		if offs := bytes.Index(data, PYINST_MAGIC[:]); offs != -1 {
			p.cookiePosition = startPosition + int64(offs)
			break
		}
		endPosition = startPosition + int64(len(PYINST_MAGIC)) - 1

		if startPosition == 0 {
			break
		}
	}
	if p.cookiePosition == -1 {
		appendLog("[!] Error : Missing cookie, unsupported pyinstaller version or not a pyinstaller archive\n")
		return false
	}
	p.fPtr.Seek(p.cookiePosition+PYINST20_COOKIE_SIZE, io.SeekStart)

	var cookie []byte = make([]byte, 64)
	if _, err := p.fPtr.Read(cookie); err != nil {
		appendLog("[!] Failed to read cookie!\n")
		return false
	}

	cookie = bytes.ToLower(cookie)
	if bytes.Contains(cookie, []byte("python")) {
		p.pyInstVersion = 21
		appendLog("[+] Pyinstaller version: 2.1+\n")
	} else {
		p.pyInstVersion = 20
		appendLog("[+] Pyinstaller version: 2.0\n")
	}
	return true
}

func (p *PyInstArchive) GetCArchiveInfo() bool {
	failFunc := func() bool {
		appendLog("[!] Error : The file is not a pyinstaller archive\n")
		return false
	}

	getPyMajMinVersion := func(version int) (int, int) {
		if version >= 100 {
			return version / 100, version % 100
		}
		return version / 10, version % 10
	}

	printPythonVerLenPkg := func(pyMajVer, pyMinVer, lenPkg int) {
		appendLog(fmt.Sprintf("[+] Python version: %d.%d\n", pyMajVer, pyMinVer))
		appendLog(fmt.Sprintf("[+] Length of package: %d bytes\n", lenPkg))
	}

	calculateTocPosition := func(cookieSize, lengthOfPackage, toc, tocLen int) {
		// Additional data after the cookie
		tailBytes := p.fileSize - p.cookiePosition - int64(cookieSize)

		// Overlay is the data appended at the end of the PE
		p.overlaySize = int64(lengthOfPackage) + tailBytes
		p.overlayPosition = p.fileSize - p.overlaySize
		p.tableOfContentsPosition = p.overlayPosition + int64(toc)
		p.tableOfContentsSize = int64(tocLen)
	}

	if _, err := p.fPtr.Seek(p.cookiePosition, io.SeekStart); err != nil {
		return failFunc()
	}

	if p.pyInstVersion == 20 {
		var pyInst20Cookie PyInst20Cookie
		cookieBuf := make([]byte, PYINST20_COOKIE_SIZE)
		if _, err := p.fPtr.Read(cookieBuf); err != nil {
			return failFunc()
		}

		if err := restruct.Unpack(cookieBuf, binary.LittleEndian, &pyInst20Cookie); err != nil {
			return failFunc()
		}

		p.pythonMajorVersion, p.pythonMinorVersion = getPyMajMinVersion(pyInst20Cookie.PythonVersion)
		printPythonVerLenPkg(p.pythonMajorVersion, p.pythonMinorVersion, pyInst20Cookie.LengthOfPackage)

		calculateTocPosition(
			PYINST20_COOKIE_SIZE,
			pyInst20Cookie.LengthOfPackage,
			pyInst20Cookie.Toc,
			pyInst20Cookie.TocLen,
		)

	} else {
		var pyInst21Cookie PyInst21Cookie
		cookieBuf := make([]byte, PYINST21_COOKIE_SIZE)
		if _, err := p.fPtr.Read(cookieBuf); err != nil {
			return failFunc()
		}
		if err := restruct.Unpack(cookieBuf, binary.LittleEndian, &pyInst21Cookie); err != nil {
			return failFunc()
		}
		appendLog("[+] Python library file: " + string(bytes.Trim(pyInst21Cookie.PythonLibName, "\x00")) + "\n")
		p.pythonMajorVersion, p.pythonMinorVersion = getPyMajMinVersion(pyInst21Cookie.PythonVersion)
		printPythonVerLenPkg(p.pythonMajorVersion, p.pythonMinorVersion, pyInst21Cookie.LengthOfPackage)

		calculateTocPosition(
			PYINST21_COOKIE_SIZE,
			pyInst21Cookie.LengthOfPackage,
			pyInst21Cookie.Toc,
			pyInst21Cookie.TocLen,
		)
	}
	return true
}

func (p *PyInstArchive) ParseTOC() {
	const CTOCEntryStructSize = 18
	p.fPtr.Seek(p.tableOfContentsPosition, io.SeekStart)

	var parsedLen int64 = 0

	// Parse table of contents
	for {
		if parsedLen >= p.tableOfContentsSize {
			break
		}
		var ctocEntry CTOCEntry

		data := make([]byte, CTOCEntryStructSize)
		p.fPtr.Read(data)
		restruct.Unpack(data, binary.LittleEndian, &ctocEntry)

		nameBuffer := make([]byte, ctocEntry.EntrySize-CTOCEntryStructSize)
		p.fPtr.Read(nameBuffer)

		nameBuffer = bytes.TrimRight(nameBuffer, "\x00")
		if len(nameBuffer) == 0 {
			ctocEntry.Name = randomString()
			appendLog(fmt.Sprintf("[!] Warning: Found an unamed file in CArchive. Using random name %s\n", ctocEntry.Name))
		} else {
			ctocEntry.Name = string(nameBuffer)
		}

		p.tableOfContents = append(p.tableOfContents, ctocEntry)
		parsedLen += int64(ctocEntry.EntrySize)
	}
	appendLog(fmt.Sprintf("[+] Found %d files in CArchive\n", len(p.tableOfContents)))
}

func (p *PyInstArchive) ExtractFiles() {
	appendLog("[+] Beginning extraction...please standby\n")

	for _, entry := range p.tableOfContents {
		p.fPtr.Seek(p.overlayPosition+int64(entry.EntryPosition), io.SeekStart)
		data := make([]byte, entry.DataSize)
		p.fPtr.Read(data)

		if entry.ComressionFlag == 1 {
			var err error
			compressedData := data[:]
			data, err = zlibDecompress(compressedData)
			if err != nil {
				appendLog(fmt.Sprintf("[!] Error: Failed to decompress %s in CArchive, extracting as-is\n", entry.Name))
				p.writeRawData(entry.Name, compressedData)
				continue
			}

			if len(data) != entry.UncompressedDataSize {
				appendLog(fmt.Sprintf("[!] Warning: Decompressed size mismatch for file %s\n", entry.Name))
			}
		}

		if entry.TypeCompressedData == 'd' || entry.TypeCompressedData == 'o' {
			// d -> ARCHIVE_ITEM_DEPENDENCY
			// o -> ARCHIVE_ITEM_RUNTIME_OPTION
			// These are runtime options, not files
			continue
		}

		if entry.TypeCompressedData == 's' {
			// s -> ARCHIVE_ITEM_PYSOURCE
			// Entry point are expected to be python scripts
			appendLog(fmt.Sprintf("[+] Possible entry point: %s.pyc\n", entry.Name))
			if !p.gotPycMagic {
				// if we don't have the pyc header yet, fix them in a later pass
				p.barePycsList = append(p.barePycsList, &barePyc{entry.Name + ".pyc", data})
			} else {
				p.writePyc(entry.Name+".pyc", data)
			}
		} else if entry.TypeCompressedData == 'M' || entry.TypeCompressedData == 'm' {
			// M -> ARCHIVE_ITEM_PYPACKAGE
			// m -> ARCHIVE_ITEM_PYMODULE
			// packages and modules are pyc files with their header intact

			// From PyInstaller 5.3 and above pyc headers are no longer stored
			// https://github.com/pyinstaller/pyinstaller/commit/a97fdf
			if data[2] == '\r' && data[3] == '\n' {
				// < pyinstaller 5.3
				if !p.gotPycMagic {
					copy(p.pycMagic[:], data[0:4])
					p.gotPycMagic = true
				}
				p.writeRawData(entry.Name+".pyc", data)
			} else {
				// >= pyinstaller 5.3
				if !p.gotPycMagic {
					// if we don't have the pyc header yet, fix them in a later pass
					p.barePycsList = append(p.barePycsList, &barePyc{entry.Name + ".pyc", data})
				} else {
					p.writePyc(entry.Name+".pyc", data)
				}
			}
		} else if entry.TypeCompressedData == 'z' || entry.TypeCompressedData == 'Z' {
			if p.pythonMajorVersion == 3 {
				p.extractPYZ(entry.Name, data)
			} else {
				appendLog(fmt.Sprintf("[!] Skipping pyz extraction as Python %d.%d is not supported\n", p.pythonMajorVersion, p.pythonMinorVersion))
				p.writeRawData(entry.Name, data)
			}
		} else {
			p.writeRawData(entry.Name, data)
		}
	}
	p.fixBarePycs()
}

func (p *PyInstArchive) fixBarePycs() {
	for _, pycFile := range p.barePycsList {
		p.writePyc(pycFile.filepath, pycFile.contents)
	}
}

func (p *PyInstArchive) extractPYZ(path string, pyzData []byte) {
	dirName := path + "_extracted"
	f := bytes.NewReader(pyzData)

	var pyzMagic []byte = make([]byte, 4)
	f.Read(pyzMagic)
	if !bytes.Equal(pyzMagic, []byte("PYZ\x00")) {
		appendLog("[!] Magic header in PYZ archive doesn't match\n")
	}

	var pyzPycMagic []byte = make([]byte, 4)
	f.Read(pyzPycMagic)

	if !p.gotPycMagic {
		copy(p.pycMagic[:], pyzPycMagic)
		p.gotPycMagic = true
	} else if !bytes.Equal(p.pycMagic[:], pyzPycMagic) {
		copy(p.pycMagic[:], pyzPycMagic)
		p.gotPycMagic = true
		appendLog("[!] Warning: pyc magic of files inside PYZ archive are different from those in CArchive\n")
	}

	var pyzTocPositionBytes []byte = make([]byte, 4)
	f.Read(pyzTocPositionBytes)
	pyzTocPosition := binary.BigEndian.Uint32(pyzTocPositionBytes)
	f.Seek(int64(pyzTocPosition), io.SeekStart)

	su := marshal.NewUnmarshaler(f)
	obj := su.Unmarshal()
	if obj == nil {
		appendLog("Unmarshalling failed\n")
	} else {
		listobj := obj.(*marshal.PyListObject)
		listobjItems := listobj.GetItems()
		appendLog(fmt.Sprintf("[+] Found %d files in PYZArchive\n", len(listobjItems)))

		for _, item := range listobjItems {
			item := item.(*marshal.PyListObject)
			name := item.GetItems()[0].(*marshal.PyStringObject).GetString()

			ispkg_position_length_tuple := item.GetItems()[1].(*marshal.PyListObject)
			ispkg := ispkg_position_length_tuple.GetItems()[0].(*marshal.PyIntegerObject).GetValue()
			position := ispkg_position_length_tuple.GetItems()[1].(*marshal.PyIntegerObject).GetValue()
			length := ispkg_position_length_tuple.GetItems()[2].(*marshal.PyIntegerObject).GetValue()

			// Prevent writing outside dirName
			filename := strings.ReplaceAll(name, "..", "__")
			filename = strings.ReplaceAll(filename, ".", string(os.PathSeparator))

			var filenamepath string
			if ispkg == 1 {
				filenamepath = filepath.Join(dirName, filename, "__init__.pyc")
			} else {
				filenamepath = filepath.Join(dirName, filename+".pyc")
			}

			f.Seek(int64(position), io.SeekStart)

			var compressedData []byte = make([]byte, length)
			f.Read(compressedData)

			decompressedData, err := zlibDecompress(compressedData)
			if err != nil {
				appendLog(fmt.Sprintf("[!] Error: Failed to decompress %s in PYZArchive, likely encrypted. Extracting as is\n", filenamepath))
				p.writeRawData(filenamepath+".pyc.encrypted", compressedData)
			} else {
				p.writePyc(filenamepath, decompressedData)
			}
		}
	}
	// f.Close()
}

func (p *PyInstArchive) writePyc(path string, data []byte) {
	f, err := p.outZip.CreateHeader(&zip.FileHeader{
		Name:   path,
		Method: zip.Store,
	})

	if err != nil {
		appendLog(fmt.Sprintf("[!] Failed to write file %s\n", path))
		return
	}
	// pyc magic
	f.Write(p.pycMagic[:])

	if p.pythonMajorVersion >= 3 && p.pythonMinorVersion >= 7 {
		// PEP 552 -- Deterministic pycs
		f.Write([]byte{0, 0, 0, 0})             //Bitfield
		f.Write([]byte{0, 0, 0, 0, 0, 0, 0, 0}) //(Timestamp + size) || hash
	} else {
		f.Write([]byte{0, 0, 0, 0}) //Timestamp
		if p.pythonMajorVersion >= 3 && p.pythonMinorVersion >= 3 {
			f.Write([]byte{0, 0, 0, 0})
		}
	}
	f.Write(data)
	p.outZip.Flush()
}

func (p *PyInstArchive) writeRawData(path string, data []byte) {
	path = strings.Trim(path, "\x00")
	path = strings.ReplaceAll(path, "\\", string(os.PathSeparator))
	path = strings.ReplaceAll(path, "/", string(os.PathSeparator))
	path = strings.ReplaceAll(path, "..", "__")

	f, _ := p.outZip.CreateHeader(&zip.FileHeader{
		Name:   path,
		Method: zip.Store,
	})

	f.Write(data)
}

func main() {
	js.Global.Set("extract_exe", extract_exe)
}

func extract_exe(fileName string, inbuf []byte, logFn *js.Object) []byte {
	logFunc = logFn
	var zipData bytes.Buffer
	arch := PyInstArchive{
		outZip:     zip.NewWriter(&zipData),
		inFilePath: fileName,
		fPtr:       bytes.NewReader(inbuf),
		fileSize:   int64(len(inbuf)),
	}

	if arch.Open() {
		if arch.CheckFile() {
			if arch.GetCArchiveInfo() {
				arch.ParseTOC()
				arch.ExtractFiles()
				appendLog(fmt.Sprintf("[+] Successfully extracted pyinstaller archive: %s\n", fileName))
				appendLog("\nYou can now use a python decompiler on the pyc files within the extracted directory\n")
				arch.outZip.Close()
				return zipData.Bytes()
			}
		}
		arch.Close()
	}
	return nil
}
