package file

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type File struct {
	FilePath   string
	buffer     *bytes.Buffer
	OutputFile *os.File
}

func NewFile() *File {
	return &File{
		buffer: &bytes.Buffer{},
	}
}

func (f *File) SetFile(fileName, path string) error {
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		log.Printf("error creating dir: %s", f.FilePath)
		return err
	}

	f.FilePath = filepath.Join(path, fileName)
	file, err := os.Create(f.FilePath)
	if err != nil {
		log.Printf("error creating file: %s", f.FilePath)
		return err
	}
	f.OutputFile = file
	return nil
}

func (f *File) Write(chunk []byte) error {
	if f.OutputFile == nil {
		return fmt.Errorf("output file is not set")
	}
	_, err := f.OutputFile.Write(chunk)
	return err
}

func (f *File) Close() error {
	if f.OutputFile == nil {
		return nil
	}
	return f.OutputFile.Close()
}
