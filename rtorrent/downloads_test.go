package rtorrent

import (
	"reflect"
	"strings"
	"testing"
)

func TestClientDownloadsAll(t *testing.T) {
	wantDownloads := []string{
		strings.Repeat("A", 40),
		strings.Repeat("B", 40),
		strings.Repeat("C", 40),
	}

	c, done := testClient(t, downloadList, []string{""}, wantDownloads)
	defer done()
	ds := &DownloadService{C: c}

	downloads, err := ds.All()
	if err != nil {
		t.Fatalf("failed call to Client.Downloads.All: %v", err)
	}

	if want, got := wantDownloads, downloads; !reflect.DeepEqual(want, got) {
		t.Fatalf("unexpected downloads:\n- want: %v\n-  got: %v",
			want, got)
	}
}

func TestClientDownloadsStarted(t *testing.T) {
	wantDownloads := []string{
		strings.Repeat("A", 40),
	}

	c, done := testClient(t, downloadList, []string{"started"}, wantDownloads)
	defer done()
	ds := &DownloadService{C: c}

	downloads, err := ds.Started()
	if err != nil {
		t.Fatalf("failed call to Client.Downloads.Started: %v", err)
	}

	if want, got := wantDownloads, downloads; !reflect.DeepEqual(want, got) {
		t.Fatalf("unexpected downloads:\n- want: %v\n-  got: %v",
			want, got)
	}
}

func TestClientDownloadsStopped(t *testing.T) {
	wantDownloads := []string{
		strings.Repeat("A", 40),
	}

	c, done := testClient(t, downloadList, []string{"stopped"}, wantDownloads)
	defer done()
	ds := &DownloadService{C: c}

	downloads, err := ds.Stopped()
	if err != nil {
		t.Fatalf("failed call to Client.Downloads.Stopped: %v", err)
	}

	if want, got := wantDownloads, downloads; !reflect.DeepEqual(want, got) {
		t.Fatalf("unexpected downloads:\n- want: %v\n-  got: %v",
			want, got)
	}
}

func TestClientDownloadsComplete(t *testing.T) {
	wantDownloads := []string{
		strings.Repeat("A", 40),
	}

	c, done := testClient(t, downloadList, []string{"complete"}, wantDownloads)
	defer done()
	ds := &DownloadService{C: c}

	downloads, err := ds.Complete()
	if err != nil {
		t.Fatalf("failed call to Client.Downloads.Complete: %v", err)
	}

	if want, got := wantDownloads, downloads; !reflect.DeepEqual(want, got) {
		t.Fatalf("unexpected downloads:\n- want: %v\n-  got: %v",
			want, got)
	}
}

func TestClientDownloadsIncomplete(t *testing.T) {
	wantDownloads := []string{
		strings.Repeat("A", 40),
	}

	c, done := testClient(t, downloadList, []string{"incomplete"}, wantDownloads)
	defer done()
	ds := &DownloadService{C: c}

	downloads, err := ds.Incomplete()
	if err != nil {
		t.Fatalf("failed call to Client.Downloads.Incomplete: %v", err)
	}

	if want, got := wantDownloads, downloads; !reflect.DeepEqual(want, got) {
		t.Fatalf("unexpected downloads:\n- want: %v\n-  got: %v",
			want, got)
	}
}

func TestClientDownloadsHashing(t *testing.T) {
	wantDownloads := []string{
		strings.Repeat("A", 40),
	}

	c, done := testClient(t, downloadList, []string{"hashing"}, wantDownloads)
	defer done()
	ds := &DownloadService{C: c}

	downloads, err := ds.Hashing()
	if err != nil {
		t.Fatalf("failed call to Client.Downloads.Hashing: %v", err)
	}

	if want, got := wantDownloads, downloads; !reflect.DeepEqual(want, got) {
		t.Fatalf("unexpected downloads:\n- want: %v\n-  got: %v",
			want, got)
	}
}

func TestClientDownloadsSeeding(t *testing.T) {
	wantDownloads := []string{
		strings.Repeat("A", 40),
	}

	c, done := testClient(t, downloadList, []string{"seeding"}, wantDownloads)
	defer done()
	ds := &DownloadService{C: c}

	downloads, err := ds.Seeding()
	if err != nil {
		t.Fatalf("failed call to Client.Downloads.Seeding: %v", err)
	}

	if want, got := wantDownloads, downloads; !reflect.DeepEqual(want, got) {
		t.Fatalf("unexpected downloads:\n- want: %v\n-  got: %v",
			want, got)
	}
}

func TestClientDownloadsLeeching(t *testing.T) {
	wantDownloads := []string{
		strings.Repeat("A", 40),
	}

	c, done := testClient(t, downloadList, []string{"leeching"}, wantDownloads)
	defer done()
	ds := &DownloadService{C: c}

	downloads, err := ds.Leeching()
	if err != nil {
		t.Fatalf("failed call to Client.Downloads.Leeching: %v", err)
	}

	if want, got := wantDownloads, downloads; !reflect.DeepEqual(want, got) {
		t.Fatalf("unexpected downloads:\n- want: %v\n-  got: %v",
			want, got)
	}
}

func TestClientDownloadsActive(t *testing.T) {
	wantDownloads := []string{
		strings.Repeat("A", 40),
	}

	c, done := testClient(t, downloadList, []string{"active"}, wantDownloads)
	defer done()
	ds := &DownloadService{C: c}

	downloads, err := ds.Active()
	if err != nil {
		t.Fatalf("failed call to Client.Downloads.Active: %v", err)
	}

	if want, got := wantDownloads, downloads; !reflect.DeepEqual(want, got) {
		t.Fatalf("unexpected downloads:\n- want: %v\n-  got: %v",
			want, got)
	}
}

func TestClientDownloadsBaseFilename(t *testing.T) {
	wantName := "foobar"
	wantHash := strings.Repeat("A", 40)

	c, done := testClient(t, "d.base_filename", []string{wantHash}, wantName)
	defer done()
	ds := &DownloadService{C: c}

	name, err := ds.BaseFilename(wantHash)
	if err != nil {
		t.Fatalf("failed call to Client.Downloads.BaseFilename: %v", err)
	}

	if want, got := wantName, name; !reflect.DeepEqual(want, got) {
		t.Fatalf("unexpected name:\n- want: %v\n-  got: %v",
			want, got)
	}
}

func TestClientDownloadsDownloadRate(t *testing.T) {
	wantRate := 1024
	wantHash := strings.Repeat("A", 40)

	c, done := testClient(t, "d.down.rate", []string{wantHash}, wantRate)
	defer done()
	ds := &DownloadService{C: c}

	rate, err := ds.DownloadRate(wantHash)
	if err != nil {
		t.Fatalf("failed call to Client.Downloads.DownloadRate: %v", err)
	}

	if want, got := wantRate, rate; !reflect.DeepEqual(want, got) {
		t.Fatalf("unexpected download rate:\n- want: %v\n-  got: %v",
			want, got)
	}
}

func TestClientDownloadsDownloadTotal(t *testing.T) {
	wantTotal := 1024
	wantHash := strings.Repeat("A", 40)

	c, done := testClient(t, "d.down.total", []string{wantHash}, wantTotal)
	defer done()
	ds := &DownloadService{C: c}

	total, err := ds.DownloadTotal(wantHash)
	if err != nil {
		t.Fatalf("failed call to Client.Downloads.DownloadTotal: %v", err)
	}

	if want, got := wantTotal, total; !reflect.DeepEqual(want, got) {
		t.Fatalf("unexpected download total:\n- want: %v\n-  got: %v",
			want, got)
	}
}

func TestClientDownloadsUploadRate(t *testing.T) {
	wantRate := 1024
	wantHash := strings.Repeat("A", 40)

	c, done := testClient(t, "d.up.rate", []string{wantHash}, wantRate)
	defer done()
	ds := &DownloadService{C: c}

	rate, err := ds.UploadRate(wantHash)
	if err != nil {
		t.Fatalf("failed call to Client.Downloads.UploadRate: %v", err)
	}

	if want, got := wantRate, rate; !reflect.DeepEqual(want, got) {
		t.Fatalf("unexpected upload rate:\n- want: %v\n-  got: %v",
			want, got)
	}
}

func TestClientDownloadsUploadTotal(t *testing.T) {
	wantTotal := 1024
	wantHash := strings.Repeat("A", 40)

	c, done := testClient(t, "d.up.total", []string{wantHash}, wantTotal)
	defer done()
	ds := &DownloadService{C: c}

	total, err := ds.UploadTotal(wantHash)
	if err != nil {
		t.Fatalf("failed call to Client.Downloads.UploadTotal: %v", err)
	}

	if want, got := wantTotal, total; !reflect.DeepEqual(want, got) {
		t.Fatalf("unexpected upload total:\n- want: %v\n-  got: %v",
			want, got)
	}
}
