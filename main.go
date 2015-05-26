package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/oneplus1000/ncopy/ncopycore"
)

func main() {
	projpath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Println(projpath)
	err = ncopycore.NCopy(projpath)
	if err != nil {
		log.Fatal(err)
	}
}
