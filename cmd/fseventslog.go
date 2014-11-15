/*
Experiment with FSEvents for File System Monitoring
(BSD, OS X, iOS)
*/

// +build darwin

package main

import (
	"bufio"
	"log"
	"os"
	"time"

	"github.com/go-fsnotify/fsevents"
)

func main() {
	stream := fsevents.New(0, fsevents.NOW, time.Second/10, fsevents.CF_FILEEVENTS, "/tmp")
	stream.Start()

	go func() {
		for msg := range stream.Chan {
			for _, event := range msg {
				logEvent(event)
			}
		}
	}()

	// press enter to continue
	in := bufio.NewReader(os.Stdin)
	in.ReadString('\n')
	stream.Close()
}

var noteDescription = map[fsevents.EventFlags]string{
	fsevents.EF_MUSTSCANSUBDIRS: "MustScanSubdirs",
	fsevents.EF_USERDROPPED:     "UserDropped",
	fsevents.EF_KERNELDROPPED:   "KernelDropped",
	fsevents.EF_EVENTIDSWRAPPED: "EventIdsWrapped",
	fsevents.EF_HISTORYDONE:     "HistoryDone",
	fsevents.EF_ROOTCHANGED:     "RootChanged",
	fsevents.EF_MOUNT:           "Mount",
	fsevents.EF_UNMOUNT:         "Unmount",

	fsevents.EF_CREATED:       "Created",
	fsevents.EF_REMOVED:       "Removed",
	fsevents.EF_INODEMETAMOD:  "InodeMetaMod",
	fsevents.EF_RENAMED:       "Renamed",
	fsevents.EF_MODIFIED:      "Modified",
	fsevents.EF_FINDERINFOMOD: "FinderInfoMod",
	fsevents.EF_CHANGEOWNER:   "ChangeOwner",
	fsevents.EF_XATTRMOD:      "XAttrMod",
	fsevents.EF_ISFILE:        "IsFile",
	fsevents.EF_ISDIR:         "IsDir",
	fsevents.EF_ISSYMLINK:     "IsSymLink",
}

func logEvent(event fsevents.Event) {
	note := ""
	for bit, description := range noteDescription {
		if event.Flags&bit == bit {
			note += description + " "
		}
	}
	log.Printf("EventID: %d Path: %s Flags: %s", event.Id, event.Path, note)
}
