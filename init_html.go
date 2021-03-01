package main

import (
//    "net/http"
    "fmt"
	"strings"
	"os"
//	"errors"
//    "github.com/gorilla/mux"
)


func init_html_file(filnam string, opt bool) (nfilnam string, err error) {
	var tdir string
	fmt.Println("creating skeleton file: ", filnam, "del: ", opt)
// check for empty file
	if len(filnam) < 2 {
		return "error", fmt.Errorf("error init -- filename too short! %v",err)
	}
// check whether file + directory
	file_idx  := strings.LastIndex(filnam, "/")
	if file_idx > 0 {
		tdir = filnam[:file_idx]
//		fmt.Println("dir: ", tdir)
// check whether filnam is directory
		info, err := os.Stat(tdir)
		if os.IsNotExist(err) {
			return "error", fmt.Errorf("error init -- directory %v does not exist! %v", tdir, err)
		}
		if !info.IsDir() {
			return "error", fmt.Errorf("error init -- argument in cmd %v is a file not a directory! %v", err)
		}
	}

// check whether file has an .html extension
// if file does not have an extension, add one

//	hasext := true
	if !strings.Contains(filnam, ".html") {
//		hasext = false
		filnam += ".html"
	}

//	fmt.Println("file: ", filnam, "del: ", opt)

	_, err = os.Stat(filnam)

	if err == nil {
		fmt.Println("file: ", filnam, " already exists!")
//		return fmt.Errorf("error init -- file %v already exists! %v", filnam, err)

		if !opt {
			return "error", fmt.Errorf("error init -- no option set to delete file %v!", filnam)
		}
		fmt.Println("deleting existing file!")
		err = os.Remove(filnam)
		if err != nil {
				return "error", fmt.Errorf("error init -- could not delete file %v! %v", filnam, err)
		}
	}
	fmt.Println("creating new skeleton file name: ", filnam)

	fil, err := os.Create(filnam)
	fil.WriteString("<!DOCTYPE html>\n")
	fil.WriteString("<html lang=\"en\">\n")
	fil.WriteString("<head>\n")

	fil.WriteString("  <meta charset=\"UTF-8\">\n")
	fil.WriteString("  <meta name=\"description\" content=\"blog about software writing and using\">\n")
	fil.WriteString("  <meta name=\"keywords\" content=\"Go\">\n")
	fil.WriteString("  <meta name=\"author\" content=\"prr\">\n")
	fil.WriteString("  <meta name=\"date\" content=\"1\\3\\2021\">\n")
	fil.WriteString("  <meta  name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n")


	fil.WriteString("  <title>Here is the Title</title>\n")
	fil.WriteString("</head>\n")
	fil.WriteString("<body>\n")
	fil.WriteString("  <h1>Hello Index</h1>\n")
	fil.WriteString("</body>\n")
	fil.WriteString("</html>\n")

	fil.Close()
// remove file
//	os.remove(path)

	return nfilnam, nil
}

func main() {

	arg_num := len(os.Args)
	opt := false
//	fmt.Println("args: ", arg_num)

	if arg_num < 2 {
		fmt.Println("error -- insufficient arguments!")
		fmt.Println("usage: ./init_html file opt")
		os.Exit(1)
	}

	if arg_num > 3 {
		fmt.Println("error -- too many arguments!")
		fmt.Println("usage: ./init_html file opt")
		os.Exit(1)
	}

	if arg_num == 3 {
		if os.Args[2] == "-o" {
			opt = true
		}
		if !opt {
		  fmt.Println("error -- false option argument! ")
		  fmt.Println("usage: ./init_html file opt")
		  os.Exit(1)
		}
	}

	filnam := os.Args[1]
//	fmt.Println("from command line -- file name: ", filnam, " del: ", opt)

	filnam, err := init_html_file(filnam, opt)
	if err != nil {
		fmt.Println("error -- creating html skeleton file: ", err)
		os.Exit(1)
	}
	fmt.Println("success -- created html skeleton file! ", filnam)
}
