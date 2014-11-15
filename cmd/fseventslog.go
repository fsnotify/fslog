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
	stream := fsevents.New(0, fsevents.SinceNow, time.Second/10, fsevents.FileEvents, "/tmp")
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
	fsevents.MustScanSubDirs: "MustScanSubdirs",
	fsevents.UserDropped:     "UserDropped",
	fsevents.KernelDropped:   "KernelDropped",
	fsevents.EventIdsWrapped: "EventIdsWrapped",
	fsevents.HistoryDone:     "HistoryDone",
	fsevents.RootChanged:     "RootChanged",
	fsevents.Mount:           "Mount",
	fsevents.Unmount:         "Unmount",

	fsevents.Created:       "Created",
	fsevents.Removed:       "Removed",
	fsevents.InodeMetaMod:  "InodeMetaMod",
	fsevents.Renamed:       "Renamed",
	fsevents.Modified:      "Modified",
	fsevents.FinderInfoMod: "FinderInfoMod",
	fsevents.ChangeOwner:   "ChangeOwner",
	fsevents.XattrMod:      "XAttrMod",
	fsevents.IsFile:        "IsFile",
	fsevents.IsDir:         "IsDir",
	fsevents.IsSymlink:     "IsSymLink",
}

func logEvent(event fsevents.Event) {
	note := ""
	for bit, description := range noteDescription {
		if event.Flags&bit == bit {
			note += description + " "
		}
	}
	log.Printf("EventID: %d Path: %s Flags: %s", event.ID, event.Path, note)
}
