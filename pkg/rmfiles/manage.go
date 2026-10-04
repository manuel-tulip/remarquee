package rmfiles

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/juruen/rmapi/api"
	"github.com/juruen/rmapi/model"
	"github.com/pkg/errors"
)

// CreateFolder creates a directory named name under parentPath.
func (s *Service) CreateFolder(parentPath, name string) (Entry, error) {
	name = strings.TrimSpace(name)
	if name == "" || strings.ContainsAny(name, "/\\") {
		return Entry{}, ErrInvalid
	}
	var entry Entry
	err := s.mutate(func(c api.ApiCtx) error {
		parent, err := s.resolveDirLocked(parentPath)
		if err != nil {
			return err
		}
		if existing, err := parent.FindByName(name); err == nil && existing != nil {
			return ErrConflict
		}
		doc, err := c.CreateDir(parent.Id(), name, true)
		if err != nil {
			return errors.Wrap(err, "create folder")
		}
		c.Filetree().AddDocument(doc)
		entry = nodeToEntry(c.Filetree().NodeById(doc.ID))
		return nil
	})
	return entry, err
}

// Rename renames an entry in place (same parent).
func (s *Service) Rename(id, newName string) (Entry, error) {
	newName = strings.TrimSpace(newName)
	if newName == "" || strings.ContainsAny(newName, "/\\") {
		return Entry{}, ErrInvalid
	}
	var entry Entry
	err := s.mutate(func(c api.ApiCtx) error {
		node, err := s.nodeByIDLocked(id)
		if err != nil {
			return err
		}
		parent := node.Parent
		if parent == nil {
			return ErrInvalid
		}
		if sib, err := parent.FindByName(newName); err == nil && sib.Id() != node.Id() {
			return ErrConflict
		}
		n, err := c.MoveEntry(node, parent, newName)
		if err != nil {
			return errors.Wrap(err, "rename entry")
		}
		c.Filetree().MoveNode(node, n)
		entry = nodeToEntry(node)
		return nil
	})
	return entry, err
}

// Move moves an entry into destDirPath, keeping its name.
func (s *Service) Move(id, destDirPath string) (Entry, error) {
	var entry Entry
	err := s.mutate(func(c api.ApiCtx) error {
		node, err := s.nodeByIDLocked(id)
		if err != nil {
			return err
		}
		dest, err := s.resolveDirLocked(destDirPath)
		if err != nil {
			return err
		}
		if isSubdir(node, dest) {
			return errors.Wrap(ErrInvalid, "cannot move an entry into itself")
		}
		if sib, err := dest.FindByName(node.Name()); err == nil && sib.Id() != node.Id() {
			return ErrConflict
		}
		n, err := c.MoveEntry(node, dest, node.Name())
		if err != nil {
			return errors.Wrap(err, "move entry")
		}
		c.Filetree().MoveNode(node, n)
		entry = nodeToEntry(node)
		return nil
	})
	return entry, err
}

// Delete removes entries by ID. Folders require recursive=true when non-empty.
func (s *Service) Delete(ids []string, recursive bool) (int, error) {
	if len(ids) == 0 {
		return 0, ErrInvalid
	}
	deleted := 0
	err := s.mutate(func(c api.ApiCtx) error {
		for _, id := range ids {
			node, err := s.nodeByIDLocked(id)
			if err != nil {
				return err
			}
			if err := c.DeleteEntry(node, recursive, true); err != nil {
				return errors.Wrapf(err, "delete %q", node.Name())
			}
			c.Filetree().DeleteNode(node)
			deleted++
		}
		return nil
	})
	return deleted, err
}

// Download fetches a document as an .rmdoc archive and returns its bytes and a
// suggested filename.
func (s *Service) Download(id string) ([]byte, string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	node, err := s.nodeByIDLocked(id)
	if err != nil {
		return nil, "", err
	}
	if node.IsDirectory() {
		return nil, "", ErrUnsupported
	}
	tmpDir, err := os.MkdirTemp("", "rmfiles-dl-")
	if err != nil {
		return nil, "", err
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()
	out := filepath.Join(tmpDir, node.Name()+".rmdoc")
	if err := s.apiCtx.FetchDocument(id, out); err != nil {
		return nil, "", errors.Wrap(err, "fetch document")
	}
	data, err := os.ReadFile(out)
	if err != nil {
		return nil, "", err
	}
	return data, node.Name() + ".rmdoc", nil
}

// isSubdir reports whether child is the same node as parent or nested beneath
// it. Used to reject "move into own subtree" operations that would detach data.
func isSubdir(parent, child *model.Node) bool {
	for n := child; n != nil; n = n.Parent {
		if n.Id() == parent.Id() {
			return true
		}
	}
	return false
}
