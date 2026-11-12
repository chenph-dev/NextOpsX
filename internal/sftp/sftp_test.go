package sftp

import (
	"errors"
	"testing"
	"time"

	"nextopsx/internal/sshexec"
)

func TestSortEntries(t *testing.T) {
	entries := sortEntries([]Entry{
		{Name: "b.txt"},
		{Name: "Alpha"},
		{Name: "zdir", IsDir: true},
		{Name: "adir", IsDir: true},
	})
	want := []string{"adir", "zdir", "Alpha", "b.txt"}
	for i, name := range want {
		if entries[i].Name != name {
			t.Fatalf("order = %+v, want %v", entries, want)
		}
	}
}

func TestQueuedTaskEmitsRunningEvent(t *testing.T) {
	events := make(chan TaskInfo, 16)
	m := NewManager(
		func(string) (*sshexec.Host, error) { return nil, errors.New("no host") },
		func(t *Task) { events <- t.Snapshot() },
		nil,
	)
	tsk := m.newTask("h1", "upload", "/does/not/matter", "/remote/x", 1)
	m.enqueue(tsk)

	deadline := time.After(3 * time.Second)
	sawRunning := false
	for {
		select {
		case info := <-events:
			if info.Status == string(TaskRunning) {
				sawRunning = true
			}
			if info.Status == string(TaskError) {
				if !sawRunning {
					t.Fatal("task errored without a running event")
				}
				return
			}
		case <-deadline:
			t.Fatalf("no terminal event; running seen=%v", sawRunning)
		}
	}
}
