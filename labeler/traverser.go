package labeler

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var acceptedExt = map[string]bool{
	".jpg":  true,
	".png":  true,
	".jpeg": true,
}

// Traverse walks through all the file in a directory
func Traverse(root string, ch chan string) {
	root = filepath.Clean(root)
	defer close(ch)
	defer func() {
		if r := recover(); r != nil {
			log.Println("error traversing the path:", r)
		}
	}()

	walk(root, ch)
}

func walk(root string, ch chan string) {
	info, err := os.Stat(root)
	if err != nil {
		panic("cannot read file status")
	}

	if !info.IsDir() {
		msg := fmt.Errorf("labeler:walk: %v is not a directory", info.Name())
		panic(msg)
	}

	dir, err := os.Open(root)
	if err != nil {
		panic(err)
	}
	defer dir.Close()

	for {
		fileInfo, err := dir.Readdir(1)
		if err != nil {
			if err == io.EOF {
				break
			}

			log.Printf("Cannot read file info: %v\n", err)
			continue
		}

		if file := fileInfo[0]; file.IsDir() {
			walk(root+"/"+file.Name(), ch)
		} else {
			ext := strings.ToLower(filepath.Ext(file.Name()))
			if acceptedExt[ext] {
				ch <- root + "/" + file.Name()
			}
		}
	}
}
