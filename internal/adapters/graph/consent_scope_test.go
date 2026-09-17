package graph

import (
	"net/http"
	"testing"
)

func TestFakeRejectsSharedAndSites(t *testing.T) {
	srv := NewFakeServer(Seed())
	defer srv.Close()
	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/users/other/calendars", nil)
	req.Header.Set("Authorization", "Bearer fake-calendar")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("users path %d", res.StatusCode)
	}
	req, _ = http.NewRequest(http.MethodGet, srv.URL+"/sites/x/drive", nil)
	req.Header.Set("Authorization", "Bearer fake-files")
	res, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("sites %d", res.StatusCode)
	}
}

func TestFakeBothDeniesCalendar(t *testing.T) {
	srv := NewFakeServer(Seed())
	defer srv.Close()
	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/me/calendars", nil)
	req.Header.Set("Authorization", "Bearer fake-both")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("fake-both calendar %d", res.StatusCode)
	}
}
