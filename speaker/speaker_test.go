package speaker

import "testing"

func TestNew(t *testing.T) {
	myAS := int16(12345)
	router := New(myAS)
	if router == nil {
		t.Error("Did not expect our new speaker to be nil")
	}
	if router.myAS != myAS {
		t.Errorf("Expected myAS to be %d but got %d", myAS, router.myAS)
	}
	if len(router.peers) != 0 {
		t.Errorf("Expected no peers but found %d", len(router.peers))
	}
}
