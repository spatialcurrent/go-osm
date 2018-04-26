package xmlutil

import (
  "strings"
  "os"
)

import (
  "github.com/pkg/errors"
)

func WriteToDisk(output_path string, output_bytes []byte) error {

  output_file, err := os.OpenFile(output_path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
  if err != nil {
    return errors.Wrap(err, "error opening file to write bytes to disk")
  }

  if strings.HasSuffix(output_path, ".gz") {
    return WriteToGzipFile(output_file, output_bytes)
  }

  if strings.HasSuffix(output_path, ".bz2") {
    return errors.New("Golang does not support writing to bzip2 files")
  }

  return WriteToPlainFile(output_file, output_bytes)
}
