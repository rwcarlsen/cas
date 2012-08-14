
package blob

type Share struct {
  RcasType string
  TargetRefs []string
  Auth *Authorization
}

type Authorization struct {
  StaticGet bool // only access refs explicitly listed in TargetRefs
  DynamicGet bool// access any ref that is part of shared object ref
  DynamicPut bool// true allows users to store new versions of an object
}

// NewShare creates a map containing meta-data for a file
// at the specified path.
func NewShare() *Share {
  return &Share{
    RcasType: ShareType,
    Auth: &Authorization{},
  }
}

// AuthorizedGet returns true if this share allows retrieval of b
func (sh *Share) AuthorizedGet(b *Blob) bool {
  good := false

  if sh.Auth.StaticGet {
    good = sh.have(b.Ref())
  } else if sh.Auth.DynamicGet {
    good = sh.have(b.ObjectRef())
  }

  return good
}

// AuthorizedPut returns true if this share allows attaching new versions
// to object identified by b's RcasObjectRef.
func (sh *Share) AuthorizedPut(b *Blob) bool {
  if sh.Auth.DynamicPut {
    return sh.have(b.ObjectRef())
  }
  return false
}

func (sh *Share) have(ref string) bool {
  for _, targ := range sh.TargetRefs {
    if targ == ref {
      return true
    }
  }
  return false
}

