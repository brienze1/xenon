package entity

import "testing"

func TestAIActivityTargetsNameItemsAndAdHocRequests(t *testing.T) {
	if got := NewAIItemTarget("req-1"); got != (AIActivityTarget{ID: "req-1"}) {
		t.Errorf("NewAIItemTarget() = %+v", got)
	}
	if got := NewAIRequestTarget("GET", "https://api.test/users"); got != (AIActivityTarget{Name: "GET https://api.test/users"}) {
		t.Errorf("NewAIRequestTarget() = %+v", got)
	}
}
