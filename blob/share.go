
package blob

type AuthType string

const (
  HaveRef AuthType = "have-ref"
)

type ShareMeta struct {
  RcasType string "rcasType"
  Auth AuthType "AuthType"
  Target string
}

// NewFileMeta creates a map containing meta-data for a file
// at the specified path.
func NewShareMeta() *ShareMeta {
  return &ShareMeta{
    RcasType: Share,
  }
}

