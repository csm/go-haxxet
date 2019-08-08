package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/hpcloud/tail"
	"github.com/mgutz/ansi"
)

var lineCountFlag int
var help bool

func init() {
	flag.IntVar(&lineCountFlag, "count", 10, "Number of lines to display per file after opening")
	flag.BoolVar(&help, "help", false, "Show this help and exit")
}

func main() {
	os.Exit(runMain())
}

func usage() int {
	fmt.Println("usage: haxxet file1[:color] [file2[:color] ...]")
	fmt.Println("")
	fmt.Println("Each entry is a file path to follow and print lines to the console.")
	fmt.Println("An optional color name can be given after the file name, which will")
	fmt.Println("color the output with that ANSI color name.")
	flag.PrintDefaults()
	return 0
}

func tailOffset(path string, lineCount int) (int64, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	fileLen, err := file.Seek(0, os.SEEK_END)
	if err != nil {
		return 0, err
	}
	var offset int64 = fileLen - 1
	buffer := make([]byte, 1)
	for lineCount > 0 && offset > 0 {
		n, err := file.ReadAt(buffer, offset)
		if err != nil {
			return 0, err
		} else if n > 0 && buffer[0] == 0xa || buffer[0] == 0xd {
			lineCount = lineCount - 1
		}
		offset = offset - 1
	}
	return offset, nil
}

type Line struct {
	path string
	text string
	err  error
}

func tailFile(c chan *Line, path, color string, lineCount int) {
	offset, err := tailOffset(path, lineCount)
	if err != nil {
		c <- &Line{path: path, text: "", err: err}
	}
	seek := tail.SeekInfo{Offset: offset, Whence: os.SEEK_SET}
	t, err := tail.TailFile(path, tail.Config{Follow: true, MustExist: true, ReOpen: true, Location: &seek, Logger: tail.DiscardingLogger})
	p := strings.Split(path, "/")
	fileName := p[len(p)-1]
	if err != nil {
		c <- &Line{path: path, text: "", err: err}
	} else {
		for line := range t.Lines {
			text := color + fileName + ": " + line.Text + ansi.Reset
			c <- &Line{path: path, text: text, err: nil}
		}
	}
}

func runMain() int {
	flag.Parse()
	args := flag.Args()
	if help || len(args) == 0 {
		return usage()
	}
	c := make(chan *Line)
	for _, arg := range args {
		var path = ""
		var color = ""
		parts := strings.Split(arg, ":")
		if len(parts) == 1 {
			path = parts[0]
			color = ansi.Reset
		} else {
			path = strings.Join(parts[0:len(parts)-1], ":")
			color = ansi.ColorCode(parts[len(parts)-1])
		}
		go tailFile(c, path, color, lineCountFlag)
	}
	for {
		line := <-c
		if line.err != nil {
			fmt.Println("error opening", line.path, line.err)
			return 1
		} else {
			fmt.Println(line.text)
		}
	}
	return 0
}
