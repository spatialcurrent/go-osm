package xmlutil

import (
	"bufio"
	"encoding/xml"
	"os"
)

import (
  "github.com/pkg/errors"
)

func WriteToPlainFile(output_file *os.File, output_bytes []byte) error {
  w := bufio.NewWriter(output_file)

  _, err := w.WriteString(xml.Header)
  if err != nil {
		return errors.Wrap(err, "Error writing XML Header to xml file")
  }

  _, err = w.Write(output_bytes)
  if err != nil {
		return errors.Wrap(err, "Error writing string to xml file")
  }

  _, err = w.WriteString("\n")
  if err != nil {
		return errors.Wrap(err, "Error writing last newline to xml file")
  }

  w.Flush()
  if err != nil {
		return errors.Wrap(err, "Error flushing output to bufio writer for xml file")
  }

  err = output_file.Close()
  if err != nil {
		return errors.Wrap(err, "Error closing file writer for xml file.")
  }

	return nil
}
