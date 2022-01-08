package wcache

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
)

// Encoding interface
type Encoding interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
}

type JSONEncoding struct{}

func (JSONEncoding) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (JSONEncoding) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

type JSONGzipEncoding struct{}

func (JSONGzipEncoding) Marshal(v interface{}) ([]byte, error) {
	buf := &bytes.Buffer{}
	writer, err := gzip.NewWriterLevel(buf, gzip.BestCompression)
	if err != nil {
		return nil, err
	}
	err = json.NewEncoder(writer).Encode(v)
	if err != nil {
		writer.Close()
		return nil, err
	}
	writer.Close()
	return buf.Bytes(), nil
}

func (JSONGzipEncoding) Unmarshal(data []byte, v interface{}) error {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer func() {
		reader.Close()
	}()
	return json.NewDecoder(reader).Decode(v)
}
