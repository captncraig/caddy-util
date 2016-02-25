package util

import (
	"fmt"
	"net"
	"reflect"
	"strings"
	"testing"

	"github.com/mholt/caddy/caddy/setup"
)

type TestStruct struct {
	A string            // a foo
	B int               // b 42
	C bool              // c
	D []string          // d foo (allowed multiple times)
	E []int             // e 13 (allowed multiple times)
	G net.IP            // g 1.2.3.4
	H []net.IP          // h 1.2.3.4 (multiple)
	I net.Addr          // i 1.2.3.0\16
	J []net.Addr        // multiple cidrs
	K map[string]string // k key1 val1 \n k key2 val2
	L [][]string        // each line is one slice. All args are included
	Z string            `caddy:"m"`
}

var c1, c2 net.Addr

func init() {
	_, c1, _ = net.ParseCIDR("1.2.3.0/0")
	_, c2, _ = net.ParseCIDR("2.2.3.0/0")
}

func TestUnmarshal(t *testing.T) {
	tsts := []struct {
		input    []string
		expected *TestStruct
	}{
		{[]string{"a foo"}, &TestStruct{A: "foo"}},
		{[]string{"b 42"}, &TestStruct{B: 42}},
		{[]string{"C"}, &TestStruct{C: true}},
		{[]string{"C true"}, &TestStruct{C: true}},
		{[]string{"c false"}, &TestStruct{C: false}},
		{[]string{"d foo", "d bar"}, &TestStruct{D: []string{"foo", "bar"}}},
		{[]string{"E 17", "e -42"}, &TestStruct{E: []int{17, -42}}},
		{[]string{"g 1.2.3.4"}, &TestStruct{G: net.ParseIP("1.2.3.4")}},
		{[]string{"h 1.2.3.4", "h 2.3.4.5"}, &TestStruct{H: []net.IP{net.ParseIP("1.2.3.4"), net.ParseIP("2.3.4.5")}}},
		{[]string{"i 1.2.3.0/0"}, &TestStruct{I: c1}},
		{[]string{"j 1.2.3.0/0", "j 1.2.3.0/0"}, &TestStruct{J: []net.Addr{c1, c2}}},
		{[]string{"k a boo", `k foo "a b c d e"`}, &TestStruct{K: map[string]string{"a": "boo", "foo": "a b c d e"}}},
		{[]string{"l a b c", "l d e f g h"}, &TestStruct{L: [][]string{[]string{"a", "b", "c"}, []string{"d", "e", "f", "g", "h"}}}},
		{[]string{"m foo"}, &TestStruct{Z: "foo"}},
	}
	for i, tst := range tsts {
		input := fmt.Sprintf("{\n  %s\n}", strings.Join(tst.input, "\n  "))
		c := setup.NewTestController(input)
		ts := &TestStruct{}
		err := Unmarshal(c, ts)
		if err != nil {
			t.Errorf("Test %d error. %s", i, err)
			continue
		}
		if !reflect.DeepEqual(ts, tst.expected) {
			t.Errorf("Test %d: structs don't match", i)
		}
	}
}

type ArgTest0 struct {
	Path string `caddy:"path,arg0"`
}

type ArgComplicated struct {
	Addr net.Addr `caddy:",arg0"`
	IP   net.IP   `caddy:",arg1"`
	Num  int      `caddy:",arg2"`
	B    bool     `caddy:",arg3"`
}

type ArrayArgs struct {
	Paths []string `caddy:"p,arg0"`
}

func TestArgs(t *testing.T) {
	var argTests = []struct {
		input    string
		expected interface{}
		baseObj  interface{}
	}{
		{``, &ArgTest0{}, &ArgTest0{}}, //nothing set
		{`{
        path abc
    }`, &ArgTest0{"abc"}, &ArgTest0{}}, //standard field works
		{`abc`, &ArgTest0{"abc"}, &ArgTest0{}}, //as only arg
		{`abc {
        path def
    }`, &ArgTest0{"def"}, &ArgTest0{}}, //internal one overwrites. Not so good.
		{`1.2.3.0/0 1.2.3.4 42 true`, &ArgComplicated{c1, net.ParseIP("1.2.3.4"), 42, true}, &ArgComplicated{}},
		{`a {
            p b
            p c
        }`, &ArrayArgs{[]string{"a", "b", "c"}}, &ArrayArgs{}},
	}
	for i, tst := range argTests {
		c := setup.NewTestController(tst.input)
		err := Unmarshal(c, tst.baseObj)
		if err != nil {
			t.Errorf("Test %d error. %s", i, err)
			continue
		}
		if !reflect.DeepEqual(tst.baseObj, tst.expected) {
			t.Errorf("Test %d: structs don't match", i)
			fmt.Println(tst.expected)
			fmt.Println(tst.baseObj)
		}
	}
}
