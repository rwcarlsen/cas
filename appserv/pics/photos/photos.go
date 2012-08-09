
package photos

import (
  "github.com/rwcarlsen/cas/blob"
  "strings"
  "path"
)

type Photo struct {
  Who []string
  Tags []string
  Exif map[string]string
  ThumbFileRef string
}

func IsValidImage(m *blob.FileMeta) bool {
  if m.RcasType != blob.File {
    return false
  }
  switch strings.ToLower(path.Ext(m.Name)) {
    case ".jpg", ".jpeg", ".gif", ".png": return true
  }
  return false
}
