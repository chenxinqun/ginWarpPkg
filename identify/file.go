package identify

import "os"

func IsExists(filename string) (info os.FileInfo, ret bool) {
	var err error
	info, err = os.Stat(filename)
	ret = os.IsExist(err)
	return
}
