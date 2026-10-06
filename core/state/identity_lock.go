package state

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/sys/unix"
)

// The lock inode is retained: unlinking it would permit concurrent holders on
// different inodes. Separate Store instances and processes share this lock.
func (s *Store) lockIdentityState() (func(), error) {
	f, err := os.OpenFile(filepath.Join(s.stateDir, ".identity.lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	deadline := time.Now().Add(5 * time.Second)
	for {
		err = unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB)
		if err == nil {
			return func() { _ = unix.Flock(int(f.Fd()), unix.LOCK_UN); _ = f.Close() }, nil
		}
		if !errors.Is(err, unix.EWOULDBLOCK) && !errors.Is(err, unix.EINTR) {
			_ = f.Close()
			return nil, err
		}
		if time.Now().After(deadline) {
			_ = f.Close()
			return nil, fmt.Errorf("state: identity lock timeout: %w", err)
		}
		time.Sleep(10 * time.Millisecond)
	}
}
