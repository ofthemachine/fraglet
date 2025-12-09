package fraglet

import (
	"io/fs"

	"github.com/ofthemachine/fraglet/pkg/embed"
)

// getEmbeddedEnvelopesFS returns the embedded envelopes filesystem
func getEmbeddedEnvelopesFS() fs.FS {
	// Extract the envelopes subdirectory from the embedded FS
	sub, err := fs.Sub(embed.Envelopes, "envelopes")
	if err != nil {
		// This shouldn't happen, but return the root FS as fallback
		return embed.Envelopes
	}
	return sub
}
