package git

import (
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func TestRoot(t *testing.T) {
	// When
	root, err := Root()

	// Then
	require.NoError(t, err)
	gfp := filepath.Join(root, ".git")
	_, err = os.Stat(gfp)
	require.NoError(t, err, "expected directory not found in the git root")
}
