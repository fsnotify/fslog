/*
Experiment with Kernel Queues for File System Monitoring
(BSD, OS X, iOS)

"The thread-safety of kqueues is not documented,
so do not use the same kqueue in multiple threads
simultaneously." - Mark Dalrymple

*/

package main

import (
	"log"
	"os"
	"syscall"
)

const (
	NOTE_ALL_EVENTS = syscall.NOTE_DELETE | syscall.NOTE_WRITE |
		syscall.NOTE_EXTEND | syscall.NOTE_ATTRIB | syscall.NOTE_LINK |
		syscall.NOTE_RENAME | syscall.NOTE_REVOKE
)

var (
	noteDescription = map[uint32]string{
		syscall.NOTE_DELETE: "Delete",
		syscall.NOTE_WRITE:  "Write",
		syscall.NOTE_EXTEND: "Extend",
		syscall.NOTE_ATTRIB: "Attrib",
		syscall.NOTE_LINK:   "Link",
		syscall.NOTE_RENAME: "Rename",
		syscall.NOTE_REVOKE: "Revoke",
	}
)

func main() {
	kq, err := Kqueue()
	if err != nil {
		log.Fatal(err)
	}
	defer syscall.Close(kq)

	dirname := "/tmp"

	// syscall.O_NONBLOCK|syscall.O_RDONLY
	fd, err := syscall.Open(dirname, syscall.O_EVTONLY, 0700)
	// could use os.OpenFile with File.Fd(), but some of these flags don't exist in os. Hm.
	// should I be checking fd == -1 instead?
	if err != nil {
		log.Fatal(os.NewSyscallError("Open", err))
	}

	Add(kq, fd, NOTE_ALL_EVENTS)

	eventBuffer := make([]syscall.Kevent_t, 10)

	for {
		events, err := Read(kq, eventBuffer)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("%d events", len(events))
		for _, event := range events {
			logEvent(event)
		}
	}
}

func logEvent(event syscall.Kevent_t) {
	note := ""
	for bit, description := range noteDescription {
		if event.Fflags&bit == bit {
			note += description + " "
		}
	}
	log.Printf("fd: %d fflags: %s%+v", event.Ident, note, event)
}

// Kqueue creates a new kernel event queue and returns
// a descriptor
func Kqueue() (kq int, err error) {
	kq, err = syscall.Kqueue()
	return kq, os.NewSyscallError("Kqueue", err)
}

// Kevent registers events with the queue
func Add(kq, fd, fflags int) error {
	// TODO: multiple descriptors at once?
	changes := make([]syscall.Kevent_t, 1)

	const flags = syscall.EV_ADD | syscall.EV_CLEAR | syscall.EV_ENABLE

	// SetKevent converts ints to the platform-specific types:
	syscall.SetKevent(&changes[0], fd, syscall.EVFILT_VNODE, flags)
	changes[0].Fflags = uint32(fflags)

	// Udata could be useful for storing the file path with the event
	// but passing strings to C and back while avoiding GC issues may
	// not be worth it. Udata is usually *byte, but it a intptr_t on NetBSD.

	// register the event
	_, err := syscall.Kevent(kq, changes, nil, nil)
	// should I be checking success == -1?
	return os.NewSyscallError("Kevent", err)
}

// Read retrieves pending events
func Read(kq int, events []syscall.Kevent_t) ([]syscall.Kevent_t, error) {
	// block here until an event arrives
	n, err := syscall.Kevent(kq, nil, events, nil)
	return events[0:n], err
}
