package emby

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTestConnection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/System/Info/Public" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("X-Emby-Token") != "test-key" {
			t.Errorf("missing or wrong X-Emby-Token header: %s", r.Header.Get("X-Emby-Token"))
		}

		json.NewEncoder(w).Encode(SystemInfo{
			ServerName: "Test Emby",
			Version:    "4.8.0.0",
			ID:         "abc123",
		})
	}))
	defer server.Close()

	client := New(server.URL, "test-key")
	info, err := client.TestConnection(context.Background())
	if err != nil {
		t.Fatalf("TestConnection failed: %v", err)
	}

	if info.ServerName != "Test Emby" {
		t.Errorf("expected server name 'Test Emby', got %q", info.ServerName)
	}
	if info.Version != "4.8.0.0" {
		t.Errorf("expected version '4.8.0.0', got %q", info.Version)
	}
}

func TestGetUsers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Users" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		json.NewEncoder(w).Encode([]*User{
			{ID: "user1", Name: "Admin", Policy: &UserPolicy{IsAdministrator: true, EnableAllFolders: true}},
			{ID: "user2", Name: "Viewer", Policy: &UserPolicy{EnableAllFolders: false, EnabledFolders: []string{"lib1"}}},
		})
	}))
	defer server.Close()

	client := New(server.URL, "test-key")
	users, err := client.GetUsers(context.Background())
	if err != nil {
		t.Fatalf("GetUsers failed: %v", err)
	}

	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}
	if users[0].Name != "Admin" {
		t.Errorf("expected first user 'Admin', got %q", users[0].Name)
	}
	if !users[0].Policy.IsAdministrator {
		t.Error("expected first user to be admin")
	}
}

func TestGetLibraries(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Library/MediaFolders" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		json.NewEncoder(w).Encode(MediaFoldersResponse{
			Items: []Library{
				{ID: "lib1", Name: "Movies", CollectionType: "movies"},
				{ID: "lib2", Name: "TV Shows", CollectionType: "tvshows"},
			},
		})
	}))
	defer server.Close()

	client := New(server.URL, "test-key")
	libs, err := client.GetLibraries(context.Background())
	if err != nil {
		t.Fatalf("GetLibraries failed: %v", err)
	}

	if len(libs) != 2 {
		t.Fatalf("expected 2 libraries, got %d", len(libs))
	}
	if libs[0].Name != "Movies" {
		t.Errorf("expected first library 'Movies', got %q", libs[0].Name)
	}
}

func TestGetUserItems(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Users/user1/Items" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("IncludeItemTypes") != "Movie" {
			t.Errorf("expected IncludeItemTypes=Movie, got %q", r.URL.Query().Get("IncludeItemTypes"))
		}
		if r.URL.Query().Get("Recursive") != "true" {
			t.Errorf("expected Recursive=true, got %q", r.URL.Query().Get("Recursive"))
		}

		json.NewEncoder(w).Encode(ItemsResult{
			Items: []*Item{
				{ID: "item1", Name: "Test Movie", Type: "Movie"},
			},
			TotalRecordCount: 1,
		})
	}))
	defer server.Close()

	client := New(server.URL, "test-key")
	result, err := client.GetUserItems(context.Background(), "user1", &ItemQuery{
		IncludeTypes: "Movie",
		Recursive:    true,
	})
	if err != nil {
		t.Fatalf("GetUserItems failed: %v", err)
	}

	if result.TotalRecordCount != 1 {
		t.Errorf("expected 1 total record, got %d", result.TotalRecordCount)
	}
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(result.Items))
	}
	if result.Items[0].Name != "Test Movie" {
		t.Errorf("expected 'Test Movie', got %q", result.Items[0].Name)
	}
}

func TestConnectionFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Access denied"))
	}))
	defer server.Close()

	client := New(server.URL, "bad-key")
	_, err := client.TestConnection(context.Background())
	if err == nil {
		t.Error("expected error for unauthorized request")
	}
}

func TestTrailingSlashTrimmed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/System/Info/Public" {
			t.Errorf("unexpected path (double slash?): %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(SystemInfo{ServerName: "Test", Version: "1.0"})
	}))
	defer server.Close()

	client := New(server.URL+"/", "test-key")
	_, err := client.TestConnection(context.Background())
	if err != nil {
		t.Fatalf("TestConnection with trailing slash failed: %v", err)
	}
}
