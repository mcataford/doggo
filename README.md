# doggo
🐕🔎 Inspecting big Datadog traces in the CLI

## Why?

The Datadog trace-viewing UI performance greatly degrades as the number of spans in a single trace increases and within
a trace, it is often painful to find the span you are interested in as there is no way to "jump" to a span directly.

This is an attempt at solving that problem: using `doggo`, you can inspect large trace json blobs and get to the subtree
corresponding to spans of interest in a few keystrokes without having to worry about your browser running out of memory.
At present, you will have to have the json you want to inspect handy -- the tool simply makes it more palatable.

## Installation

### Pre-built binaries

Each release includes pre-built binaries, drop those in your path (i.e. in `/usr/local/bin`) and you are ready to go!

### Building your own

You can also clone this repository and `go build .` to get your own executable.

## Usage

To get `doggo` on your machine, you can grab a [pre-compiled release](https://github.com/mcataford/doggo/releases) for your machine if one is available or build it locally by cloning the repository and running `go build`.

You can either include the executable in your `$PATH` or run the executable wherever it lives:

```
doggo <path-to-trace-json> <resource-name> [-vv] [--depth={depth-limit}]
```

The `resource-name` provided is the resource name associated with the spans that interest you in the provided trace.
Doggo will display all the span subtrees up to `depth-limit` depth (unlimited depth if not specified) that have a
resource name that matches your query (either complete or partial).

Default verbosity will include resource names and span duration in millis. Adding `-v` will toss in some extra
information about each resource and `-vv` will include any metadata (i.e. tagging) that is available in the trace for
that span.

## Contributing

Doggo welcomes contributions. The main use case is pretty tailored to my day-to-day, but sensible suggested changes are
welcome!
