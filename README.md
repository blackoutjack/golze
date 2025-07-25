# Golze

Golze is a proof-of-concept API discovery dataflow analysis written on top of
Go's built-in dataflow analysis engine.

## Getting started

```
# Clone/build the repo
git clone https://github.com/blackoutjack/golze.git
cd golze

# Install and run the test framework
make test
```

## Current support

This is a work in progress. Only the following endpoint registration methods are supported:
- `gorilla/mux.HandleFunc`

Parameter discovery is based on observation of the following methods:
- `encoding/json.Decoder.Decode`

Stay tuned for more.
