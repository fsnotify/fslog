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

func main() {
	log.Println("Hello")

	kq, err := kqueue()
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
	changes[0].Fflags = syscall.NOTE_WRITE // uint32

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
		if err != nil {
			log.Fatal(err)
		}
		log.Println(n, events)
	}

	syscall.Close(kq)
}

// kqueue creates a new kernel event queue and returns
// a descriptor
func kqueue() (fd int, err error) {
	fd, err = syscall.Kqueue()
	return fd, os.NewSyscallError("Kqueue", err)
}
