package ncopycore

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"code.google.com/p/gcfg"
)

const PATH_FOLDER = ".ncopy"
const NCOPY_INI = "ncopy.ini"

type NCopy struct {
	cfg          Conf
	ignoreRegexs []string
}

func (me *NCopy) Copy(projpath string) error {

	ncopyini := filepath.Join(projpath, PATH_FOLDER, NCOPY_INI)
	if _, err := os.Stat(ncopyini); os.IsNotExist(err) {
		return err
	}

	var cfg Conf
	err := gcfg.ReadFileInto(&cfg, ncopyini)
	if err != nil {
		return err
	}
	me.cfg = cfg
	me.ignoreToRegexs()
	fmt.Printf("%#v\n", me.ignoreRegexs)

	if _, err := os.Stat(cfg.Src.Path); os.IsNotExist(err) {
		return err
	}

	me.copyfiles(cfg.Src.Path)

	return nil
}

func (me *NCopy) copyfiles(path string) {

	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer f.Close()

	isDir, err := me.isDir(f)
	if err != nil {
		log.Fatal(err)
		return
	}

	if isDir {
		finfos, err := ioutil.ReadDir(path)
		if err != nil {
			log.Fatal(err)
			return
		}
		for _, finfo := range finfos {
			fpath := filepath.Join(path, finfo.Name())
			if !me.isIgnore(me.virPath(fpath)) {
				me.copyfiles(fpath)
			}
		}
	} else {
		if !me.isIgnore(me.virPath(path)) {
			//TODO COPY
			//fmt.Printf("%s\n", me.virPath(path))
		}
	}
}

func (me *NCopy) ignoreToRegexs() {
	for _, s := range me.cfg.Ignore.Files {
		rx := strings.Replace(s, "\\", "\\\\", -1)
		rx = strings.Replace(s, "/", "\\/", -1)
		rx = strings.Replace(rx, ".", "\\.", -1)
		rx = strings.Replace(rx, "~", "\\~", -1)
		rx = "^" + rx + "$"
		me.ignoreRegexs = append(me.ignoreRegexs, rx)
	}
}

func (me *NCopy) isIgnore(virpath string) bool {

	for _, rx := range me.ignoreRegexs {
		matched, err := regexp.MatchString(rx, virpath)
		if err != nil {
			log.Fatal(err)
			return false
		}

		if matched {
			fmt.Printf("%s\n", virpath)
			return true //ignore
		}
	}
	return false
}

func (me *NCopy) virPath(path string) string {
	l := len(me.cfg.Src.Path) + 1
	vpath := path[l:]
	return vpath
}

func (me *NCopy) isDir(file *os.File) (bool, error) {

	finfo, err := file.Stat()
	if err != nil {
		return false, err
	}
	mode := finfo.Mode()
	if mode.IsDir() { //dir
		return true, nil
	}
	return false, nil

}
