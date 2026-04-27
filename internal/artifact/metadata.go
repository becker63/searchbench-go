package artifact

import (
	"encoding/hex"
	"path/filepath"
	"time"
)

func buildMetadata(bundleID string, createdAt time.Time, files []BundleFile) BundleMetadata {
	return BundleMetadata{
		SchemaVersion: schemaVersion,
		BundleID:      bundleID,
		CreatedAt:     createdAt.UTC(),
		Files:         append([]BundleFile(nil), files...),
	}
}

func fileRecord(kind string, path string, mediaType string, sha []byte) BundleFile {
	file := BundleFile{
		Kind:      kind,
		Path:      filepath.ToSlash(path),
		MediaType: mediaType,
	}
	if len(sha) > 0 {
		file.SHA256 = hex.EncodeToString(sha)
	}
	return file
}
