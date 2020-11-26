package lib

import (
	"fmt"
	"io/ioutil"
	"runtime"
	"ubc/tools"
)

func GetDbDir() string {

	sys_version := runtime.GOOS

	if sys_version == "windows" {
		isexist, _ := tools.PathExists("E:\\tmp")
		if isexist {
			dir, err := ioutil.TempDir("E:\\tmp", "abci-kvstore-test") // TODO
			if err != nil {
				fmt.Println(err)
			}
			return dir
		} else {
			panic("Db file not exist!")
		}

	}
	dir, err := ioutil.TempDir("/tmp", "abci-kvstore-test") // TODO
	if err != nil {
		fmt.Println(err)
	}
	return dir
}
