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
	log.Println("Hello")

	kq, err := Kqueue()
	if err != nil {
		log.Fatal(err)
	}

	dirname := "/tmp"

	// syscall.O_NONBLOCK|syscall.O_RDONLY
	fd, err := syscall.Open(dirname, syscall.O_EVTONLY, 0700)
	// could use os.OpenFile with File.Fd(), but some of these flags don't exist in os. Hm.
	// should I be checking fd == -1 instead?
	if err != nil {
		log.Fatal(os.NewSyscallError("Open", err))
	}

	changes := make([]syscall.Kevent_t, 1)

	flags := syscall.EV_ADD | syscall.EV_CLEAR | syscall.EV_ENABLE
	changes[0].Fflags = NOTE_ALL_EVENTS // uint32

	// Udata is usually *byte, but it a intptr_t on netbsd.
	// could use syscall.BytePtrFromString or some cgo
	// https://code.google.com/p/go-wiki/wiki/cgo
	// but having the string not GC too soon, and then having to
	// clean it up all sounds like a bad idea

	// SetKevent converts ints to the platform-specific types
	syscall.SetKevent(&changes[0], fd, syscall.EVFILT_VNODE, flags)
	log.Printf("%+v", changes)

	// register the event
	_, err = syscall.Kevent(kq, changes, nil, nil)
	if err != nil {
		log.Fatal(err)
	}

	for {
		events := make([]syscall.Kevent_t, 1)
		// block here until an event arrives
		n, err := syscall.Kevent(kq, nil, events, nil)
		log.Printf("%d events", n)
		if err != nil {
			log.Fatal(err)
		}
		for _, event := range events {
			logEvent(event)
		}
	}

	syscall.Close(kq)
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
func Kqueue() (fd int, err error) {
	fd, err = syscall.Kqueue()
	return fd, os.NewSyscallError("Kqueue", err)
}
