
package main

import (
  "fmt"
  "time"
  "os"
  "log"
  //"path/filepath"
  "github.com/rwcarlsen/cas/blob"
  "github.com/rwcarlsen/cas/blobdb"
  "github.com/rwcarlsen/cas/blobserv"
  "github.com/rwcarlsen/cas/query"
  "github.com/rwcarlsen/cas/blobserv/timeindex"
)

var (
  home string = os.Getenv("HOME")
  dbpath = "./testdb" //blobserv.DefaultDb

  testdirpath = "./foodir"
  testfilepath = "./foodir/foo.txt"
)


func main() {
  //testRaw()
  //testFile()
  testDir()
  //testQuery()
  //testTimeIndex()
  testBlobServer()
}

func testRaw() {
  db, err := blobdb.New(dbpath)
  if err != nil {
    fmt.Println(err)
    return
  }

  b := blob.NewRaw([]byte("hello monkey man"))
  err = db.Put(b)
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
  db, err := blobdb.New(dbpath)
  if err != nil {
    fmt.Println(err)
    return
  }

  meta := blob.NewFileMeta()
  blobs, err := meta.LoadFromPath(testfilepath)
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
  db, _ := blobdb.New(dbpath)

  metas, blobs, err := blob.DirBlobsAndMeta(testdirpath)
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

  q := query.New()

  isjson := q.NewFilter(query.IsJson)
  right := q.NewFilter(query.Contains("right"))

  isjson.SendTo(right)
  q.SetRoots(isjson)

  q.Open()
  defer q.Close()

  q.Process(b1, b2, b3)

  fmt.Println("results: ", q.Results)
}

func testTimeIndex() {
  ti := timeindex.New()

  m1 := map[string]string{}
  m2 := map[string]string{}
  m3 := map[string]string{}
  m4 := map[string]string{}
  m5 := map[string]string{}

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

  var m map[string]string
  blob.Unmarshal(b4, &m)
  t, _ := time.Parse(blob.TimeFormat, m[blob.Timestamp])

  i := ti.IndexNear(t.Add(time.Millisecond * -1))
  ref := ti.RefAt(i)

  fmt.Println("retrieved ref:", ref)

  fmt.Println("all refs:")
  fmt.Println(b1.Ref())
  fmt.Println(b2.Ref())
  fmt.Println(b3.Ref())
  fmt.Println(b4.Ref())
  fmt.Println(b5.Ref())

  if ref == b3.Ref() {
    fmt.Println("success!")
  } else {
    fmt.Println("failured")
  }
}

func testBlobServer() {
  fmt.Println("running blob server...")
  log.Fatal(blobserv.ListenAndServe(blobserv.DefaultAddr, dbpath))
}

