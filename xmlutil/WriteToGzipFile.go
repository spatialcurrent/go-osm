package xmlutil

import (
	"bufio"
	"encoding/xml"
	"compress/gzip"
	"os"
)

import (
  "github.com/pkg/errors"
)

func WriteToGzipFile(output_file *os.File, output_bytes []byte) error {

  gw := gzip.NewWriter(output_file)
  w := bufio.NewWriter(gw)
  _, err := w.WriteString(xml.Header)
  if err != nil {
    return errors.Wrap(err, "Error writing XML Header to xml file")
  }

  _, err = w.Write(output_bytes)
  if err != nil {
    return errors.Wrap(err, "Error writing string to gzip writer for xml file")
  }

  _, err = w.WriteString("\n")
  if err != nil {
    return errors.Wrap(err, "Error writing last newline to gzip writer for xml file")
  }

  err = w.Flush()
  if err != nil {
    return errors.Wrap(err, "Error flushing output to bufio writer for xml file.")
  }

  err = gw.Flush()
  if err != nil {
    return errors.Wrap(err, "Error flushing output to gzip writer for xml file.")
  }

  err = gw.Close()
  if err != nil {
    return errors.Wrap(err, "Error closing gzip writer for xml file.")
  }

  err = output_file.Close()
  if err != nil {
    return errors.Wrap(err, "Error closing file writer for xml file.")
  }

  return nil
}
