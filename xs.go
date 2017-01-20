package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

var parser = regexp.MustCompile(`([\w\-\_\.\+]+)\s-\s(.+)$`)

func main() {
	pattern := os.Args[1]

	ch := Show(PkgInfo(Parse(pattern, Search(pattern))))
	matches := 0
	for info := range ch {
		fmt.Print(info)
		matches += 1
	}
	fmt.Printf("Found %d matches\n", matches)
}

func Search(pattern string) <-chan string {
	cmd := exec.Command("apt-cache", "search", pattern)
	results, err := cmd.Output()
	if err != nil {
		panic(err)
	}

	out := make(chan string)
	go func() {
		for _, line := range strings.Split(string(results), "\n") {
			out <- line
		}
		close(out)
	}()
	return out
}

func Parse(pattern string, in <-chan string) <-chan string {
	out := make(chan string)
	go func() {
		for pkgdesc := range in {
			data := parser.FindStringSubmatch(pkgdesc)
			if len(data) > 0 {
				pkgname := data[1]
				pattern_parser := regexp.MustCompile(regexp.QuoteMeta(pattern))
				if pattern_parser.MatchString(pkgname) {
					out <- pkgname
				}
			}
		}
		close(out)
	}()
	return out
}

func PkgInfo(in <-chan string) <-chan map[string]string {
	out := make(chan map[string]string)
	go func() {
		for pkg := range in {
			cmdout, err := exec.Command("apt-cache", "show", pkg).Output()
			if err == nil {
				result := make(map[string]string, 0)
				parser := regexp.MustCompile(`([\w\-]+):\s(.+)$`)
				for _, line := range strings.Split(string(cmdout), "\n") {
					data := parser.FindStringSubmatch(line)
					if len(data) > 0 {
						result[strings.ToLower(data[1])] = data[2]
					}
				}

				result["installed"] = isInstalled(pkg)
				out <- result
			}
		}
		close(out)
	}()
	return out
}

func Show(in <-chan map[string]string) <-chan string {
	out := make(chan string)

	go func() {
		for result := range in {
			var buffer bytes.Buffer
			buffer.WriteString(" \033[92m*\033[00m ")
			buffer.WriteString(result["section"])
			buffer.WriteString("/\033[01m")
			buffer.WriteString(result["package"])
			buffer.WriteString("\033[00m  ")

			if result["installed"] == "y" {
				buffer.WriteString("[\033[01;32mINSTALLED\033[00m]\n")
			} else {
				buffer.WriteString("\n")
			}

			buffer.WriteString("     \033[32mVersion:\033[00m \033[96m")
			buffer.WriteString(result["version"])
			buffer.WriteString("\033[00m\n")

			if len(result["homepage"]) > 0 {
				buffer.WriteString("     \033[32mHomepage:\033[00m ")
				buffer.WriteString(result["homepage"])
				buffer.WriteString("\n")
			}

			buffer.WriteString("     \033[32mDescription:\033[00m ")
			buffer.WriteString(result["description-en"])
			buffer.WriteString("\n")

			size, _ := strconv.Atoi(result["size"])
			str := strconv.Itoa((size + 1023) / 1024)

			buffer.WriteString("     \033[32mDownload size:\033[00m ")
			buffer.WriteString(str)
			buffer.WriteString(" KiB\n\n")

			out <- buffer.String()
		}
		close(out)
	}()

	return out
}

func isInstalled(pkg string) string {
	out, _ := exec.Command("dpkg", "-l", pkg).Output()
	parser := regexp.MustCompile(`^ii\s+` + regexp.QuoteMeta(pkg) + `\s+`)
	for _, line := range strings.Split(string(out), "\n") {
		if parser.MatchString(line) {
			return "y"
		}
	}
	return "n"
}
