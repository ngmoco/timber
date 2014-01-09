package timber

import (
	"fmt"
	"testing"
	"time"
)

var lr = &LogRecord{
	Level:       WARNING,
	Timestamp:   time.Unix(0, 1319230347383485000),
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
	{"%T", "21:52:27.383\n"},
	{"%t", "21:52:27\n"},
	{"%D", "2011-10-21\n"},
	{"%d", "2011/10/21\n"},
	{"%-10L", "WARN      \n"},
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
	out := "short:[2011/10/21 21:52:27] good:[2011-10-21 21:52:27.383] levelPadded:[WARN      ] " +
		"long:/blah/der/some_file.go:7 short:some_file.go:7 xs: some_file Msg:hellooooo nurse! Fnc:hi.Zoot Pkg:hi\n"
	pf := NewPatFormatter(in)
	verify(t, in, pf.Format(lr), out)
}

func TestRealPatternFormat(t *testing.T) {
	in := "[%D %T] [%L] %-10x %M"
	out := "[2011-10-21 21:52:27.383] [WARN] some_file  hellooooo nurse!\n"
	pf := NewPatFormatter(in)
	verify(t, in, pf.Format(lr), out)
}

func TestRealPatternFormatLong(t *testing.T) {
        in := "[%D %T] [%l] %-10x %M"
        out := "[2011-10-21 21:52:27.383] [WARNING] some_file  hellooooo nurse!\n"
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
			"levelPadded:[%-10s] long:%s short:%s xs:%10s Msg:%s Fnc:%s Pkg:%s\n", 2011, 10, 21, 23, 39, 7,
			2011, 10, 21, 23, 39, 7, 383, "WARN", "/blah/der/some_file.go:7", "some_file.go:7", "some_file", "hellooooo nurse!", "hi.Zoot", "hi")
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
		fmt.Sprintf("[%d-%02d-%02d %02d:%02d:%02d.%03d] [%s] %-10s %s\n", 2011, 10, 21, 23, 39, 7, 383, "WARN", "some_file", "hellooooo nurse!")
	}
}
