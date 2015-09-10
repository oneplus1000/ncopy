package ncopycore

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"code.google.com/p/gcfg"
)

const PATH_FOLDER = ".ncopy"
const NCOPY_INI = "ncopy.ini"

var ErrDestDirNoEmpty = errors.New("Folder is not empty!")
var ErrSrcPathNotDir = errors.New("Source path is not folder!")

type NCopy struct {
	cfg          Conf
	ignoreRegexs []string
	projpath     string
}

func (n *NCopy) getDestPath() string {
	return n.projpath
}

func (n *NCopy) InitDestFolder(projpath string) error {
	isEmpty, err := n.IsDirEmpty(projpath)
	if err != nil {
		return err
	} else if !isEmpty {
		return ErrDestDirNoEmpty
	}

	var paramInit ParamInit
	fmt.Printf("source folder: ")
	_, err = fmt.Scanln(&paramInit.SrcDir)
	if err != nil {
		return err
	}

	isDir, err := n.IsDir(paramInit.SrcDir)
	if err != nil {
		return err
	} else if !isDir {
		return ErrSrcPathNotDir
	}

	err = n.createDotNCopyFolder(projpath)
	if err != nil {
		return err
	}

	err = n.createDefaultNCopyIni(projpath, paramInit)
	if err != nil {
		return err
	}

	return nil
}

func (n *NCopy) createDotNCopyFolder(projpath string) error {
	err := os.Mkdir(filepath.Join(projpath, PATH_FOLDER), 0777)
	if err != nil {
		return err
	}
	return nil
}

var TmplNCopyIni = "[src]\n" +
	"path = \"{{.SrcDir}}\"\n" +
	"[ignore]\n" +
	"files = \".git\"\n" +
	"files = \".gitignore\"\n"

func (n *NCopy) createDefaultNCopyIni(projpath string, paramInit ParamInit) error {

	tmpl, err := template.New("ncopy_ini").Parse(TmplNCopyIni)
	if err != nil {
		return nil
	}

	var buff bytes.Buffer
	err = tmpl.Execute(&buff, paramInit)
	if err != nil {
		return err
	}

	ncopyinipath := filepath.Join(projpath, PATH_FOLDER, NCOPY_INI)
	err = ioutil.WriteFile(ncopyinipath, buff.Bytes(), 0777)
	if err != nil {
		return err
	}

	return nil
}

func (n *NCopy) Copy(projpath string) error {

	n.projpath = projpath

	ncopyini := filepath.Join(projpath, PATH_FOLDER, NCOPY_INI)
	if _, err := os.Stat(ncopyini); os.IsNotExist(err) {
		return err
	}

	var cfg Conf
	err := gcfg.ReadFileInto(&cfg, ncopyini)
	if err != nil {
		return err
	}
	n.cfg = cfg
	n.ignoreToRegexs()
	//fmt.Printf("%#v\n", me.ignoreRegexs)

	if _, err := os.Stat(cfg.Src.Path); os.IsNotExist(err) {
		return err
	}

	n.copyfiles(cfg.Src.Path)

	return nil
}

func (n *NCopy) copyfiles(path string) {

	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer f.Close()

	isDir, err := n.isDir(f)
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
			if !n.isIgnore(n.virPath(fpath)) {
				n.copyfiles(fpath)
			}
		}
	} else {
		if !n.isIgnore(n.virPath(path)) {
			//TODO COPY
			err := n.copyfile(path, filepath.Join(n.getDestPath(), n.virPath(path)))
			if err != nil {
				log.Fatal(err)
				return
			}
		}
	}
}

func (n *NCopy) exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (n *NCopy) copyfile(source string, dest string) (err error) {

	fmt.Printf("copy %s\n", n.virPath(source))

	dirpath, _ := filepath.Split(dest)
	ex, err := n.exists(dirpath)
	if err != nil {
		return err
	}

	if !ex {
		err = os.MkdirAll(dirpath, 0777)
		if err != nil {
			return err
		}
	}

	err = n.cp(source, dest)
	if err != nil {
		return err
	}

	return nil
}

func (n *NCopy) ignoreToRegexs() {
	for _, s := range n.cfg.Ignore.Files {
		rx := strings.Replace(s, "\\", "\\\\", -1)
		rx = strings.Replace(s, "/", "\\/", -1)
		rx = strings.Replace(rx, ".", "\\.", -1)
		rx = strings.Replace(rx, "~", "\\~", -1)
		rx = "^" + rx + "$"
		n.ignoreRegexs = append(n.ignoreRegexs, rx)
	}
}

func (n *NCopy) isIgnore(virpath string) bool {

	for _, rx := range n.ignoreRegexs {
		matched, err := regexp.MatchString(rx, virpath)
		if err != nil {
			log.Fatal(err)
			return false
		}

		if matched {
			//fmt.Printf("%s\n", virpath)
			return true //ignore
		}
	}
	return false
}

func (n *NCopy) cp(src string, dst string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	// no need to check errors on read only file, we already got everything
	// we need from the filesystem, so nothing can go wrong now.
	defer s.Close()
	d, err := os.Create(dst)
	if err != nil {
		return err
	}

	if _, err := io.Copy(d, s); err != nil {
		d.Close()
		return err
	}
	return d.Close()
}

func (n *NCopy) virPath(path string) string {
	l := len(n.cfg.Src.Path) + 1
	vpath := path[l:]
	return vpath
}

func (n *NCopy) IsDir(path string) (bool, error) {

	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()
	return n.isDir(f)
}

func (n *NCopy) isDir(file *os.File) (bool, error) {

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

func (n *NCopy) IsDirEmpty(dir string) (bool, error) {

	f, err := os.Open(dir)
	if err != nil {
		return false, err
	}
	defer f.Close()
	_, err = f.Readdir(1) //อ่านแค่ file เดียว
	if err == io.EOF {    //เจอ EOF เลย ว่างแน่ๆ
		return true, nil
	}
	return false, err
}
