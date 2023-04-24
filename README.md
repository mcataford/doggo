# doggo
Inspecting big Datadog traces in the CLI

## Why?

The Datadog trace-viewing UI performance greatly degrades as the number of spans in a single trace increases and within
a trace, it is often painful to find the span you are interested in as there is no way to "jump" to a span directly.

This is an attempt at solving that problem: using `doggo`, you can inspect large trace json blobs and get to the subtree
corresponding to spans of interest in a few keystrokes without having to worry about your browser running out of memory.
At present, you will have to have the json you want to inspect handy -- the tool simply makes it more palatable.
