package timber

import (
	"fmt"
	"testing"
)

var lr = &LogRecord{
	Level:       INFO,
	Timestamp:   1319150347383485000,
	SourceFile:  "/blah/der/some_file.go",
	SourceLine:  7,
	Message:     "hellooooo nurse!",
	FuncPath:    "hi.Zoot",
	PackagePath: "hi",
}

var optiontests = []struct {
	in  string
	out string
}{
	{"%T", "15:39:07.383\n"},
	{"%t", "15:39:07\n"},
	{"%D", "2011-10-20\n"},
	{"%d", "2011/10/20\n"},
	{"%-10L", "INFO      \n"},
	{"%S", "/blah/der/some_file.go:7\n"},
	{"%s", "some_file.go:7\n"},
	{"%x", "some_file\n"},
	{"%M", "hellooooo nurse!\n"},
	//{"%%", "%\n"}, // TODO fix
	{"%P", "hi.Zoot\n"},
	{"%p", "hi\n"},
}

func verify(t *testing.T, input, output, expected string) {
	if output != expected {
		t.Errorf("%s: output %s != %s", input, output, expected)
	}
}

func TestOptions(t *testing.T) {
	for _, tt := range optiontests {
		pf := NewPatFormatter(tt.in)
		verify(t, tt.in, pf.Format(lr), tt.out)
	}
}

func TestWorstPatternFormat(t *testing.T) {
	in := "short:[%d %t] good:[%D %T] levelPadded:[%-10L] long:%S short:%s xs:%10x Msg:%M Fnc:%P Pkg:%p"
	out := "short:[2011/10/20 15:39:07] good:[2011-10-20 15:39:07.383] levelPadded:[INFO      ] " +
		"long:/blah/der/some_file.go:7 short:some_file.go:7 xs: some_file Msg:hellooooo nurse! Fnc:hi.Zoot Pkg:hi\n"
	pf := NewPatFormatter(in)
	verify(t, in, pf.Format(lr), out)
}

func TestRealPatternFormat(t *testing.T) {
	in := "[%D %T] [%L] %-10x %M"
	out := "[2011-10-20 15:39:07.383] [INFO] some_file  hellooooo nurse!\n"
	pf := NewPatFormatter(in)
	verify(t, in, pf.Format(lr), out)
}

func BenchmarkWorstPatternFormat(b *testing.B) {
	pf := NewPatFormatter("short:[%d %t] good:[%D %T] levelPadded:[%-10L] long:%S short:%s xs:%10x Msg:%M Fnc:%P Pkg:%p")
	for i := 0; i < b.N; i++ {
		pf.Format(lr)
	}
}

func BenchmarkWorstJustSprintf(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fmt.Sprintf("short:[%d/%02d/%02d %02d:%02d:%02d] good:[%d-%02d-%02d %02d:%02d:%02d.%03d] "+
			"levelPadded:[%-10s] long:%s short:%s xs:%10s Msg:%s Fnc:%s Pkg:%s\n", 2011, 10, 20, 15, 39, 7,
			2011, 10, 20, 15, 39, 7, 383, "INFO", "/blah/der/some_file.go:7", "some_file.go:7", "some_file", "hellooooo nurse!", "hi.Zoot", "hi")
	}
}

func BenchmarkRealPatternFormat(b *testing.B) {
	pf := NewPatFormatter("[%D %T] [%L] %-10x %M")
	for i := 0; i < b.N; i++ {
		pf.Format(lr)
	}
}

func BenchmarkReallJustSprintf(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fmt.Sprintf("[%d-%02d-%02d %02d:%02d:%02d.%03d] [%s] %-10s %s\n", 2011, 10, 20, 15, 39, 7, 383, "INFO", "some_file", "hellooooo nurse!")
	}
}
