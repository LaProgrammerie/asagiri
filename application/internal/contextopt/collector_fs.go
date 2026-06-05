package contextopt

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func statFile(path string) (fs.FileInfo, error) {
	return os.Stat(path)
}

func readFileLimited(path string, maxBytes int64) ([]byte, fs.FileInfo, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = f.Close() }()
	st, err := f.Stat()
	if err != nil {
		return nil, nil, err
	}
	size := st.Size()
	toRead := size
	if maxBytes > 0 && toRead > maxBytes {
		toRead = maxBytes
	}
	buf := make([]byte, toRead)
	n, err := io.ReadFull(f, buf)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return nil, nil, err
	}
	return buf[:n], st, nil
}

func walkShallow(root, repoRoot string, fn func(path, rel string, info fs.FileInfo) error) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return err
		}
		return fn(path, rel, info)
	})
}
