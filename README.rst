
Cas status and todo
===================

New
---

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

    * is the share valid (expiration, password, # uses, etc.)

    * is the attempted operation within the share's privelages (get vs
      put, if get - is ref authorized, etc.)

  - a share blob will help the blobserver know if the requested get/put
    blob operation is allowed

* cli tools

  - rework mount/modify/rm toolchain to be as follows:

    - A tool fad-stat to stat meta-data of piped in refs (e.g. return their timestamp, objectref,
      etc.)

    - Use examples::

        // mount blobs that have a certain path into direcory mymount with
        // /home/robert prefix removed
        fad-find -path=/home/robert | fad-mount -root=mymount -prefix=/home/robert 
        // modify meta-data on a file: mark 'hidden' field as true and add
        // 'failure' to a list of tags
        fad-ref foo.txt | fad-stat mymount/foo.txt 
        fad-ref foo.txt | fad-mod hidden=true tag+=failure

  - creating share blobs corresponding to mounted files::

      fad-share ??????

In Progress
-----------

preliminarily done
------------------

- use https (TLS) instead of http on the blobserver

- move blob.Meta Hidden field to be inside the filemeta Notes

- move Meta.Path into the Notes field somehow

  - Modify fad-mount to take list of blobrefs (not object refs) to mount piped in

- A generic fad-find tool that returns a list of blobrefs from a blobserver

  * based on path, tags, or other arbitrary meta-data

- A tool fad-mod that modifies arbitrary Notes meta-data of listed or
  piped file names. - so far only modifies mount meta-data

- kill fad-rm tool (superseded by fad-ref and fad-mod)

* things affecting blob schemas

  - rename blob.FileMeta to just Meta and make its RcasType be blob.MetaType

decided against
---------------

- What if somebody wants an app that deals with metablobs that don't point to
  any files? I want them to still use the standard Meta concept where
  many apps can each pile their own meta-data into a single file (the Notes
  field).

  * maybe I could put the ContentRefs inside the Notes field? - interesting
    idea eh? - S

  * why I don't like it:

    Actually - I would like the meta-blobs to really stay meta-blobs.
    If somebody has real, hardy, application data, it should go in
    separate blobs referenced in ContentRefs.

- cli tool for mounting/inspecting an object's history. Just use fad-find,
  fad-ref, fad-mount instead.

- A cli tool fad-ref that returns the object ref for the mounted file -
      doesn't fit well with mount tools. Scrapped.

