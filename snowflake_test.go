package snowflake

import (
	"bytes"
	"reflect"
	"sync"
	"testing"
)

func TestNewSnowflake(t *testing.T) {
	var i, j int64
	for i = 0; i < 32; i++ {
		for j = 0; j < 32; j++ {
			_, err := NewSnowflake(i, j)
			if err != nil {
				t.Error(err)
			}
		}
	}
}

func TestNextVal(t *testing.T) {
	s, err := NewSnowflake(0, 0)
	if err != nil {
		t.Error(err)
	}
	var i int64
	for i = 0; i < sequenceMask*10; i++ {
		val := s.NextVal()
		if val == 0 {
			t.Fail()
		}
	}
}

func TestUnique(t *testing.T) {
	var wg sync.WaitGroup
	var check sync.Map
	s, err := NewSnowflake(0, 0)
	if err != nil {
		t.Error(err)
	}
	for i := 0; i < 1000000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Add(-1)
			val := s.NextVal()
			if _, ok := check.Load(val); ok {
				t.Fail()
				return
			}
			check.Store(val, 0)
			if val == 0 {
				t.Fail()
				return
			}
		}()
	}
	wg.Wait()
}

func TestGetTime(t *testing.T) {
	s, err := NewSnowflake(0, 1)
	if err != nil {
		t.Error(err)
	}
	val := s.NextVal()
	t.Logf("time:%v", GetGenTime(val))
}

func TestGetDeviceID(t *testing.T) {
	s, err := NewSnowflake(28, 11)
	if err != nil {
		t.Error(err)
	}
	val := s.NextVal()
	datacenterID, workerID := GetDeviceID(int64(val))
	if datacenterID != 28 || workerID != 11 {
		t.Fail()
	}
}

//******************************************************************************
// General Test funcs

func TestNewNode(t *testing.T) {

	_, err := NewSnowflake(1, 1)
	if err != nil {
		t.Fatalf("error creating NewNode, %s", err)
	}

	_, err = NewSnowflake(10, 1)
	if err == nil {
		t.Fatalf("no error creating NewNode, %s", err)
	}

}

// lazy check if Generate will create duplicate IDs
// would be good to later enhance this with more smarts
func TestGenerateDuplicateID(t *testing.T) {

	node, _ := NewSnowflake(1, 1)

	var x, y ID
	for i := 0; i < 1000000; i++ {
		y = node.NextVal()
		if x == y {
			t.Errorf("x(%d) & y(%d) are the same", x, y)
		}
		x = y
	}
}

// I feel like there's probably a better way
func TestRace(t *testing.T) {

	node, _ := NewSnowflake(1, 1)

	go func() {
		for i := 0; i < 1000000000; i++ {

			NewSnowflake(1, 1)
		}
	}()

	for i := 0; i < 4000; i++ {

		node.NextVal()
	}

}

//******************************************************************************
// Converters/Parsers Test funcs
// We should have funcs here to test conversion both ways for everything

func TestPrintAll(t *testing.T) {
	node, err := NewSnowflake(1, 1)
	if err != nil {
		t.Fatalf("error creating NewNode, %s", err)
	}

	id := node.NextVal()

	t.Logf("Int64    : %#v", id.Int64())
	t.Logf("String   : %#v", id.String())
	t.Logf("Base2    : %#v", id.Base2())
	t.Logf("Base32   : %#v", id.Base32())
	t.Logf("Base36   : %#v", id.Base36())
	t.Logf("Base58   : %#v", id.Base58())
	t.Logf("Base64   : %#v", id.Base64())
	t.Logf("Bytes    : %#v", id.Bytes())
	t.Logf("IntBytes : %#v", id.IntBytes())

}

func TestInt64(t *testing.T) {
	node, err := NewSnowflake(1, 1)
	if err != nil {
		t.Fatalf("error creating NewNode, %s", err)
	}

	oID := node.NextVal()
	i := oID.Int64()

	pID := ParseInt64(i)
	if pID != oID {
		t.Fatalf("pID %v != oID %v", pID, oID)
	}

	mi := int64(1116766490855473152)
	pID = ParseInt64(mi)
	if pID.Int64() != mi {
		t.Fatalf("pID %v != mi %v", pID.Int64(), mi)
	}

}

func TestString(t *testing.T) {
	node, err := NewSnowflake(1, 1)
	if err != nil {
		t.Fatalf("error creating NewNode, %s", err)
	}

	oID := node.NextVal()
	si := oID.String()

	pID, err := ParseString(si)
	if err != nil {
		t.Fatalf("error parsing, %s", err)
	}

	if pID != oID {
		t.Fatalf("pID %v != oID %v", pID, oID)
	}

	ms := `1116766490855473152`
	_, err = ParseString(ms)
	if err != nil {
		t.Fatalf("error parsing, %s", err)
	}

	ms = `1112316766490855473152`
	_, err = ParseString(ms)
	if err == nil {
		t.Fatalf("no error parsing %s", ms)
	}
}

func TestBase2(t *testing.T) {
	node, err := NewSnowflake(1, 1)
	if err != nil {
		t.Fatalf("error creating NewNode, %s", err)
	}

	oID := node.NextVal()
	i := oID.Base2()

	pID, err := ParseBase2(i)
	if err != nil {
		t.Fatalf("error parsing, %s", err)
	}
	if pID != oID {
		t.Fatalf("pID %v != oID %v", pID, oID)
	}

	ms := `111101111111101110110101100101001000000000000000000000000000`
	_, err = ParseBase2(ms)
	if err != nil {
		t.Fatalf("error parsing, %s", err)
	}

	ms = `1112316766490855473152`
	_, err = ParseBase2(ms)
	if err == nil {
		t.Fatalf("no error parsing %s", ms)
	}
}

func TestBase32(t *testing.T) {

	node, err := NewSnowflake(1, 1)
	if err != nil {
		t.Fatalf("error creating NewNode, %s", err)
	}

	for i := 0; i < 100; i++ {

		sf := node.NextVal()
		b32i := sf.Base32()
		psf, err := ParseBase32([]byte(b32i))
		if err != nil {
			t.Fatal(err)
		}
		if sf != psf {
			t.Fatal("Parsed does not match String.")
		}
	}
}

func TestBase36(t *testing.T) {
	node, err := NewSnowflake(1, 1)
	if err != nil {
		t.Fatalf("error creating NewNode, %s", err)
	}

	oID := node.NextVal()
	i := oID.Base36()

	pID, err := ParseBase36(i)
	if err != nil {
		t.Fatalf("error parsing, %s", err)
	}
	if pID != oID {
		t.Fatalf("pID %v != oID %v", pID, oID)
	}

	ms := `8hgmw4blvlkw`
	_, err = ParseBase36(ms)
	if err != nil {
		t.Fatalf("error parsing, %s", err)
	}

	ms = `68h5gmw443blv2lk1w`
	_, err = ParseBase36(ms)
	if err == nil {
		t.Fatalf("no error parsing, %s", err)
	}
}

func TestBase58(t *testing.T) {

	node, err := NewSnowflake(1, 1)
	if err != nil {
		t.Fatalf("error creating NewNode, %s", err)
	}

	for i := 0; i < 10; i++ {

		sf := node.NextVal()
		b58 := sf.Base58()
		psf, err := ParseBase58([]byte(b58))
		if err != nil {
			t.Fatal(err)
		}
		if sf != psf {
			t.Fatal("Parsed does not match String.")
		}
	}
}

func TestBase64(t *testing.T) {
	node, err := NewSnowflake(1, 1)
	if err != nil {
		t.Fatalf("error creating NewNode, %s", err)
	}

	oID := node.NextVal()
	i := oID.Base64()

	pID, err := ParseBase64(i)
	if err != nil {
		t.Fatalf("error parsing, %s", err)
	}
	if pID != oID {
		t.Fatalf("pID %v != oID %v", pID, oID)
	}

	ms := `MTExNjgxOTQ5NDY2MDk5NzEyMA==`
	_, err = ParseBase64(ms)
	if err != nil {
		t.Fatalf("error parsing, %s", err)
	}

	ms = `MTExNjgxOTQ5NDY2MDk5NzEyMA`
	_, err = ParseBase64(ms)
	if err == nil {
		t.Fatalf("no error parsing, %s", err)
	}
}

func TestBytes(t *testing.T) {
	node, err := NewSnowflake(1, 1)
	if err != nil {
		t.Fatalf("error creating NewNode, %s", err)
	}

	oID := node.NextVal()
	i := oID.Bytes()

	pID, err := ParseBytes(i)
	if err != nil {
		t.Fatalf("error parsing, %s", err)
	}
	if pID != oID {
		t.Fatalf("pID %v != oID %v", pID, oID)
	}

	ms := []byte{0x31, 0x31, 0x31, 0x36, 0x38, 0x32, 0x31, 0x36, 0x37, 0x39, 0x35, 0x37, 0x30, 0x34, 0x31, 0x39, 0x37, 0x31, 0x32}
	_, err = ParseBytes(ms)
	if err != nil {
		t.Fatalf("error parsing, %#v", err)
	}

	ms = []byte{0xFF, 0xFF, 0xFF, 0x31, 0x31, 0x31, 0x36, 0x38, 0x32, 0x31, 0x36, 0x37, 0x39, 0x35, 0x37, 0x30, 0x34, 0x31, 0x39, 0x37, 0x31, 0x32}
	_, err = ParseBytes(ms)
	if err == nil {
		t.Fatalf("no error parsing, %#v", err)
	}
}

func TestIntBytes(t *testing.T) {
	node, err := NewSnowflake(1, 1)
	if err != nil {
		t.Fatalf("error creating NewNode, %s", err)
	}

	oID := node.NextVal()
	i := oID.IntBytes()

	pID := ParseIntBytes(i)
	if pID != oID {
		t.Fatalf("pID %v != oID %v", pID, oID)
	}

	ms := [8]uint8{0xf, 0x7f, 0xc0, 0xfc, 0x2f, 0x80, 0x0, 0x0}
	mi := int64(1116823421972381696)
	pID = ParseIntBytes(ms)
	if pID.Int64() != mi {
		t.Fatalf("pID %v != mi %v", pID.Int64(), mi)
	}

}

//******************************************************************************
// Marshall Test Methods

func TestMarshalJSON(t *testing.T) {
	id := ID(13587)
	expected := "\"13587\""

	bytes, err := id.MarshalJSON()
	if err != nil {
		t.Fatalf("Unexpected error during MarshalJSON")
	}

	if string(bytes) != expected {
		t.Fatalf("Got %s, expected %s", string(bytes), expected)
	}
}

func TestMarshalsIntBytes(t *testing.T) {
	id := ID(13587).IntBytes()
	expected := []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x35, 0x13}
	if !bytes.Equal(id[:], expected) {
		t.Fatalf("Expected ID to be encoded as %v, got %v", expected, id)
	}
}

func TestUnmarshalJSON(t *testing.T) {
	tt := []struct {
		json        string
		expectedID  ID
		expectedErr error
	}{
		{`"13587"`, 13587, nil},
		{`1`, 0, JSONSyntaxError{[]byte(`1`)}},
		{`"invalid`, 0, JSONSyntaxError{[]byte(`"invalid`)}},
	}

	for _, tc := range tt {
		var id ID
		err := id.UnmarshalJSON([]byte(tc.json))
		if !reflect.DeepEqual(err, tc.expectedErr) {
			t.Fatalf("Expected to get error '%s' decoding JSON, but got '%s'", tc.expectedErr, err)
		}

		if id != tc.expectedID {
			t.Fatalf("Expected to get ID '%s' decoding JSON, but got '%s'", tc.expectedID, id)
		}
	}
}
