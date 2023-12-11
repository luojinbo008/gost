package gxlog

import (
	"github.com/davecgh/go-spew/spew"

	"github.com/k0kubun/pp"
)

func PrettyString(i interface{}) string {
	return spew.Sdump(i)
}

func ColorSprint(i interface{}) string {
	return pp.Sprint(i)
}

func ColorSprintln(i interface{}) string {
	return pp.Sprintln(i)
}

func ColorSprintf(fmt string, args ...interface{}) string {
	return pp.Sprintf(fmt, args...)
}

func ColorPrint(i interface{}) {
	pp.Print(i)
}

func ColorPrintln(i interface{}) {
	pp.Println(i)
}

func ColorPrintf(fmt string, args ...interface{}) {
	pp.Printf(fmt, args...)
}