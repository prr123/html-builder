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

type InclFilTyp struct {
	Name string
	Exist bool
	Start int64
	End int64
}

func parseBlock(inBuf []byte, files []InclFilTyp)(err error) {

	return nil
}

func ListInclFiles(srcfil *os.File) (inclFilNames *[]InclFilTyp, err error) {
	var files []InclFilTyp
	var inBuf []byte
	var i int64

	inBuf = make([]byte, BlockSize)
	filinfo, err := srcfil.Stat()
	if err != nil {
		return nil, fmt.Errorf("srcfil.Stat: %v", err)
	}

	blocks := filinfo.Size()/BlockSize
	blockRem := filinfo.Size() - blocks*BlockSize

	for i=0; i< blocks; i++ {
		_, err = srcfil.Seek(i*BlockSize, 0)
		if err != nil {
			return nil, fmt.Errorf("srcfil.Seek: %v", err)
		}
		nbyt, err := srcfil.Read(inBuf)
		if err != nil {
			return nil, fmt.Errorf("srcfil.Read block %d: %v", i, err)
		}
		if nbyt != BlockSize {
			return nil, fmt.Errorf("srcfil.Read block %d num bytes %d < BlockSize: %v", i, nbyt, err)
		}
		err = parseBlock(inBuf, files)
		if err != nil {
			return nil, fmt.Errorf("parseBlock block %d: %v", i, err)
		}
	}

	nbyt, err := srcfil.Read(inBuf)
	if err != nil {
		return nil, fmt.Errorf("srcfil.Read last block: %v", err)
	}
	if nbyt != int(blockRem) {
		return nil, fmt.Errorf("srcfil.Read last block: num bytes %d < BlockSize: %v", nbyt, err)
	}

	err = parseBlock(inBuf, files)
	if err != nil {
		return nil, fmt.Errorf("parseBlock last block: %v", i, err)
	}


	return &files, nil
}
