package rmfiles

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/juruen/rmapi/api"
	"github.com/juruen/rmapi/filetree"
	"github.com/juruen/rmapi/model"
)

// fakeApiCtx is a minimal in-memory implementation of api.ApiCtx for tests.
type fakeApiCtx struct {
	ft        filetree.FileTreeCtx
	mu        sync.Mutex
	nextID    int
	syncCount int
	moved     int
	deleted   int
}

var _ api.ApiCtx = (*fakeApiCtx)(nil)

func newFake() *fakeApiCtx {
	f := &fakeApiCtx{ft: filetree.CreateFileTreeCtx()}
	return f
}

func (f *fakeApiCtx) id() string {
	f.nextID++
	return "id-" + itoa(f.nextID)
}

func (f *fakeApiCtx) add(parentID, name, typ string) *model.Document {
	doc := &model.Document{ID: f.id(), Name: name, Type: typ, Parent: parentID}
	f.ft.AddDocument(doc)
	return doc
}

func (f *fakeApiCtx) Filetree() *filetree.FileTreeCtx { return &f.ft }
func (f *fakeApiCtx) SyncComplete() error             { f.syncCount++; return nil }
func (f *fakeApiCtx) Nuke() error                     { return nil }
func (f *fakeApiCtx) Refresh() (string, int64, error) { return "", 0, nil }
func (f *fakeApiCtx) ReplaceDocumentFile(string, string, bool) error {
	return nil
}
func (f *fakeApiCtx) FetchDocument(docId, dstPath string) error {
	return os.WriteFile(dstPath, []byte("rmdoc-bytes"), 0o600)
}

func (f *fakeApiCtx) CreateDir(parentId, name string, _ bool) (*model.Document, error) {
	return f.add(parentId, name, model.DirectoryType), nil
}

func (f *fakeApiCtx) UploadDocument(parentId, sourceDocPath string, _ bool, _, _, _ *int, _ *string) (*model.Document, error) {
	stem := filepath.Base(sourceDocPath)
	stem = stem[:len(stem)-len(filepath.Ext(stem))]
	return &model.Document{ID: f.id(), Name: stem, Type: model.DocumentType, Parent: parentId}, nil
}

func (f *fakeApiCtx) MoveEntry(src, dstDir *model.Node, name string) (*model.Node, error) {
	f.moved++
	doc := *src.Document
	doc.Name = name
	doc.Parent = dstDir.Id()
	return &model.Node{Document: &doc, Parent: dstDir}, nil
}

func (f *fakeApiCtx) DeleteEntry(node *model.Node, _, _ bool) error {
	f.deleted++
	return nil
}

// tree builds a small tree and returns the fake plus useful IDs.
func tree(t *testing.T) (*fakeApiCtx, string, string, string) {
	t.Helper()
	f := newFake()
	ai := f.add("", "ai", model.DirectoryType)
	notes := f.add(ai.ID, "Notes", model.DirectoryType)
	doc := f.add(notes.ID, "Project Plan", model.DocumentType)
	// a template document that must be hidden by default
	f.add(notes.ID, "hidden template", model.TemplateType)
	return f, ai.ID, notes.ID, doc.ID
}

func TestWalkEntriesPathsAndFiltering(t *testing.T) {
	f, _, notesID, docID := tree(t)
	entries := WalkEntries(f.ft.Root(), false)
	byID := map[string]Entry{}
	for _, e := range entries {
		byID[e.ID] = e
	}
	if _, ok := byID[notesID]; !ok {
		t.Fatalf("expected notes folder in entries: %+v", entries)
	}
	if got := byID[docID].Path; got != "/ai/Notes/Project Plan" {
		t.Fatalf("path = %q, want /ai/Notes/Project Plan", got)
	}
	for _, e := range entries {
		if e.Name == "hidden template" {
			t.Fatalf("template should be hidden by default")
		}
	}
	if with := WalkEntries(f.ft.Root(), true); len(with) != len(entries)+1 {
		t.Fatalf("include-templates length = %d, want %d", len(with), len(entries)+1)
	}
}

func TestSearchRanksNamePrefixFirst(t *testing.T) {
	f, _, _, _ := tree(t)
	entries := WalkEntries(f.ft.Root(), false)
	got := Search(entries, "proj", 10)
	if len(got) == 0 {
		t.Fatal("expected a match for proj")
	}
	if got[0].Entry.Name != "Project Plan" {
		t.Fatalf("top result = %q, want Project Plan", got[0].Entry.Name)
	}
	if len(Search(entries, "", 10)) != len(entries) {
		t.Fatal("empty query should return all entries")
	}
}

func TestCreateFolderAndReindex(t *testing.T) {
	f, _, notesID, _ := tree(t)
	s := NewService(f, Config{})
	if _, err := s.CreateFolder("/ai/Notes", "Ideas"); err != nil {
		t.Fatalf("CreateFolder: %v", err)
	}
	if got := len(s.Entries()); got != 4 { // ai, Notes, Project Plan, Ideas
		t.Fatalf("entries = %d, want 4", got)
	}
	// duplicate must conflict
	if _, err := s.CreateFolder("/ai/Notes", "Ideas"); err != ErrConflict {
		t.Fatalf("duplicate error = %v, want ErrConflict", err)
	}
	_ = notesID
}

func TestRenameAndMove(t *testing.T) {
	f, aiID, _, docID := tree(t)
	s := NewService(f, Config{})
	renamed, err := s.Rename(docID, "Renamed Plan")
	if err != nil {
		t.Fatalf("Rename: %v", err)
	}
	if renamed.Name != "Renamed Plan" {
		t.Fatalf("renamed name = %q", renamed.Name)
	}
	// rename to a sibling's name conflicts
	if _, err := s.Rename(docID, "hidden template"); err != ErrConflict {
		t.Fatalf("conflict error = %v, want ErrConflict", err)
	}
	moved, err := s.Move(docID, "/")
	if err != nil {
		t.Fatalf("Move: %v", err)
	}
	if moved.ParentID != "" {
		t.Fatalf("moved parent = %q, want root", moved.ParentID)
	}
	_ = aiID
}

func TestMoveIntoSubtreeRejected(t *testing.T) {
	f, _, notesID, _ := tree(t)
	s := NewService(f, Config{})
	// ai is not in the fixture return; resolve it.
	ai, err := s.List("/")
	if err != nil {
		t.Fatal(err)
	}
	var aiID string
	for _, e := range ai {
		if e.Name == "ai" {
			aiID = e.ID
		}
	}
	if _, err := s.Move(aiID, "/ai/Notes"); err == nil {
		t.Fatalf("moving a folder into its own subtree should fail")
	}
	_ = notesID
}

func TestDelete(t *testing.T) {
	f, _, _, docID := tree(t)
	s := NewService(f, Config{})
	n, err := s.Delete([]string{docID}, false)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if n != 1 || f.deleted != 1 {
		t.Fatalf("deleted = %d (fake %d), want 1", n, f.deleted)
	}
	if _, err := s.Delete(nil, false); err != ErrInvalid {
		t.Fatalf("empty delete error = %v, want ErrInvalid", err)
	}
}

func TestStatus(t *testing.T) {
	f, _, _, _ := tree(t)
	s := NewService(f, Config{DefaultRemoteDir: "/ai"})
	st := s.Status()
	if !st.Authenticated {
		t.Fatal("expected authenticated")
	}
	if st.DefaultRemoteDir != "/ai" {
		t.Fatalf("default dir = %q", st.DefaultRemoteDir)
	}
}

func TestSanitizeStem(t *testing.T) {
	cases := map[string]string{
		"My Notes":          "My_Notes",
		"a/b:c":             "abc",
		"__weird--":         "weird",
		"":                  "document",
		"2026-10-04 report": "2026-10-04_report",
	}
	for in, want := range cases {
		if got := sanitizeStem(in); got != want {
			t.Errorf("sanitizeStem(%q) = %q, want %q", in, got, want)
		}
	}
}
