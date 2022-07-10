package priv

import (
	"fmt"
	"os/user"
	"strconv"
	"syscall"
)

func Drop(nobodyUsername string) (bool, error) {

	uid := syscall.Getuid()

	if uid == 0 {

		nobody, err := user.Lookup(nobodyUsername)
		if err != nil {
			return false, fmt.Errorf("failed to look up user %s: %v", nobodyUsername, err)
		}

		nobodyGid, err := strconv.Atoi(nobody.Gid)
		if err != nil {
			return false, fmt.Errorf("failed to convert gid %s to int: %v", nobody.Gid, err)
		}
		nobodyUid, err := strconv.Atoi(nobody.Uid)
		if err != nil {
			return false, fmt.Errorf("failed to convert uid %s to int: %v", nobodyUsername, err)
		}
		if err := syscall.Setgid(nobodyGid); err != nil {
			return false, fmt.Errorf("failed to setgid: %v", err)
		}

		if err := syscall.Setuid(nobodyUid); err != nil {
			return false, fmt.Errorf("failed to setuid: %v", err)
		}

		return true, nil

	}

	return false, nil

}
