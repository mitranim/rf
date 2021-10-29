## Overview

Short for "**r**e**f**lect utils": utilities missing from the "reflect" package in the Go standard library. Small and dependency-free.

See the full documentation at https://pkg.go.dev/github.com/mitranim/rf.

## Changelog

### v0.2.0

Complete revision.

* Removed useless or rarely-used utils.
* Added many missing utils.
* New approach to walking / traversal. The old naive approach walked the entire structure every time. The new approach is to JIT-compile a precise walker that visits just what you need, caching it for a combination of type + filter. This makes walking dramatically more efficient.
* Added `Cache` for generating and caching arbitrary type-dependent structures.
* Renamed from `github.com/mitranim/refut` to `github.com/mitranim/rf` for brevity.

## License

https://unlicense.org

## Misc

I'm receptive to suggestions. If this library _almost_ satisfies you but needs changes, open an issue or chat me up. Contacts: https://mitranim.com/#contacts
