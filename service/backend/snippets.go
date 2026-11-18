package backend

import snippetsstore "nextopsx/store/snippets"

// ListCommandSnippets returns all command snippets.
func (b *Backend) ListCommandSnippets() ([]snippetsstore.Snippet, error) {
	return b.snippets.List()
}

// SaveCommandSnippet inserts or updates a command snippet.
func (b *Backend) SaveCommandSnippet(s snippetsstore.Snippet) (snippetsstore.Snippet, error) {
	return b.snippets.Save(s)
}

// DeleteCommandSnippet removes a command snippet.
func (b *Backend) DeleteCommandSnippet(id string) error {
	return b.snippets.Delete(id)
}
