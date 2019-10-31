// +build !windows

package converter

import (
	"log"
	"os"
	"os/user"
	"strconv"
	"syscall"
)

// SetUserGroupMode Set destination file parameters
func SetUserGroupMode(file os.FileInfo, rawdata *Parameters) {
	if rawdata.User == "" {
		fileuser, _ := user.LookupId(strconv.Itoa(int(file.Sys().(*syscall.Stat_t).Uid)))
		log.Printf("user not provided, using '%s' as the original file", fileuser.Username)
		rawdata.User = fileuser.Username
	}
	if rawdata.Group == "" {
		filegroup, _ := user.LookupId(strconv.Itoa(int(file.Sys().(*syscall.Stat_t).Gid)))
		log.Printf("group not provided, using '%s' as the original file", filegroup.Username)
		rawdata.Group = filegroup.Username
	}
	if rawdata.Mode == 0 {
		filemode := file.Mode().Perm()
		log.Printf("mode not provided, using '%#o' as the original file", filemode)
		// Ignition requires decimal
		rawdata.Mode = int(filemode)
	}
}
