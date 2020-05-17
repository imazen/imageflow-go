package imageflow

import (
	"io/ioutil"
	"net/http"
)

type ioOperation interface {
	toBuffer() ([]byte, error)
	toOutput([]byte, map[string][]byte) map[string][]byte
	setIo(id uint)
	getIo() uint
}

func (file File) toBuffer() ([]byte, error) {
	bytes, errorInRead := ioutil.ReadFile(file.filename)
	if errorInRead != nil {
		return nil, errorInRead
	}
	return bytes, nil
}

func (file File) toOutput(data []byte, m map[string][]byte) map[string][]byte {
	ioutil.WriteFile(file.filename, data, 0644)
	return m
}

func (file *File) setIo(id uint) {
	file.iOID = id
}

func (file File) getIo() uint {
	return file.iOID
}

func (file URL) toBuffer() ([]byte, error) {
	bytes, errorInURL := http.Get(file.url)
	if errorInURL != nil {
		return nil, errorInURL
	}
	data, errorInRead := ioutil.ReadAll(bytes.Body)
	if errorInRead != nil {
		return nil, errorInRead
	}
	return data, nil
}

func (file URL) toOutput(data []byte, m map[string][]byte) map[string][]byte {
	return m
}

func (file *URL) setIo(id uint) {
	file.iOID = id
}

func (file URL) getIo() uint {
	return file.iOID
}

// URL is used to make a http request to get file and use it
type URL struct {
	url  string
	iOID uint
}

// NewURL is used to create a new url operation
func NewURL(url string) *URL {
	return &URL{
		url: url,
	}
}

// NewBuffer create a buffer operation
func NewBuffer(buffer []byte) *Buffer {
	return &Buffer{
		buffer: buffer,
	}
}

// GetBuffer is used to get key
func GetBuffer(key string) *Buffer {
	return &Buffer{
		key: key,
	}
}

// Buffer is io operation related to []byte
type Buffer struct {
	iOID   uint
	buffer []byte
	key    string
}

func (file Buffer) toBuffer() ([]byte, error) {
	return file.buffer, nil
}

func (file Buffer) toOutput(data []byte, m map[string][]byte) map[string][]byte {
	m[file.key] = data
	return m
}

func (file *Buffer) setIo(id uint) {
	file.iOID = id
}

func (file Buffer) getIo() uint {
	return file.iOID
}

// File is io operation related to file
type File struct {
	iOID     uint
	filename string
}

// NewFile is used to create a file io
func NewFile(filename string) *File {
	return &File{
		filename: filename,
	}
}
