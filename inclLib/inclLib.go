// inclLib.go
// library to parse html source file and merge content into dest file
// author: prr
// date: 23/4/2022
// copyright 2022 prr, azul software
//
package inclLib

import (
	"fmt"
	"os"
)

const BlockSize=4096

func ListInclFiles(srcfil *os.File) (inclFilNames *[]string, err error) {
	var files []string
	var inBuf []byte
	var i int64

	inBuf = make([]byte, BlockSize)
	nb, err := srcfil.Stat()
	if err != nil {
		return nil, fmt.Errorf("srcfil.Stat: %v", err)
	}

	blocks := nb.Size()/BlockSize
	blockRem := nb.Size() - blocks*BlockSize

	for i=0; i< blocks; i++ {

	}

//	os.Read
	return &files, nil
}
