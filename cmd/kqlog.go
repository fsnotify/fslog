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
	"time"
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

	fd := open("/tmp")
	defer syscall.Close(fd)
	log.Printf("fd: %d for %s", fd, "/tmp")

	fd2 := open("./kqlog.go")
	defer syscall.Close(fd2)
	log.Printf("fd: %d for %s", fd2, "./kqlog.go")

	if err := Add(kq, []int{fd, fd2}, NOTE_ALL_EVENTS); err != nil {
		log.Fatal(err)
	}

	eventBuffer := make([]syscall.Kevent_t, 10)
	// timespec := DurationToTimespec(100 * time.Millisecond)

	for {
		events, err := Read(kq, eventBuffer, nil)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("%d events", len(events))
		for _, event := range events {
			logEvent(event)
		}
	}
}

func open(name string) (fd int) {
	// syscall.O_NONBLOCK|syscall.O_RDONLY
	// why 0700?
	fd, err := syscall.Open(name, syscall.O_EVTONLY, 0700)
	// could use os.OpenFile with File.Fd(), but some of these flags don't exist in os. Hm.
	// also, what about directories?
	// should I be checking fd == -1 instead?
	if err != nil {
		log.Fatal(os.NewSyscallError("Open", err))
	}
	return fd
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

func DurationToTimespec(d time.Duration) syscall.Timespec {
	return syscall.NsecToTimespec(d.Nanoseconds())
}
