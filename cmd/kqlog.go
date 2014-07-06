/*
Experiment with Kernel Queues for File System Monitoring
(BSD, OS X, iOS)

"The thread-safety of kqueues is not documented,
so do not use the same kqueue in multiple threads
simultaneously." - Mark Dalrymple
*/

// +build freebsd openbsd netbsd darwin

package main

import (
	"log"
	"os"
	"syscall"
	"time"

	"github.com/fsnotify/fslog/internal/kqueue"
)

func main() {
	kq, err := kqueue.Kqueue()
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

	if err := kqueue.Add(kq, []int{fd, fd2}, kqueue.NOTE_ALL_EVENTS); err != nil {
		log.Fatal(err)
	}

	eventBuffer := make([]syscall.Kevent_t, 10)
	timespec := durationToTimespec(100 * time.Millisecond)

	for {
		events, err := kqueue.Read(kq, eventBuffer, &timespec)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("%d events", len(events))
		for _, event := range events {
			logEvent(event)
		}
	}
}

var noteDescription = map[uint32]string{
	syscall.NOTE_DELETE: "Delete",
	syscall.NOTE_WRITE:  "Write",
	syscall.NOTE_EXTEND: "Extend",
	syscall.NOTE_ATTRIB: "Attrib",
	syscall.NOTE_LINK:   "Link",
	syscall.NOTE_RENAME: "Rename",
	syscall.NOTE_REVOKE: "Revoke",
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

func open(name string) (fd int) {
	// darwin: syscall.O_EVTONLY
	// otherwise: syscall.O_NONBLOCK|syscall.O_RDONLY
	// why 0700?
	fd, err := syscall.Open(name, syscall.O_EVTONLY, 0700)
	// should I be checking fd == -1 instead?
	if err != nil {
		log.Fatal(os.NewSyscallError("Open", err))
	}
	return fd
}

// durationToTimespec prepares a timeout value
func durationToTimespec(d time.Duration) syscall.Timespec {
	return syscall.NsecToTimespec(d.Nanoseconds())
}
