package mdconf

import (
	"bufio"
	"errors"
	"fmt"
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
			if len(val) <= 0 {
				currentSection.ValueMap[key] = ""
			} else if val[len(val)-1] == '\\' {
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

func escapeValue(s string) string {
	if len(s) >= 2 {
		if isWhite(s[0]) {
			if isWhite(s[len(s)-1]) {
				return "\\" + s[:len(s)-1] + "\\" + s[len(s)-1:]
			} else {
				return "\\" + s
			}
		} else if isWhite(s[len(s)-1]) {
			return s[:len(s)-1] + "\\" + s[len(s)-1:]
		} else {
			return s
		}
	} else if len(s) == 1 {
		if isWhite(s[0]) {
			return "\\" + s
		} else {
			return s
		}
	} else { return "" }
}

func toString(br *strings.Builder, mdc *MDConfSection) {
	br.WriteString(fmt.Sprintf("%s %s\n", strings.Repeat("#", int(mdc.Level)), mdc.SectionName))
	if mdc.ValueMap != nil {
		for k, v := range mdc.ValueMap {
			ss := strings.Split(v, "\n")
			if len(ss) > 1 {
				br.WriteString(fmt.Sprintf("+ %s: %s\\", k, escapeValue(ss[0])))
				i := 1
				for i < len(ss)-1 {
					br.WriteString(ss[i])
					br.WriteString("\\\n")
				}
				br.WriteString(ss[len(ss)-1])
				br.WriteString("\n")
			} else {
				br.WriteString(fmt.Sprintf("+ %s: %s\n", k, escapeValue(v)))
			}
		}
		br.WriteString("\n")
	}
	if mdc.Subsection != nil {
		for _, k := range mdc.Subsection {
			toString(br, k)
			br.WriteString("\n")
		}
	}
}
func (mdc *MDConfSection) ToString() string {
	res := new(strings.Builder)
	toString(res, mdc)
	return res.String()
}

func ParseString(s string) *MDConfSection {
	nr := strings.NewReader(s)
	return Parse(nr)
}

var ErrNotFound = errors.New("key not found")
var ErrEmptyKey = errors.New("key is empty")
var ErrEmptySubject = errors.New("subject is empty")

func (mdc *MDConfSection) LocalQueryKey(k string) (string, error) {
	if mdc == nil { return "", ErrNotFound }
	if mdc.ValueMap == nil { return "", ErrNotFound }
	v, ok := mdc.ValueMap[k]
	if !ok { return "", ErrNotFound }
	return v, nil
}

func (mdc *MDConfSection) QueryKey(k []string) (string, error) {
	if mdc == nil { return "", ErrNotFound }
	i := 0
	subj := mdc
	if len(k) <= 0 { return "", ErrEmptyKey }
	for i < len(k) - 1 {
		if subj.Subsection == nil { return "", ErrNotFound }
		found := false
		for _, subsec := range subj.Subsection {
			if subsec.SectionName != k[i] { continue }
			found = true
			subj = subsec
			break
		}
		if !found { return "", ErrNotFound }
		i += 1
	}
	if subj.ValueMap == nil { return "", ErrNotFound }
	v, ok := subj.ValueMap[k[i]]
	if !ok { return "", ErrNotFound }
	return v, nil
}

func (mdc *MDConfSection) LocalQuerySection(k string) (*MDConfSection, error) {
	if mdc == nil { return nil, ErrNotFound }
	if mdc.Subsection == nil { return nil, ErrNotFound }
	for _, subsec := range mdc.Subsection {
		if subsec.SectionName == k { return subsec, nil }
	}
	return nil, ErrNotFound
}

func (mdc *MDConfSection) QuerySection(k []string) (*MDConfSection, error) {
	if mdc == nil { return nil, ErrNotFound }
	i := 0
	subj := mdc
	if len(k) <= 0 { return mdc, nil }
	for i < len(k) {
		if subj.Subsection == nil { return nil, ErrNotFound }
		found := false
		for _, subsec := range subj.Subsection {
			if subsec.SectionName != k[i] { continue }
			found = true
			subj = subsec
			break
		}
		if !found { return nil, ErrNotFound }
		i += 1
	}
	return subj, nil
}

func (mdc *MDConfSection) SetKey(k []string, val string) error {
	if mdc == nil { return ErrEmptySubject }
	i := 0
	subj := mdc
	if len(k) <= 0 { return ErrEmptyKey }
	for i < len(k) - 1 {
		if subj.Subsection == nil { return ErrNotFound }
		found := false
		for _, subsec := range subj.Subsection {
			if subsec.SectionName != k[i] { continue }
			found = true
			subj = subsec
			break
		}
		if !found { return ErrNotFound }
		i += 1
	}
	if subj.ValueMap == nil {
		subj.ValueMap = make(map[string]string, 0)
	}
	subj.ValueMap[k[len(k)-1]] = val
	return nil
}

func (mdc *MDConfSection) LocalSetKey(k string, val string) error {
	if mdc == nil { return ErrEmptySubject }
	if mdc.ValueMap == nil {
		mdc.ValueMap = make(map[string]string, 0)
	}
	mdc.ValueMap[k] = val
	return nil
}

func (mdc *MDConfSection) AddSection(k []string, name string) (*MDConfSection, error) {
	// Add a new subsection at location specified by the key `k`.  The
	// section name should be not used by other subsections.  If
	// subsections with the same name exists, this function would
	// simply return that subsection and not do anything.
	if mdc == nil { return nil, ErrEmptySubject }
	i := 0
	subj := mdc
	for i < len(k) {
		if subj.Subsection == nil { return nil, ErrNotFound }
		found := false
		for _, subsec := range subj.Subsection {
			if subsec.SectionName != k[i] { continue }
			found = true
			subj = subsec
			break
		}
		if !found { return nil, ErrNotFound }
		i += 1
	}
	return subj.LocalAddSection(name)
}

func (mdc *MDConfSection) LocalAddSection(name string) (*MDConfSection, error) {
	// Add a new subsection directly at the current section.  The
	// section name should be not used by other subsections.  If
	// subsections with the same name exists, this function would
	// simply return that subsection and not do anything.
	if mdc == nil { return nil, ErrEmptySubject }
	if mdc.Subsection == nil {
		mdc.Subsection = make([]*MDConfSection, 0)
	}
	for _, v := range mdc.Subsection {
		if v.SectionName == name { return v, nil }
	}
	res := &MDConfSection{
		Level: mdc.Level + 1,
		SectionName: name,
		ValueMap: nil,
		Subsection: nil,
	}
	mdc.Subsection = append(mdc.Subsection, res)
	return res, nil
}


