## Overview

Short for "**ref**lect **ut**ils": utilities missing from the "reflect" package in the Go standard library. Small and dependency-free.

See the full documentation at https://godoc.org/github.com/mitranim/refut.

## Changelog

### 0.1.1

* Bugfix: `TraverseStruct` and `TraverseStructRval` no longer attempt to traverse nil embedded struct pointers.
* `TraverseStruct` and `TraverseStructRval` now allow a nil struct pointer as input, without traversing its fields. This behavior is consistent with nil embedded struct pointers.
* Added `RvalDerefAlloc`.
* Added `RvalFieldByPathAlloc`.

### 0.1.0

First tagged release.

## License

https://unlicense.org

## Misc

I'm receptive to suggestions. If this library _almost_ satisfies you but needs changes, open an issue or chat me up. Contacts: https://mitranim.com/#contacts
