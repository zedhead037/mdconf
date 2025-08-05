package mdconf

import (
	"bufio"
	"errors"
	"io"
	"regexp"
	"strings"
)

// this is here because UnreadByte somehow can't unread a whole line.
type trueLineReader struct {
	reader *bufio.Reader
	buffer []string
}

func newTrueLineReader(r *bufio.Reader) *trueLineReader {
	return &trueLineReader{
		reader: r,
		buffer: make([]string, 0),
	}
}
func (r *trueLineReader) readLine() (string, error) {
	if len(r.buffer) > 0 {
		rb, e := r.buffer[:len(r.buffer)-1], r.buffer[len(r.buffer)-1]
		r.buffer = rb
		return e, nil
	}
	rr, err := r.reader.ReadString('\n')
	if err != nil { return "", err }
	return strings.TrimSuffix(rr, "\n"), nil
}
func (r *trueLineReader) unreadLine(l string) {
	r.buffer = append(r.buffer, l)
}

type MDConfSection struct {
	Level int8
	SectionName string
	ValueMap map[string]string
	Subsection []*MDConfSection
}

var whitespaceStr = " \t\r\n\v\b\f"
func isWhite(c byte) bool {
	return c == ' ' || c == '\t' || c == '\r' || c == '\n' || c == '\v' || c == '\b' || c == '\f'
}

var reEscape = regexp.MustCompile(`\\(.)`)
func unescape(s string) string {
	return reEscape.ReplaceAllString(s, "$1")
}

var reSectionHeader = regexp.MustCompile(`^\s*(#+)\s*((?:\\.|.)*)\s*$`)
var reKV = regexp.MustCompile(`^\s*\+\s*((?:\\.|[^:\s])*)\s*:\s*((?:\\.|.)*?)$`)

func parseInner(currentSection *MDConfSection, br *trueLineReader) *MDConfSection {
	currentMultiLineValueLineStack := make([]string, 0)
	currentMultiLineKey := ""
	currentlyReadingMultiLine := false

	for {
		r, err := br.readLine()
		if errors.Is(err, io.EOF) { break }
		rLen := len(r)
		if currentlyReadingMultiLine {
			if strings.HasSuffix(r, "\\") {
				currentMultiLineValueLineStack = append(currentMultiLineValueLineStack, unescape(r[:rLen-1]))
				continue
			}
			currentSection.ValueMap[currentMultiLineKey] = strings.Join(currentMultiLineValueLineStack, "\n")
			currentlyReadingMultiLine = false
			continue
		}
		
		trimmed := strings.TrimSpace(r)
		trimmedLen := len(trimmed)
		if trimmedLen <= 0 { continue }
		if strings.HasPrefix(trimmed, "//") { continue }
		matchres := reKV.FindStringSubmatch(r)
		if len(matchres) > 0 {
			if currentSection.ValueMap == nil {
				currentSection.ValueMap = make(map[string]string, 0)
			}
			key := matchres[1]
			val := matchres[2]
			if val[len(val)-1] == '\\' {
				currentMultiLineKey = key
				currentMultiLineValueLineStack = make([]string, 0)
				currentMultiLineValueLineStack = append(currentMultiLineValueLineStack, unescape(val[:len(val)-1]))
				currentlyReadingMultiLine = true
			} else {
				i := len(val)-1
				for i >= 0 && isWhite(val[i]) { i -= 1 }
				if i > 0 {
					if val[i] == '\\' {
						i += 2
						val = val[:i]
					} else {
						i += 1
						val = val[:i]
					}
				} else {
					val = ""
				}
				currentSection.ValueMap[key] = unescape(val)
			}
			continue
		}
		matchres = reSectionHeader.FindStringSubmatch(r)
		if len(matchres) > 0 {
			level := int8(len(matchres[1]))
			sectionHeader := matchres[2]
			if level <= currentSection.Level {
				br.unreadLine(r)
				return currentSection
			}
			if currentSection.Subsection == nil {
				currentSection.Subsection = make([]*MDConfSection, 0)
			}
			newSubSection := &MDConfSection{
				Level: level,
				SectionName: sectionHeader,
				ValueMap: nil,
				Subsection: nil,
			}
			currentSection.Subsection = append(currentSection.Subsection, parseInner(newSubSection, br))
		}
	}
	if currentlyReadingMultiLine {
		currentSection.ValueMap[currentMultiLineKey] = strings.Join(currentMultiLineValueLineStack, "\n")
	}
	return currentSection
}
	
func Parse(r io.Reader) *MDConfSection {
	br := bufio.NewReader(r)
	tlr := newTrueLineReader(br)
	result := &MDConfSection{
		Level: 0,
		SectionName: "",
		ValueMap: nil,
		Subsection: nil,
	}
	return parseInner(result, tlr)
}


