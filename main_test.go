package mdconf

import (
	"errors"
	"testing"
)

func TestParserPrinterSanityCheck(t *testing.T) {
	r := ParseString(`
+ key1: value1
+ key2: value2

# subsection1
+ key1.1: value1.1
+ key1.2: value1.2

## subsection1.1
+ key1.1.1: 

# subsection2
//empty section
# subsection3
+ key3.1:
+ : dsdf
`)
	r1 := r.ToString()
	r2 := ParseString(r1)
	r3 := r2.ToString()
	if r1 != r3 {
		t.Error("ToString() should be able to handle empty key/value and reproduce the same result on this one")
	}
}

func TestRootKVOnly(t *testing.T) {
	r := ParseString(`
+ key1: value1
+ key2: value2
`)
	t.Run("basic things", func(t *testing.T) {
		if r.Level != 0 {
			t.Error("root section level should be 0")
		}
		if r.SectionName != "" {
			t.Error("root section name should be empty")
		}
		if r.ValueMap == nil {
			t.Error("should have kv")
		}
		if r.Subsection != nil {
			t.Error("should have no subsection")
		}
		if len(r.ValueMap) != 2 {
			t.Error("should have 2 entries")
		}
		v, ok := r.ValueMap["key1"]
		if !ok { t.Error("should have key \"key1\"") }
		if v != "value1" { t.Error("key1 should have value \"value1\"") }
		v, ok = r.ValueMap["key2"]
		if !ok { t.Error("should have key \"key2\"") }
		if v != "value2" { t.Error("key2 should have value \"value2\"") }
	})

	t.Run("query key", func(t *testing.T) {
		v, err := r.QueryKey([]string{})
		if err == nil { t.Error("should have error.") }
		if !errors.Is(err, ErrEmptyKey) { t.Error("should report ErrEmptyKey") }
		v, err = r.QueryKey([]string{"key1"})
		if err != nil {
			t.Errorf("shouldn't error: %s", err.Error())
		}
		if v != "value1" {
			t.Error("should be \"value1\"")
		}
		v, err = r.QueryKey([]string{"key2"})
		if err != nil {
			t.Errorf("shouldn't error: %s", err.Error())
		}
		if v != "value2" {
			t.Error("should be \"value2\"")
		}
	})

	t.Run("query section", func(t *testing.T) {
		v, err := r.QuerySection([]string{})
		if err != nil { t.Error("should not have error") }
		if v != r { t.Error("empty query section should return itself") }
		v, err = r.QuerySection([]string{"blah1", "blah2"})
		if err == nil { t.Error("should return error") }
		if !errors.Is(err, ErrNotFound) { t.Error("should report not found") }
	})
}


func TestRootSubsection(t *testing.T) {
	r := ParseString(`
+ key1: value1
+ key2: value2

# subsection1
+ key1.1: value1.1
+ key1.2: value1.2

## subsection1.1
+ key1.1.1: value1.1.1

# subsection2
//empty section
# subsection3
+ key3.1: value3.1
`)
	t.Run("basic things", func(t *testing.T) {
		if r.Level != 0 {
			t.Error("root section level should be 0")
		}
		if r.SectionName != "" {
			t.Error("root section name should be empty")
		}
		if r.ValueMap == nil {
			t.Error("should have kv")
		}
		if r.Subsection == nil {
			t.Error("should have subsection")
		}
		if len(r.ValueMap) != 2 {
			t.Error("kv should have 2 entries")
		}
		v, ok := r.ValueMap["key1"]
		if !ok { t.Error("should have key \"key1\"") }
		if v != "value1" { t.Error("key1 should have value \"value1\"") }
		v, ok = r.ValueMap["key2"]
		if !ok { t.Error("should have key \"key2\"") }
		if v != "value2" { t.Error("key2 should have value \"value2\"") }
		if len(r.Subsection) != 3 {
			t.Error("should have 3 subsection")
		}
		if r.Subsection[0].Level != 1 { t.Error("first subsection should be level 1") }
		if r.Subsection[0].SectionName != "subsection1" { t.Error("first subsection title should be \"subsection1\"") }
		if r.Subsection[0].ValueMap == nil { t.Error("first subsection should have kv") }
		if len(r.Subsection[0].ValueMap) != 2 { t.Error("first subsection kv should have 2 pair") }
		v, ok = r.Subsection[0].ValueMap["key1.1"]
		if !ok { t.Error("first subsection should have key \"key1.1\"") }
		if v != "value1.1" { t.Error("first subsection key \"key1.1\" should have value \"value1.1\"") }
		v, ok = r.Subsection[0].ValueMap["key1.2"]
		if !ok { t.Error("first subsection should have key \"key1.2\"") }
		if v != "value1.2" { t.Error("first subsection key \"key1.1\" should have value \"value1.2\"") }
		if r.Subsection[0].Subsection == nil {
			t.Error("first subsetion should have subsection")
		}
		if len(r.Subsection[0].Subsection) != 1 {
			t.Error("first subsection should have 1 subsection")
		}
		if r.Subsection[0].Subsection[0].Level != 2 {
			t.Error("first subsection of first subsection should have level 2")
		}
		if r.Subsection[0].Subsection[0].SectionName != "subsection1.1" {
			t.Error("first subsection of first subsection should have name \"subsection1.1\"")
		}
		if r.Subsection[0].Subsection[0].ValueMap == nil {
			t.Error("first subsection of first subsection should have kv")
		}
		if len(r.Subsection[0].Subsection[0].ValueMap) != 1 {
			t.Error("kv of first subsection of first subsection should have 1 pair")
		}
		v, ok = r.Subsection[0].Subsection[0].ValueMap["key1.1.1"]
		if !ok { t.Error("first subsection of first subsection should have key \"key.1.1\"") }
		if v != "value1.1.1" { t.Error("key1.1.1 of first subsection of first subsection should have value \"value1.1.1\"") }
		if r.Subsection[1].Level != 1 {
			t.Error("second subsection should have level 1")
		}
		if r.Subsection[1].SectionName != "subsection2" {
			t.Error("second subsection should have title \"subsection2\"")
		}
		if r.Subsection[1].ValueMap != nil {
			t.Error("second subsection should have no value map")
		}
		if r.Subsection[1].Subsection != nil {
			t.Error("second subsection should have no subsection")
		}
		if r.Subsection[2].Level != 1 {
			t.Error("third subsection should have level 1")
		}
		if r.Subsection[2].SectionName != "subsection3" {
			t.Error("third subsection should have title \"subsection2\"")
		}
		if r.Subsection[2].ValueMap == nil {
			t.Error("third subsection should have value map")
		}
		if len(r.Subsection[2].ValueMap) != 1 {
			t.Error("third subsection kv should have 1 pair")
		}
		v, ok = r.Subsection[2].ValueMap["key3.1"]
		if !ok {
			t.Error("third subsection kv should have key \"key3.1\"")
		}
		if v != "value3.1" {
			t.Error("third subsection kv \"key3.1\" should have value \"value3.1\"")
		}
		if r.Subsection[2].Subsection != nil {
			t.Error("third subsection should have no subsection")
		}
	})
	t.Run("query key", func(t *testing.T) {
		v, err := r.QueryKey([]string{"key1"})
		if err != nil { t.Error("querying key1 shouldn't report error") }
		if v != "value1" { t.Error("key1 should return value1") }
		v, err = r.QueryKey([]string{"subsection1", "key1.1"})
		if err != nil { t.Error("querying subsection1:key1.1 shouldn't report error") }
		if v != "value1.1" { t.Error("subsection1:key1.1 should return value1.1") }
		v, err = r.QueryKey([]string{"subsection1", "key1.2"})
		if err != nil { t.Error("querying subsection1:key1.2 shouldn't report error") }
		if v != "value1.2" { t.Error("subsection1:key1.2 should return value1.1") }
		v, err = r.QueryKey([]string{"subsection1", "subsection1.1", "key1.1.1"})
		if err != nil { t.Error("querying subsection1:subsection1.1:key1.1.1 shouldn't report error") }
		if v != "value1.1.1" { t.Error("subsection1:subsection1.1:key1.1.1 should return value1.1.1") }
		v, err = r.QueryKey([]string{"subsection3", "key3.1"})
		if err != nil { t.Error("querying subsection3:key3.1 shouldn't report error") }
		if v != "value3.1" { t.Error("subsection3:key3.1 should return value3.1") }
	})
	t.Run("query section", func(t *testing.T) {
		v, err := r.QuerySection([]string{"subsection1"})
		if err != nil { t.Error("querying subsection1 shouldn't report error") }
		if v.Level != 1 { t.Error("subsection1 should be level 1") }
		if v.SectionName != "subsection1" { t.Error("querying for subsection1 should give subsection1") }
		v, err = r.QuerySection([]string{"subsection1", "subsection1.1"})
		if err != nil { t.Error("querying subsection1:subsection1.1 shouldn't report error") }
		if v.Level != 2 { t.Error("subsection1:subsection1.1 should be level 2") }
		if v.SectionName != "subsection1.1" { t.Error("querying for subsection1:subsection1.1 should give subsection1.1") }
		v, err = r.QuerySection([]string{"subsection2"})
		if err != nil { t.Error("querying subsection2 shouldn't report error") }
		if v.Level != 1 { t.Error("subsection2 should be level 1") }
		if v.SectionName != "subsection2" { t.Error("querying for subsection2 should give subsection2") }
		v, err = r.QuerySection([]string{"subsection3"})
		if err != nil { t.Error("querying subsection3 shouldn't report error") }
		if v.Level != 1 { t.Error("subsection3 should be level 1") }
		if v.SectionName != "subsection3" { t.Error("querying for subsection3 should give subsection3") }
		
	})
}

func TestSetKeyAddSection(t *testing.T) {
	r := ParseString("")
	r.SetKey([]string{"first-key"}, "first-value")
	r.SetKey([]string{"second-key"}, "second-value")
	v, err := r.QueryKey([]string{"first-key"})
	if err != nil { t.Error("shouldn't produce error here") }
	if v != "first-value" { t.Error("should produce \"first-value\" here") }
	v, err = r.QueryKey([]string{"second-key"})
	if err != nil { t.Error("shouldn't produce error here") }
	if v != "second-value" { t.Error("should produce \"second-value\" here") }
	r.AddSection([]string{}, "first-section")
	s, err := r.QuerySection([]string{"first-section"})
	if err != nil { t.Error("shouldn't produce error here") }
	if s == nil { t.Error("shouldn't produce error here") }
	if s.Level != 1 { t.Error("should produce level 1 section here") }
	if s.SectionName != "first-section" { t.Error("the name should be \"first-section\" here") }
	s.SetKey([]string{"third-key"}, "third-value")
	v, err = r.QueryKey([]string{"first-section", "third-key"})
	if err != nil { t.Error("shouldn't produce error here") }
	if v != "third-value" { t.Error("should be \"third-value\" here") }
}

