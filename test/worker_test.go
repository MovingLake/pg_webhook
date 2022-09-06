package test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/movinglake/pg_webhook/lib"
	"github.com/movinglake/pg_webhook/providers"
)

type MockDB struct {
	Inserts []string
	Closes  []string
	Opens   []string
}

func (m *MockDB) InsertTask(t providers.Task) {
	m.Inserts = append(m.Inserts, fmt.Sprintf("%v", t))
}

func (m *MockDB) Close() {
	m.Closes = append(m.Closes, "closed")
}

func (m *MockDB) OpenDB() {
	m.Opens = append(m.Opens, "opened")
}

type MockHttpClient struct {
	Recorder      []string
	PostResponses []http.Response
	PostErrors    []error
	Index         int
}

func (m *MockHttpClient) Post(url, contentType string, body io.Reader) (*http.Response, error) {
	fmt.Print("post")
	m.Recorder = append(m.Recorder, fmt.Sprintf("%v %v %v", url, contentType, body))
	defer func() { m.Index += 1 }()
	return &m.PostResponses[m.Index], m.PostErrors[m.Index]
}

func TestProcessWebhookTask(t *testing.T) {
	tasks := make(chan providers.Task, 2)
	db := MockDB{}
	cli := MockHttpClient{}
	cli.PostResponses = []http.Response{
		{StatusCode: 200},
		{StatusCode: 200},
	}
	cli.PostErrors = []error{
		nil, nil,
	}
	// Put two in. As soon as there is zero we know it completed the first loop.
	for i := 0; i < 2; i++ {
		tasks <- providers.Task{
			Ctx:    context.Background(),
			Schema: "public",
			Table:  "users",
			Verb:   "INSERT",
			Values: map[string]interface{}{"id": 1, "name": "test"},
		}
	}
	go lib.ProcessWebhookTasks(tasks, "http://localhost:8080", &db, &cli)
	t.Log("waiting for worker to finish")
	for len(tasks) > 0 {
	}

	if len(cli.Recorder) < 1 {
		t.Fatalf("Expected 1 call to Post, got %v", len(cli.Recorder))
	}
	if cli.Recorder[0] != "http://localhost:8080 application/json {\"_mlake_schema\":\"public\",\"_mlake_table\":\"users\",\"_mlake_verb\":\"INSERT\",\"id\":1,\"name\":\"test\"}" {
		t.Errorf("Expected call to Post with url http://localhost:8080, content-type application/json, and body {\"_mlake_schema\":\"public\",\"_mlake_table\":\"users\",\"_mlake_verb\":\"INSERT\",\"id\":1,\"name\":\"test\"}, got %v", cli.Recorder[0])
	}

}

func TestProcessWebhookTaskRetriesExceeded(t *testing.T) {
	tasks := make(chan providers.Task, 2)
	db := MockDB{}
	cli := MockHttpClient{}
	cli.PostResponses = []http.Response{
		{StatusCode: 500},
		{StatusCode: 200},
	}
	cli.PostErrors = []error{
		nil, fmt.Errorf("error"),
	}
	// Put two in. As soon as there is zero we know it completed the first loop.
	for i := 0; i < 2; i++ {
		tasks <- providers.Task{
			Ctx:     context.Background(),
			Schema:  "public",
			Table:   "users",
			Verb:    "INSERT",
			Values:  map[string]interface{}{"id": 1, "name": "test"},
			Retries: lib.MaxRetries + 1,
		}
	}
	go lib.ProcessWebhookTasks(tasks, "http://localhost:8080", &db, &cli)
	t.Log("waiting for worker to finish")
	for len(tasks) > 0 {
	}

	if len(db.Inserts) < 1 {
		t.Fatalf("Expected 1 call to InsertTask, got %v", len(db.Inserts))
	}
	if db.Inserts[0] != "{context.Background public users INSERT map[id:1 name:test] 401}" {
		t.Errorf("Expected call to InsertTask with {context.Background public users INSERT map[id:1 name:test] 401}, got %v", db.Inserts[0])
	}

}
