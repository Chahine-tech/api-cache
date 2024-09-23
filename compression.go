package apicache

import (
	"bytes"
	"compress/gzip"
	"io"
)

func compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(data); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func decompress(data []byte) ([]byte, error) {
	buf := bytes.NewReader(data)
	gz, err := gzip.NewReader(buf)
	if err != nil {
		return nil, err
	}
	defer gz.Close()
	var outBuf bytes.Buffer
	if _, err := io.Copy(&outBuf, gz); err != nil {
		return nil, err
	}
	return outBuf.Bytes(), nil
}
