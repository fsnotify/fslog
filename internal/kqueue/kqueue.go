// +build freebsd openbsd netbsd darwin

// Package kqueue provides access to Kernel Queues for File System
// Monitoring.
package kqueue

import (
	"os"
	"syscall"
)

const (
	NOTE_ALL_EVENTS = syscall.NOTE_DELETE | syscall.NOTE_WRITE |
		syscall.NOTE_EXTEND | syscall.NOTE_ATTRIB | syscall.NOTE_LINK |
		syscall.NOTE_RENAME | syscall.NOTE_REVOKE
)

// Kqueue creates a new kernel event queue and returns
// a descriptor
func Kqueue() (kq int, err error) {
	kq, err = syscall.Kqueue()
	return kq, os.NewSyscallError("Kqueue", err)
}

// Kevent registers events with the queue
func Add(kq int, fds []int, fflags int) error {
	changes := make([]syscall.Kevent_t, len(fds))

	const flags = syscall.EV_ADD | syscall.EV_CLEAR | syscall.EV_ENABLE

	// SetKevent converts ints to the platform-specific types:
	for i, fd := range fds {
		syscall.SetKevent(&changes[i], fd, syscall.EVFILT_VNODE, flags)
		changes[i].Fflags = uint32(fflags)
		// Udata could be useful for storing the file path with the event
		// but passing strings to C and back while avoiding GC issues may
		// not be worth it.
		// Udata is usually *byte, but it is a intptr_t on NetBSD.
	}

	// register the events
	_, err := syscall.Kevent(kq, changes, nil, nil)
	// should I be checking success == -1?
	return os.NewSyscallError("Kevent", err)
}

// Read retrieves pending events
// A timeout of nil blocks indefinitely, while 0 polls the queue.
func Read(kq int, events []syscall.Kevent_t, timeout *syscall.Timespec) ([]syscall.Kevent_t, error) {
	n, err := syscall.Kevent(kq, nil, events, timeout)
	if err != nil {
		return nil, os.NewSyscallError("Kevent", err)
	}
	return events[0:n], nil
}
