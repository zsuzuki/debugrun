package exec

import (
	"bytes"
	"io"
	"os/exec"
	"regexp"
	"strings"
)

var mangledSymbolPattern = regexp.MustCompile(`_Z[A-Za-z0-9_$]+`)

type demangleWriter struct {
	dst        io.Writer
	cxxfilt    string
	lineBuffer []byte
}

func wrapDemangler(dst io.Writer) io.Writer {
	if dst == nil {
		return nil
	}
	cxxfilt, err := exec.LookPath("c++filt")
	if err != nil {
		return dst
	}
	return &demangleWriter{dst: dst, cxxfilt: cxxfilt}
}

func (w *demangleWriter) Write(p []byte) (int, error) {
	w.lineBuffer = append(w.lineBuffer, p...)

	for {
		idx := bytes.IndexByte(w.lineBuffer, '\n')
		if idx < 0 {
			return len(p), nil
		}

		line := string(w.lineBuffer[:idx+1])
		w.lineBuffer = w.lineBuffer[idx+1:]
		if _, err := io.WriteString(w.dst, demangleLine(w.cxxfilt, line)); err != nil {
			return 0, err
		}
	}
}

func (w *demangleWriter) Flush() error {
	if len(w.lineBuffer) == 0 {
		return nil
	}
	line := string(w.lineBuffer)
	w.lineBuffer = nil
	_, err := io.WriteString(w.dst, demangleLine(w.cxxfilt, line))
	return err
}

func flushDemangler(w io.Writer) error {
	demangler, ok := w.(*demangleWriter)
	if !ok {
		return nil
	}
	return demangler.Flush()
}

func flushDemanglers(err error, writers ...io.Writer) error {
	for _, writer := range writers {
		flushErr := flushDemangler(writer)
		if flushErr != nil && err == nil {
			err = flushErr
		}
	}
	return err
}

func demangleLine(cxxfilt, line string) string {
	matches := mangledSymbolPattern.FindAllString(line, -1)
	if len(matches) == 0 {
		return line
	}

	unique := make([]string, 0, len(matches))
	seen := map[string]bool{}
	for _, match := range matches {
		if seen[match] {
			continue
		}
		seen[match] = true
		unique = append(unique, match)
	}

	cmd := exec.Command(cxxfilt, unique...)
	out, err := cmd.Output()
	if err != nil {
		return line
	}

	replacements := strings.Split(strings.TrimSuffix(string(out), "\n"), "\n")
	if len(replacements) != len(unique) {
		return line
	}

	demangled := line
	for i, original := range unique {
		demangled = strings.ReplaceAll(demangled, original, replacements[i])
	}
	return demangled
}
