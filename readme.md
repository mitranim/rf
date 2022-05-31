## Overview

Short for "**r**e**f**lect utils": utilities missing from the "reflect" package in the Go standard library. Small and dependency-free.

See the full documentation at https://pkg.go.dev/github.com/mitranim/rf.

## Changelog

### v0.5.0

Converted the following tools to generics, for better type safety and efficiency:

* `TypeFilter`
* `IfaceFilter`
* `ShallowIfaceFilter`
* `Appender`
* `Trawl`
* `TrawlWith`

Added `KindFilter` (untested).

Added `Type`.

### v0.4.2

Rename `WalkFuncPtr` → `WalkPtrFunc`.

### v0.4.1

Add `WalkPtr` and `WalkFuncPtr`. Remove `DerefPtr` because it made no sense.

### v0.4.0

Added:

  * `ValueAddr`
  * `Interface`
  * `DeepFields`
  * `TypeDeepFields`
  * `OffsetFields`
  * `TypeOffsetFields`
  * `Path`
  * `Rev`

Breaking changes:

  * Renamed all functions starting with `DerefValue*` to begin with `Deref*`.
  * Replaced `CopyPath` with `Path.Copy`.

`DeepFields` and `TypeDeepFields` is a particularly useful addition, as it supports "flattening" structs, simplifying most use cases that involve struct iteration.

### v0.3.3

Add `IfaceFilterFor`, `ShallowIfaceFilterFor`.

### v0.3.2

Reverted breaking change in `v0.3.1`: `IfaceFilter` once again allows to visit descendants. Added `ShallowIfaceFilter` that doesn't visit descendants of a matching node.

### v0.3.1

Quick breaking change: `IfaceFilter` visits either self or descendants, not both.

### v0.3.0

More flexible `Filter` interface:

  * Previous approach: filter returns `bool` answering "should visit this node". 2 possible states. Implicitly walks descendants.

  * New approach: filter returns flagset where "should visit this node" and "should walk descendants" are both optional flags. 4 possible states. Walking descendants is now optional.

Renamed `Filter.ShouldVisit` to `Filter.Visit` because it's no longer a boolean. This makes it impossible to implement `Filter` and `Walker` on the same type, which is probably a good thing due to filter equality rules.

`Nop` no longer implements `Filter`.

Replaced `True` and `False` with `Self`, `Desc`, `Both`, `All`.

Replaced `Not` with `InvertSelf`.

Renamed `DerefLen` to `Len`.

### v0.2.2

Added `Fields` and `TypeFields` for micro-optimizing struct shallow walking.

### v0.2.1

`Walk` / `GetWalker` now support walking into `interface{}` values, fetching the appropriate cached walker for the given type and filter on the fly.

Added `MaybeOr`, `MaybeAnd`, `GetTypeFilter`, `TypeFilterFor` for micro-optimizing filter allocations.

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
