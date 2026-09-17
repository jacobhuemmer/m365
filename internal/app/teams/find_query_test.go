package teams

import (
	"testing"

	"github.com/masonhuemmer/m365/internal/domain"
)

func TestParseQuery(t *testing.T) {
	q, err := ParseQuery("Ajay", false)
	if err != nil || q.Intent != IntentPerson || q.Needle != "Ajay" {
		t.Fatalf("%+v %v", q, err)
	}
	q, err = ParseQuery("ajay@example.com", false)
	if err != nil || q.Intent != IntentPerson {
		t.Fatal(q, err)
	}
	q, err = ParseQuery("Ajay Kumar", false)
	if err != nil || q.Intent != IntentPerson {
		t.Fatal(q, err)
	}
	q, err = ParseQuery("NOC", true)
	if err != nil || q.Intent != IntentGroup || q.Needle != "NOC" {
		t.Fatal(q, err)
	}
	q, err = ParseQuery("the group with Ajay", false)
	if err != nil || q.Intent != IntentGroup || q.Needle != "Ajay" {
		t.Fatal(q, err)
	}
	_, err = ParseQuery("  ", false)
	if domain.ExitOf(err) != domain.ExitUsage {
		t.Fatal(err)
	}
}
