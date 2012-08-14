
package blob

const (
  HaveRef = "have-ref"
)

type ShareMeta struct {
  RcasType string
  AuthType string
  Target string
}

// NewMeta creates a map containing meta-data for a file
// at the specified path.
func NewShareMeta() *ShareMeta {
  return &ShareMeta{
    RcasType: Share,
  }
}

func (sh *ShareMeta) IsAuthorized() bool {
  return true
}
