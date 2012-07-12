
package main

import (
  "fmt"
  "time"
  "github.com/rwcarlsen/cas/blob"
  "github.com/rwcarlsen/cas/blobdb"
)



func main() {
  //testRaw()
  //testFile()
  //testDir()
  //testQuery()
  testTimeIndex()
}

func testRaw() {
  b := blob.NewRaw([]byte("hello monkey man"))
  db, _ := blobdb.New(".")

  err := db.Put(b)
  if err != nil {
    fmt.Println(err)
    return
  }

  b2, err := db.Get(b.Ref())
  if err != nil {
    fmt.Println(err)
    return
  }

  fmt.Println(b)
  fmt.Println(b2)
}

func testFile() {
  db, _ := blobdb.New(".")

  meta := blob.NewFileMeta()
  blobs, err := meta.LoadFromPath("foodir/foo.txt")
  if err != nil {
    fmt.Println(err)
    return
  }

  err = db.Put(blobs...)
  if err != nil {
    fmt.Println(err)
    return
  }

  m, _ := blob.Marshal(meta)
  err = db.Put(m)
  if err != nil {
    fmt.Println(err)
    return
  }

  for _, b := range blobs {
    fmt.Println(b)
  }
}

func testDir() {
  db, _ := blobdb.New(".")

  metas, blobs, err := blob.DirBlobsAndMeta("foodir")
  if err != nil {
    fmt.Println(err)
    return
  }

  metablobs := make([]*blob.Blob, 0)
  for _, meta := range metas {
    m, _ := blob.Marshal(meta)
    metablobs = append(metablobs, m)
  }

  err = db.Put(metablobs...)
  if err != nil {
    fmt.Println(err)
    return
  }

  err = db.Put(blobs...)
  if err != nil {
    fmt.Println(err)
    return
  }

  for _, b := range metablobs {
    fmt.Println(b)
  }
  for _, b := range blobs {
    fmt.Println(b)
  }
}

func testQuery() {
  b1 := blob.NewRaw([]byte("I am not json"))
  b2 := blob.NewRaw([]byte("{\"key\":\"I am wrong json\"}"))
  b3 := blob.NewRaw([]byte("{\"key\":\"I am right json\"}"))

  q := blobdb.NewQuery()

  isjson := q.NewFilter(blobdb.IsJson)
  right := q.NewFilter(blobdb.Contains("right"))

  isjson.SendTo(right)
  q.SetRoots(isjson)

  q.Open()
  defer q.Close()

  q.Process(b1, b2, b3)

  fmt.Println("results: ", q.Results)
}

func testTimeIndex() {
  ti := blobdb.NewTimeIndex()

  m1 := make(blob.MetaData)
  m2 := make(blob.MetaData)
  m3 := make(blob.MetaData)
  m4 := make(blob.MetaData)
  m5 := make(blob.MetaData)

  b1, _ := blob.Marshal(m1)
  time.Sleep(time.Second * 1)
  b2, _ := blob.Marshal(m2)
  time.Sleep(time.Second * 1)
  b3, _ := blob.Marshal(m3)
  time.Sleep(time.Second * 1)
  b4, _ := blob.Marshal(m4)
  time.Sleep(time.Second * 1)
  b5, _ := blob.Marshal(m5)

  ti.Notify(b1, b2, b3, b4, b5)

  var m blob.MetaData
  blob.Unmarshal(b3, &m)
  t, _ := time.Parse(blob.TimeFormat, m[blob.TimeField].(string))

  i := ti.IndexOf(t)
  ref := ti.GetRef(i)

  if ref == b3.Ref() {
    fmt.Println("success!")
  } else {
    fmt.Println("retrieved ref:", ref)

    fmt.Println("all refs:")
    fmt.Println(b1.Ref())
    fmt.Println(b2.Ref())
    fmt.Println(b3.Ref())
    fmt.Println(b4.Ref())
    fmt.Println(b5.Ref())
  }
}

