package lua

import (
	"io/ioutil"
	"path"
	"runtime"
)

var ScriptContent string

// Lua脚本文件名
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

// getCurrentDir
// 获取当前的目录的路径
func getCurrentDir() string {
	_, filename, _, _ := runtime.Caller(1)
	return path.Dir(filename)
}
