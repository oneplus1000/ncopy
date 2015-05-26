package ncopycore

import (
	"fmt"
	"os"
	"path/filepath"

	"code.google.com/p/gcfg"
)

const PATH_FOLDER = ".ncopy"
const NCOPY_INI = "ncopy.ini"

//const PATH_NCOPY_USER_CONF = "ncopy.user.ini"

func NCopy(projpath string) error {

	ncopyini := filepath.Join(projpath, PATH_FOLDER, NCOPY_INI)
	if _, err := os.Stat(ncopyini); os.IsNotExist(err) {
		return err
	}

	var cfg Conf
	err := gcfg.ReadFileInto(&cfg, ncopyini)
	if err != nil {
		return err
	}

	fmt.Printf("%#v", cfg)
	err = nCopy(cfg)
	if err != nil {
		return err
	}

	return nil
}

func nCopy(cfg Conf) error {
	if _, err := os.Stat(cfg.Src.Path); os.IsNotExist(err) {
		return err
	}
	return nil
}
