package timber

import (
	"fmt"
	"testing"
)

func TestWorstPatternFormat(t *testing.T) {
	pf := NewPatFormatter("short:[%d %t] good:[%D %T] levelPadded:[%-10L] long:%S short:%s xs:%10x Msg:%M")
	var now int64 = 1319150347383485000
	lr := LogRecord{
		Level:      INFO,
		Timestamp:  now,
		SourceFile: "/blah/der/some_file.go",
		SourceLine: 7,
		Message:    "hellooooo nurse!"}
	msg := pf.Format(lr)
	pass := "short:[2011/10/20 15:39:07] good:[2011-10-20 15:39:07.383] levelPadded:[INFO      ] long:/blah/der/some_file.go:7 short:some_file.go:7 xs: some_file Msg:hellooooo nurse!\n"
	if msg != pass {
		t.Errorf("Expect:%s != \nResult:%s", pass, msg)
	}
}
func TestRealPatternFormat(t *testing.T) {
	pf := NewPatFormatter("[%D %T] [%L] %-10x %M")
	var now int64 = 1319150347383485000
	lr := LogRecord{
		Level:      INFO,
		Timestamp:  now,
		SourceFile: "/blah/der/some_file.go",
		SourceLine: 7,
		Message:    "hellooooo nurse!"}
	msg := pf.Format(lr)
	pass := "[2011-10-20 15:39:07.383] [INFO] some_file  hellooooo nurse!\n"
	if msg != pass {
		t.Errorf("Expect:%s != \nResult:%s", pass, msg)
	}
}

func BenchmarkWorstPatternFormat(b *testing.B) {
	pf := NewPatFormatter("short:[%d %t] good:[%D %T] levelPadded:[%-10L] long:%S short:%s xs:%10x Msg:%M")
	var now int64 = 1319150347383485000
	lr := LogRecord{
		Level:      INFO,
		Timestamp:  now,
		SourceFile: "/blah/der/some_file.go",
		SourceLine: 7,
		Message:    "hellooooo nurse!"}
	for i := 0; i < b.N; i++ {
		pf.Format(lr)
	}
}
func BenchmarkWorstJustSprintf(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fmt.Sprintf("short:[%d/%02d/%02d %02d:%02d:%02d] good:[%d-%02d-%02d %02d:%02d:%02d.%03d] levelPadded:[%-10s] long:%s short:%s xs:%10s Msg:%s\n", 2011, 10, 20, 15, 39, 7, 2011, 10, 20, 15, 39, 7, 383, "INFO", "/blah/der/some_file.go:7", "some_file.go:7", "some_file", "hellooooo nurse!")
	}
}
func BenchmarkRealPatternFormat(b *testing.B) {
	pf := NewPatFormatter("[%D %T] [%L] %-10x %M")
	var now int64 = 1319150347383485000
	lr := LogRecord{
		Level:      INFO,
		Timestamp:  now,
		SourceFile: "/blah/der/some_file.go",
		SourceLine: 7,
		Message:    "hellooooo nurse!"}
	for i := 0; i < b.N; i++ {
		pf.Format(lr)
	}
}
func BenchmarkReallJustSprintf(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fmt.Sprintf("[%d-%02d-%02d %02d:%02d:%02d.%03d] [%s] %-10s %s\n", 2011, 10, 20, 15, 39, 7, 383, "INFO", "some_file", "hellooooo nurse!")
	}
}
