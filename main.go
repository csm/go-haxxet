package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/hpcloud/tail"
	"github.com/mgutz/ansi"
)

func main() {
	os.Exit(runMain())
}

func usage() int {
	fmt.Println("usage: haxxet file1[:color] [file2[:color] ...]")
	fmt.Println("")
	fmt.Println("Each entry is a file path to follow and print lines to the console.")
	fmt.Println("An optional color name can be given after the file name, which will")
	fmt.Println("color the output with that ANSI color name.")
	return 0
}

func tailFile(c chan string, path, color string) {
	seek := tail.SeekInfo{Offset: 0, Whence: os.SEEK_END}
	t, err := tail.TailFile(path, tail.Config{Follow: true, MustExist: true, ReOpen: true, Location: &seek, Logger: tail.DiscardingLogger})
	if err != nil {
		fmt.Println("error opening", path, ":", err)
		os.Exit(1)
	} else {
		for line := range t.Lines {
			text := color + line.Text + ansi.Reset
			c <- text
		}
	}
}

func runMain() int {
	if len(os.Args) == 1 {
		return usage()
	}
	for _, arg := range os.Args[1:] {
		if arg == "-h" || arg == "-help" || arg == "--help" {
			return usage()
		}
	}
	c := make(chan string)
	for _, arg := range os.Args[1:] {
		var path = ""
		var color = ""
		parts := strings.Split(arg, ":")
		fmt.Println("parts:", parts)
		if len(parts) == 1 {
			path = parts[0]
			color = ansi.Reset
		} else {
			path = strings.Join(parts[0:len(parts)-1], ":")
			color = ansi.ColorCode(parts[len(parts)-1])
		}
		go tailFile(c, path, color)
	}
	for {
		line := <-c
		fmt.Println(line)
	}
	return 0
}
