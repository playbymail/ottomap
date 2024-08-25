// Package office implements a reader for Word documents.
package office

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"os"
	"regexp"
	"strings"
)

// http://officeopenxml.com/anatomyofOOXML.php

// copied from https://github.com/lu4p/cat/blob/master/docxtxt/docxreader.go
// and licensed as provided in the COPYING file in this folder.

// docx zip struct
type docx struct {
	zipFileReader *zip.ReadCloser
	Files         []*zip.File
	FilesContent  map[string][]byte
	WordsList     []*words
}

type words struct {
	Content []string
}

// ToStr converts a .docx document file to string
func ToStr(filename string) (string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return BytesToStr(content)
}

// BytesToStr converts a []byte representation of .docx document file to string
func BytesToStr(data []byte) (string, error) {
	reader := bytes.NewReader(data)
	d, err := openDocxReader(reader)
	if err != nil {
		return "", err
	}
	d.GenWordsList()
	result := strings.Builder{}
	for _, word := range d.WordsList {
		for _, content := range word.Content {
			result.WriteByte('+')
			result.WriteString(content)
		}
		result.WriteByte('\n')
	}
	return result.String(), nil
}

// openDocxReader open and load all readers content
func openDocxReader(bytesReader *bytes.Reader) (*docx, error) {
	reader, err := zip.NewReader(bytesReader, bytesReader.Size())
	if err != nil {
		return nil, err
	}

	wordDoc := docx{
		zipFileReader: nil,
		Files:         reader.File,
		FilesContent:  map[string][]byte{},
	}

	for _, f := range wordDoc.Files {
		contents, _ := wordDoc.retrieveFileContents(f.Name)
		wordDoc.FilesContent[f.Name] = contents
	}

	return &wordDoc, nil
}

// Read all files contents
func (d *docx) retrieveFileContents(filename string) ([]byte, error) {
	var file *zip.File
	for _, f := range d.Files {
		if f.Name == filename {
			file = f
		}
	}

	if file == nil {
		return nil, errors.New(filename + " file not found")
	}

	reader, err := file.Open()
	if err != nil {
		return nil, err
	}

	return io.ReadAll(reader)
}

// GenWordsList generate a list of all words
func (d *docx) GenWordsList() {
	xmlData := string(d.FilesContent["word/document.xml"])
	d.listP(xmlData)
}

var (
	reRunT = regexp.MustCompile(`(?U)(<w:r>|<w:r .*>)(.*)(</w:r>)`)
	reT    = regexp.MustCompile(`(?U)(<w:t>|<w:t .*>)(.*)(</w:t>)`)
)

// get w:t value
func (d *docx) getT(item string) {
	var subStr string
	data := item
	w := new(words)
	content := []string{}

	wrMatch := reRunT.FindAllStringSubmatchIndex(data, -1)
	// loop r
	for _, rMatch := range wrMatch {
		rData := data[rMatch[4]:rMatch[5]]
		wtMatch := reT.FindAllStringSubmatchIndex(rData, -1)
		for _, match := range wtMatch {
			subStr = rData[match[4]:match[5]]
			content = append(content, subStr)
		}
	}
	w.Content = content
	d.WordsList = append(d.WordsList, w)
}

var (
	reP = regexp.MustCompile(`(?U)<w:p[^>]*>(.*)</w:p>`)
)

// hasP identify the paragraph
func hasP(data string) bool {
	result := reP.MatchString(data)
	return result
}

var (
	reListP = regexp.MustCompile(`(?U)<w:p[^>]*(.*)</w:p>`)
)

// listP for w:p tag value
func (d *docx) listP(data string) {
	var result []string
	for _, match := range reListP.FindAllStringSubmatch(data, -1) {
		result = append(result, match[1])
	}
	for _, item := range result {
		if hasP(item) {
			d.listP(item)
			continue
		}
		d.getT(item)
	}
}
