package claude

import "testing"

func TestExtractJSONToleratesProseAndFences(t *testing.T) {
	reply := "Here you go:\n```json\n{\"topics\": [{\"text\": \"Ownership\", \"why\": \"x\", \"default\": false}]}\n```"
	var a AgendaDraft
	if err := extractJSON(reply, &a); err != nil {
		t.Fatal(err)
	}
	if len(a.Topics) != 1 || a.Topics[0].Text != "Ownership" || *a.Topics[0].Default {
		t.Fatalf("got %+v", a)
	}
}
