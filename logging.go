package main

import (
	"log"
	"os"
)

type logg struct {
	info *log.Logger
	dbg  *log.Logger
	warn *log.Logger
	err  *log.Logger
}

var (
	Log logg
)

func (l logg) Info(format string, s ...any) {
	if l.info != nil {
		l.info.Printf(format, s...)
	}
}

func (l logg) Debug(format string, s ...any) {
	if l.dbg != nil {
		l.dbg.Printf(format, s...)
	}
}

func (l logg) Warn(format string, s ...any) {
	if l.warn != nil {
		l.warn.Printf(format, s...)
	}
}

func (l logg) Error(format string, s ...any) {
	if l.err != nil {
		l.err.Printf(format, s...)
	}
}

func (l logg) Disable(name string) {
	switch name {
	case "Info":
		Log.info = nil
	case "Debug":
		Log.dbg = nil
	case "Warning":
		Log.warn = nil
	case "Error":
		Log.err = nil
	default:
		panic("Unknown option " + name)
	}
}

func init() {
	Log = logg{}
	Log.info = log.New(os.Stdout, "I: ", log.Lshortfile)
	Log.dbg = log.New(os.Stdout, "D: ", log.Lshortfile)
	Log.warn = log.New(os.Stdout, "W: ", log.Lshortfile)
	Log.err = log.New(os.Stderr, "E: ", log.Lshortfile)
}
