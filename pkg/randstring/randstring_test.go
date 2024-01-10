package randstring

import "testing"

func TestRandString(t *testing.T) {
	strEmpty := RandString(0)
	if len(strEmpty) != 0 {
		t.Errorf("RandString(0) = \"%s\", want a string of length %d", strEmpty, 0)
	}
	strNormal := RandString(8)
	if len(strNormal) != 8 {
		t.Errorf("RandString(0) = \"%s\", want a string of length %d", strNormal, 8)
	}
	strLong := RandString(200)
	if len(strLong) != 200 {
		t.Errorf("RandString(0) = \"%s\", want a string of length %d", strLong, 200)
	}
}

func TestRandAlphanumString(t *testing.T) {
	strEmpty := RandAlphanumString(0)
	if len(strEmpty) != 0 {
		t.Errorf("RandAlphanumString(0) = \"%s\", want a string of length %d", strEmpty, 0)
	}
	strNormal := RandAlphanumString(8)
	if len(strNormal) != 8 {
		t.Errorf("RandAlphanumString(0) = \"%s\", want a string of length %d", strNormal, 8)
	}
	strLong := RandAlphanumString(200)
	if len(strLong) != 200 {
		t.Errorf("RandAlphanumString(0) = \"%s\", want a string of length %d", strLong, 200)
	}
}
