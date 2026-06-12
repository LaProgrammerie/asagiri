package replay

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const gzipSuffix = ".gz"

// CompressLargeFiles gzip-compresses files above threshold in target dirs (spec §24).
func CompressLargeFiles(root string, dirs []string, threshold int64) ([]string, error) {
	if threshold <= 0 {
		return nil, nil
	}
	var compressed []string
	for _, rel := range dirs {
		dir := filepath.Join(root, rel)
		if err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || strings.HasSuffix(path, gzipSuffix) {
				return nil
			}
			info, err := d.Info()
			if err != nil {
				return err
			}
			if info.Size() < threshold {
				return nil
			}
			gzPath, err := gzipFile(path)
			if err != nil {
				return err
			}
			compressed = append(compressed, filepath.ToSlash(strings.TrimPrefix(gzPath, root+string(os.PathSeparator))))
			return nil
		}); err != nil && !os.IsNotExist(err) {
			return compressed, fmt.Errorf("compress dir %q: %w", rel, err)
		}
	}
	return compressed, nil
}

func gzipFile(path string) (string, error) {
	src, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = src.Close() }()

	gzPath := path + gzipSuffix
	dst, err := os.Create(gzPath)
	if err != nil {
		return "", err
	}
	gw := gzip.NewWriter(dst)
	if _, err := io.Copy(gw, src); err != nil {
		_ = gw.Close()
		_ = dst.Close()
		return "", err
	}
	if err := gw.Close(); err != nil {
		_ = dst.Close()
		return "", err
	}
	if err := dst.Close(); err != nil {
		return "", err
	}
	if err := os.Remove(path); err != nil {
		return "", err
	}
	return gzPath, nil
}

// ReadMaybeCompressed reads a file or its .gz sidecar.
func ReadMaybeCompressed(path string) ([]byte, error) {
	if body, err := os.ReadFile(path); err == nil {
		return body, nil
	}
	gzPath := path + gzipSuffix
	body, err := os.ReadFile(gzPath)
	if err != nil {
		return nil, err
	}
	return decompressGzip(body)
}

func decompressGzip(body []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer func() { _ = r.Close() }()
	return io.ReadAll(r)
}
