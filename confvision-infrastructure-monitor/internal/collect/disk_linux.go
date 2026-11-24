//go:build linux

package collect

import "syscall"

func init() {
	DiskSpace = func(path string) (uint64, uint64, error) {
		var st syscall.Statfs_t
		if err := syscall.Statfs(path, &st); err != nil {
			return 0, 0, err
		}
		bsize := uint64(st.Bsize)
		return st.Blocks * bsize, st.Bavail * bsize, nil
	}
}
