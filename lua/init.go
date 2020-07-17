package lua

import (
	"io/ioutil"
	"path"
	"runtime"
)

var ScriptContent string
var scriptName = "lock.lua"

func init() {
	curDir := getCurrentDir()
	luaFile := curDir + "/" + scriptName
	dat, err := ioutil.ReadFile(luaFile)
	if err != nil {
		return
	}
	ScriptContent = string(dat)
}

func getCurrentDir() string {
	_, filename, _, _ := runtime.Caller(1)
	return path.Dir(filename)
}
