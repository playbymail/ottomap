// Copyright (c) 2026 Michael D Henderson. All rights reserved.

// Package tndocx implements a parser for Word DOCX files.
// It's more of an adapter than a parser.
package tndocx

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// ParseBytes parses from a []byte that contains a ZIP file.
// Injects a line-feed at the end of every paragraph.
func ParseBytes(b []byte, trimLeading, trimTrailing bool) ([]byte, error) {
	return ParseBytesReader(bytes.NewReader(b), trimLeading, trimTrailing)
}

// ParseBytesReader parses from a bytes.Reader that contains a ZIP file.
// Injects a line-feed at the end of every paragraph.
func ParseBytesReader(r *bytes.Reader, trimLeading, trimTrailing bool) ([]byte, error) {
	zr, err := zip.NewReader(r, r.Size())
	if err != nil {
		return nil, fmt.Errorf("bad input: %w", err)
	}
	text, err := parseZipFS(zr.File)
	if err != nil {
		return nil, err
	}
	return trimOptions(text, trimLeading, trimTrailing), nil
}

// ParsePath opens and parses a .docx file.
// Injects a line-feed at the end of every paragraph.
func ParsePath(path string, trimLeading, trimTrailing bool) ([]byte, error) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("open docx: %w", err)
	}
	defer zr.Close()
	text, err := parseZipFS(zr.File)
	if err != nil {
		return nil, err
	}
	return trimOptions(text, trimLeading, trimTrailing), nil
}

func parseZipFS(files []*zip.File) ([]byte, error) {
	for _, file := range files {
		if file.Name == "word/document.xml" {
			rc, err := file.Open()
			if err != nil {
				return nil, fmt.Errorf("open document.xml: %w", err)
			}
			defer rc.Close()
			text, err := translateWordXML(rc)
			if err != nil {
				return nil, err
			}
			return text, nil
		}
	}
	return nil, ErrWordXmlDocumentNotFound
}

// translateWordXML actually walks the XML and builds the text.
func translateWordXML(r io.Reader) ([]byte, error) {
	dec := xml.NewDecoder(r)

	var (
		buf      = &bytes.Buffer{}
		preserve = false // xml:space="preserve" on current <w:t>
		inT      = false // we're inside a <w:t>

		// if we ever need to do something in a paragraph
		// seenTextInP = false // to know if paragraph has any text
		// inP         = false // we're inside a paragraph
	)

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("xml decode: %w", err)
		}

		switch se := tok.(type) {
		case xml.StartElement:
			local := se.Name.Local

			switch local {
			case "p":
				//inP, seenTextInP = true, false

			case "t":
				inT = true
				preserve = false
				// check attributes for xml:space="preserve"
				for _, a := range se.Attr {
					if (a.Name.Local == "space" || strings.HasSuffix(a.Name.Local, "space")) &&
						a.Value == "preserve" {
						preserve = true
						break
					}
				}

			case "tab":
				buf.WriteByte('\t')
				//seenTextInP = true

			case "br":
				buf.WriteByte('\n')
				//seenTextInP = true
			}

		case xml.EndElement:
			local := se.Name.Local
			switch local {
			case "t":
				inT = false
				preserve = false
			case "p":
				// inject a line-feed at the end of every paragraph
				buf.WriteByte('\n')
				//inP, seenTextInP = false, false
			}

		case xml.CharData:
			if inT {
				text := string(se)
				if preserve {
					buf.WriteString(text)
				} else {
					// most Word text here is already “normal”, but, just
					// in case, don’t trim aggressively — just write it.
					buf.WriteString(text)
				}
				//seenTextInP = true
			}
		}
	}

	return buf.Bytes(), nil
}

func trimOptions(text []byte, trimLeading, trimTrailing bool) []byte {
	if trimLeading == false && trimTrailing == false {
		return text
	}
	const asciiSpace = " \t\n\v\f\r"
	lines := bytes.Split(text, []byte{'\n'})
	if trimLeading && trimTrailing {
		for i, line := range lines {
			lines[i] = bytes.TrimSpace(line)
		}
	} else if trimLeading {
		for i, line := range lines {
			lines[i] = bytes.TrimLeft(line, asciiSpace)
		}
	} else {
		for i, line := range lines {
			lines[i] = bytes.TrimRight(line, asciiSpace)
		}
	}
	return bytes.Join(lines, []byte{'\n'})
}

type Error string

func (e Error) Error() string {
	return string(e)
}

const (
	ErrWordXmlDocumentNotFound = Error("word/document.xml not found")
)
