package xmlutil

import (
  "encoding/xml"
  "fmt"
)

func WriteToStdout(output_bytes []byte) {
  fmt.Println(xml.Header)
  fmt.Println(string(output_bytes))
}
