*Fork Notes: This fork includes patches which makes decoding more robust with
respect to non-standard log values including values wrapped in single quotes
and other invalid characters. Upstream merge discussion can be
viewed in https://github.com/go-logfmt/logfmt/pull/5*

# logfmt

Package logfmt implements utilities to marshal and unmarshal data in the [logfmt
format](https://brandur.org/logfmt). It provides an API similar to
[encoding/json](http://golang.org/pkg/encoding/json/) and
[encoding/xml](http://golang.org/pkg/encoding/xml/).

The logfmt format was first documented by Brandur Leach in [this
article](https://brandur.org/logfmt). The format has not been formally
standardized. The most authoritative public specification to date has been the
documentation of a Go Language [package](http://godoc.org/github.com/kr/logfmt)
written by Blake Mizerany and Keith Rarick.

## Goals

This project attempts to conform as closely as possible to the prior art, while
also removing ambiguity where necessary to provide well behaved encoder and
decoder implementations.

## Non-goals

This project does not attempt to formally standardize the logfmt format. In the
event that logfmt is standardized this project would take conforming to the
standard as a goal.

## Versioning

Package logfmt publishes releases via [semver](http://semver.org/) compatible Git tags prefixed with a single 'v'.
