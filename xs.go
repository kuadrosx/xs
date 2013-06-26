package main

import (
    "os"
    "os/exec"
    "strings"
    "regexp"
    "strconv"
    "fmt"
)

func main() {
  pattern := os.Args[1]
  cmd := exec.Command("apt-cache", "search", pattern)
  out, err := cmd.Output()
  if err != nil {
    panic(err)
  }
  matches := 0
  for _,line := range(strings.Split(string(out), "\n")) {
    if Parse(line, pattern) {
      matches += 1
    }
  }
  fmt.Println("Found "+strconv.Itoa(matches)+" matches")
}

func Parse(pkgdesc string, pattern string) bool {
  var parser = regexp.MustCompile(`([\w\-\_\.\+]+)\s-\s(.+)$`)
  data := parser.FindStringSubmatch(pkgdesc)
  if len(data) > 0 {
    pkgname := data[1]
    pattern_parser := regexp.MustCompile(regexp.QuoteMeta(pattern))
    if pattern_parser.MatchString(pkgname) {
       return Show(pkgname)
    }
  }
  return false
}


func Show(pkg string) bool {
  out, err := exec.Command("apt-cache", "show", pkg).Output()
  if err != nil {
    return false
  }
  result := make(map[string] string, 0)
  parser := regexp.MustCompile(`([\w\-]+):\s(.+)$`)
  for _,line := range(strings.Split(string(out), "\n")) {
    data := parser.FindStringSubmatch(line)
    if len(data) > 0 {
      result[strings.ToLower(data[1])] = data[2]
    }
  }
  installed := isInstalled(pkg)

  fmt.Print(" \033[92m*\033[00m "+result["section"]+"/\033[01m"+result["package"]+"\033[00m  ")
  if installed {
    fmt.Println("[\033[01;32mINSTALLED\033[00m]")
  } else {
    fmt.Println()
  }

  fmt.Println("     \033[32mVersion:\033[00m \033[96m"+result["version"]+"\033[00m")
  if len(result["homepage"]) > 0 {
    fmt.Println("     \033[32mHomepage:\033[00m "+result["homepage"])
  }
  fmt.Println("     \033[32mDescription:\033[00m "+result["description-en"])

  size, _ := strconv.Atoi(result["size"])
  str := strconv.Itoa((size + 1023) / 1024)
  fmt.Println("     \033[32mDownload size:\033[00m "+ str +" KiB")
  fmt.Println()
  return true
}

func isInstalled(pkg string) bool {
  out, _ := exec.Command("dpkg", "-l", pkg).Output()
  parser := regexp.MustCompile(`^ii\s+`+regexp.QuoteMeta(pkg)+`\s+`)
  for _,line := range(strings.Split(string(out), "\n")) {
    if parser.MatchString(line) {
      return true
    }
  }
  return false
}
