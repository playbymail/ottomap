// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package stdlib

import (
	"crypto/sha1"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"time"
)

type File_t struct {
	Path     string    // full path to file
	Name     string    // file name
	Kind     string    // docx, txt
	Year     int       // year from file name
	Month    int       // month from file name
	Unit     string    // unit name from file name
	Hash     string    // SHA1 hash of the file contents
	Modified time.Time // last modified time, hopefully always UTC
}

var (
	rxTurnReportFile = regexp.MustCompile(`^(\d{4})-(\d{2})\.(\d{4}([cefg]\d)?)\.report\.(docx|txt)$`)
)

// FindAllInputs returns a list of all DOCX and TXT files in the requested path.
// The list is sorted by timestamp and then name.
func FindAllInputs(path string) ([]*File_t, error) {
	// search for report files in the requested path
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var list []*File_t
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		} else if rxTurnReportFile.FindStringSubmatch(entry.Name()) == nil {
			continue
		}
		item, err := FindInput(path, entry.Name())
		if err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	// sort files by Modified time, then name
	sort.Slice(list, func(i, j int) bool {
		if list[i].Modified.Before(list[j].Modified) {
			return true
		} else if list[i].Modified.Equal(list[j].Modified) {
			return list[i].Name < list[j].Name
		}
		return false
	})
	return list, nil
}

// FindInputs returns a list containing the input files in the requested path that match the requested names.
// The list is sorted by timestamp and then name.
func FindInputs(path string, names ...string) ([]*File_t, error) {
	// verify that all names match the pattern
	for _, name := range names {
		if rxTurnReportFile.FindStringSubmatch(name) == nil {
			return nil, fmt.Errorf("%s: invalid file name", name)
		}
	}
	// search for report files in the requested path
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var list []*File_t
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		matchesNames := false
		for _, name := range names {
			if entry.Name() == name {
				matchesNames = true
				break
			}
		}
		if !matchesNames {
			continue
		}
		item, err := FindInput(path, entry.Name())
		if err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	// sort files by Modified time, then name
	sort.Slice(list, func(i, j int) bool {
		if list[i].Modified.Before(list[j].Modified) {
			return true
		} else if list[i].Modified.Equal(list[j].Modified) {
			return list[i].Name < list[j].Name
		}
		return false
	})
	return list, nil
}

// FindInput returns a *File_t for the input file in the requested path that matches the requested name.
func FindInput(path string, name string) (*File_t, error) {
	item := &File_t{
		Path: path,
		Name: name,
	}
	if matches := rxTurnReportFile.FindStringSubmatch(name); matches == nil {
		return nil, fmt.Errorf("%s: invalid file name", name)
	} else {
		item.Year, _ = strconv.Atoi(matches[1])
		item.Month, _ = strconv.Atoi(matches[2])
		item.Unit = matches[3]
		if matches[4] == "txt" {
			item.Kind = "text"
		} else {
			item.Kind = "word"
		}
	}
	// verify that the file exists and get the last modified time
	if sb, err := os.Stat(filepath.Join(path, name)); err != nil {
		return nil, err
	} else if sb.IsDir() {
		return nil, fmt.Errorf("file is a directory")
	} else if !sb.Mode().IsRegular() {
		return nil, fmt.Errorf("file is not a regular file")
	} else {
		item.Modified = sb.ModTime().UTC()
	}
	// load and hash the file. return any errors loading or hashing the file.
	if data, err := os.ReadFile(filepath.Join(path, name)); err != nil {
		return nil, err
	} else {
		hashValue := sha1.New()
		hashValue.Write(data)
		item.Hash = fmt.Sprintf("%x", hashValue.Sum(nil))
	}
	return item, nil
}
