package fsutil

import (
	"os"
	"sync"
)

type LockedFile struct {
	*os.File
	locker sync.Locker
}

func (f *LockedFile) Close() error {
	f.locker.Unlock()
	return f.File.Close()
}

type LockedFS struct {
	lockerMap *sync.Map
}

func NewLockedFS() *LockedFS {
	return &LockedFS{new(sync.Map)}
}

func (fs *LockedFS) getLocker(name string) *sync.RWMutex {
	locker, _ := fs.lockerMap.LoadOrStore(name, new(sync.RWMutex))
	return locker.(*sync.RWMutex)
}

func (fs *LockedFS) Create(name string) (*LockedFile, error) {
	locker := fs.getLocker(name)
	locker.Lock()

	file, err := os.Create(name)
	if err != nil {
		locker.Unlock()
		return nil, err
	}
	return &LockedFile{file, locker}, nil
}

func (fs *LockedFS) Open(name string) (*LockedFile, error) {
	locker := fs.getLocker(name)
	locker.RLock()

	file, err := os.Open(name)
	if err != nil {
		locker.RUnlock()
		return nil, err
	}
	return &LockedFile{file, locker.RLocker()}, nil
}
