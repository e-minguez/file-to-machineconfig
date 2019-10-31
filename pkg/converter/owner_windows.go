// +build windows

package converter

import (
	"log"
	"os"
	"runtime"
)

var defaultMode = 0644
var defaultUsername = "root"
var defaultGroupname = "root"

// SetUserGroupMode Set destination file parameters
func SetUserGroupMode(file os.FileInfo, rawdata *Parameters) {
	if rawdata.User == "" && runtime.GOOS == "windows" {
		log.Printf("user not provided, using '%s' as default", defaultUsername)
		rawdata.User = defaultUsername
	}
	defaultGroupname := "root"
	if rawdata.Group == "" && runtime.GOOS == "windows" {
		log.Printf("group not provided, using '%s' as default", defaultGroupname)
		rawdata.Group = defaultGroupname
	}
	if rawdata.Mode == 0 && runtime.GOOS == "windows" {
		log.Printf("mode not provided, using '%#o' as default", defaultMode)
		rawdata.Mode = int(defaultMode)
	}
}
