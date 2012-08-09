
Todo
----

* things affecting blob schemas

  - move blob.FileMeta Hidden field to be inside the filemeta Notes

  - consider moving FileMeta.Path into the Notes field somehow

  - rename FileMeta to just Meta and make its RcasType be blob.MetaType

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

* make cli tools

  - rework mount/modify/rm toolchain to be as follows:

    - kill fad-rm tool

    - fad-snap should work as it currently does

    - A generic fad-find tool that returns a list of blobrefs from a blobserver

       * based on path, tags, or other arbitrary meta-data

    - A tool fad-mod that modifies arbitrary Notes meta-data of piped in object refs

    - Modify fad-mount to take list of blobrefs (not object refs) to mount piped in

    - A tool (fad-stat) to stat mounted files (e.g. return their timestamp, objectref,
      etc.)

    - Use examples::

        // mount blobs that have a certain path into direcory mymount with
        // /home/robert prefix removed
        fad-find -path=/home/robert | fad-mount -root=mymount -prefix=/home/robert

        // modify meta-data on a file: mark 'hidden' field as true and add
        // 'failure' to a list of tags
        fad-stat mymount/foo.txt | fad-mod hidden=true tag+=failure

  - mounting/inspecting an object's history::

      fad-hist ??????

  - creating share blobs corresponding to mounted files::

      fad-share ??????

decided against
---------------

  - What if somebody wants an app that deals with metablobs that don't point to
    any files? I want them to still use the standard FileMeta concept where
    many apps can each pile their own meta-data into a single file (the Notes
    field).

      * maybe I could put the ContentRefs inside the Notes field? - interesting
        idea eh? - S

      * why I don't like it:

          Actually - I would like the meta-blobs to really stay meta-blobs.
          If somebody has real, hardy, application data, it should go in
          separate blobs referenced in ContentRefs.

preliminarily done
------------------

  - use https (TLS) instead of http on the blobserver


