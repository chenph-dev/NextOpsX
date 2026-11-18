// Package snippets manages reusable command snippets.
package snippets

import snippetsstore "nextopsx/store/snippets"

// Service manages reusable command snippets.
type Service struct {
	store *snippetsstore.SnippetStore
}

// New builds the snippet service.
func New(store *snippetsstore.SnippetStore) *Service {
	return &Service{store: store}
}

// ListCommandSnippets returns all command snippets.
func (s *Service) ListCommandSnippets() ([]snippetsstore.Snippet, error) {
	return s.store.List()
}

// SaveCommandSnippet inserts or updates a command snippet.
func (s *Service) SaveCommandSnippet(snp snippetsstore.Snippet) (snippetsstore.Snippet, error) {
	return s.store.Save(snp)
}

// DeleteCommandSnippet removes a command snippet.
func (s *Service) DeleteCommandSnippet(id string) error {
	return s.store.Delete(id)
}
