package xmlutil

import (
  "encoding/xml"
  "fmt"
  "os"
)

func WriteToStderr(output_bytes []byte) {
  fmt.Fprintf(os.Stderr, xml.Header)
  fmt.Fprintf(os.Stderr, string(output_bytes))
}
