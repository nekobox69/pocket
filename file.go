// Package pocket Create at 2020-11-25 10:08
package pocket

import (
	"io/ioutil"
	"log"
	"path/filepath"
)

func ListAllFiles(path string) ([]string, error) {
	path = filepath.ToSlash(path)
	list := make([]string, 0)
	if "" == path {
		return list, nil
	}
	fileInfos, err := ioutil.ReadDir(path)
	if nil != err {
		log.Println(err.Error())
		return nil, err
	}
	for _, info := range fileInfos {
		if info.IsDir() {
			sub, err := ListAllFiles(path + "/" + info.Name())
			if nil != err {
				return []string{}, err
			}
			list = append(list, sub...)
		} else {
			list = append(list, path+"/"+info.Name())
		}
	}
	return list, nil
}
