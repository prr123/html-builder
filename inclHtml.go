// inclHtml.go
// program that parses an html file for include statements and merges the include file into source file
// usage: inclHtml source dest opt
//
// author: prr
// date: 23/4/2022
// copyright 2022 prr, azul software
//
package main

import (
	"os"
	"fmt"
	incl "utility/incl/inclLib"
)

func main() {

	if len(os.Args) < 3 {
		fmt.Printf("insufficient command arguments!\n")
		fmt.Printf("usage is: \"./incHtml src dest \" \n")
		os.Exit(1)
	}

	srcFilnam := os.Args[1]
	_, err := os.Stat(srcFilnam)
	if os.IsNotExist(err) {
		fmt.Printf("source file: %s does not exist!\n", srcFilnam)
		os.Exit(1)
	}

	dstFilnam := os.Args[2]
	_, err = os.Stat(dstFilnam)
	if !os.IsNotExist(err) {
		fmt.Printf("destination file: %s does exist!\n", dstFilnam)
		fmt.Printf("deleting old destination file!\n")
		err1:= os.Remove(dstFilnam)
        if err1 != nil {
            fmt.Printf("os.Remove: cannot remove html file: %s! error: %v", dstFilnam, err1)
			os.Exit(1)
        }
	}

	srcfil, err := os.Open("srcFilnam")
	if err != nil {
		fmt.Printf("cannot open source file: %s!\n", srcFilnam)
		os.Exit(1)
	}

	dstfil, err := os.Create("dstFilnam")
	if err != nil {
		fmt.Printf("cannot open destination file %d!\n", dstFilnam)
		os.Exit(1)
	}

	files, err := incl.ListInclFiles(srcfil)

	infil.Close()
	dstfil.Close()
	fmt.Println("success!")
}
