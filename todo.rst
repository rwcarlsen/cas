
* Security/authentication

  - way to manage multiple logins (maybe through share blobs)

  - allow rsa (and others?) public key authentication for blobserver
    authentication

* sharing - decide how

  - use the via technique on the standard blobserver '/ref/' path. The
    blobserver will then decide if the via allows access to the requested
    operation (get/put a blob). Possibly requesting password/rsa signing,
    etc. e.g.::
      
      http://blobhost/ref/[hash-of-requested-blob]?via=[hash-of-share-blob]

  - authentication requires two concepts:

    * is the attempted operation within the 

  - a share blob will help the blobserver know if the requested get/put
    blob operation is allowed

* things affecting blob schemas

  - move blob.FileMeta Hidden field to be inside the filemeta Notes

* make cli tools

  - tagging mounted files with arbitrary meta-data - e.g.::

      fad-tag [filename(s)] // list existing tags
      fad-tag [tag1[,tag2]...] [filename(s)] // add new tags to files
      fad-tag -d [tag1[,tag2]...] [filename(s)] // remove existing tags

  - features to add to fad-mount::

      fad-mount -t=tag1,tag2 ... // adds additional constraint to mounting

      // mounts all matching file blobs into the root directory
      // mount fails if two files have the same name (or maybe it changes
      // name to be unique?)
      fad-mount -[f,flat] ... 

  - mounting/inspecting an object's history::

      fad-hist ??????

  - creating share blobs corresponding to mounted files::

      fad-share ??????

* preliminarily done:

  - use https (TLS) instead of http on the blobserver


