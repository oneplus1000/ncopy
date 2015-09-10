package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/oneplus1000/ncopy/ncopycore"
)

var ncopyInit = flag.Bool("init", false, "init destination folder.")

func main() {
	flag.Parse()
	projpath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	var ncopy ncopycore.NCopy
	if *ncopyInit { //init
		err = ncopy.InitDestFolder(projpath)
		if err != nil {
			log.Fatal(err)
			return
		}
	} else { //copy
		err = ncopy.Copy(projpath)
		if err != nil {
			log.Fatal(err)
			return
		}
	}
}

/*
func main() {
	var str string
	fmt.Scanf("%s", &str)
	fmt.Printf("%s", str)
}*/
