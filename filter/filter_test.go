package filter

import (
	"testing"
	"time"
)

func TestFilter(t *testing.T) {

	f := NewUniqueMessageFilter(2)

	hash1 := "abc"
	hash2 := "def"

	result := f.IfNotContainsAdd(hash1)
	if result == false {
		t.Errorf("could not add the first entry to the filter")
	}

	size := f.size()
	if size != 1 {
		t.Errorf("expected size is 1 but received %d", size)
	}

	result = f.IfNotContainsAdd(hash1)
	if result == true {
		t.Errorf("added the first entry second time")
	}

	result = f.IfNotContainsAdd(hash2)
	if result == false {
		t.Errorf("could not add the second entry to the filter")
	}

	size = f.size()
	if size != 2 {
		t.Errorf("expected size is 2 but received %d", size)
	}

	//sleeps to simulate the passing time more than ttl
	time.Sleep(3 * time.Second)

	result = f.IfNotContainsAdd(hash1)
	if result {
		t.Errorf("added the first entry second time")
	}

	size = f.size()
	if size != 1 {
		t.Errorf("expected size is 1 but received %d", size)
	}

	result = f.IfNotContainsAdd(hash2)
	if result == false {
		t.Errorf("could not add the second entry to the filter")
	}

}
