# op2pass

A quick hack to import 1Password items as provided from `op get item` into
`pass` ([password store](https://www.passwordstore.org/)).

Items are named as `/website-domain/username`. Content is the password, unless other
fields have been discovered in which case it is made a multiline entry.

No support for OTP or anything fancy.

## Usage

Run `go install` from the this repo then run `op2pass <file.json>` where
`file.json` is the result of doing `op get item`.

If you wish to import the whole 1Password vault, run the following (these are
for the `fish` shell, `bash` is similar):

```fish
$ mkdir /dev/shm/op
$ cd /dev/shm/op
$ op list items | jq -r '.[] | .uuid' > uuids.txt
$ for uuid in (cat uuids.txt); echo $uuid; op get item $uuid | jq > $uuid.json; end
$ for f in (ls *.json); echo $f; op2pass $f; mv $f $f.done; end
```

Whenever there are several options for the `website-domain` part, you will be
interactively asked to select one. Options come from the item Title and URL
fields.

## License

GPL 3.0
