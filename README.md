# Golze

Golze is a proof-of-concept API discovery dataflow analysis written on top of
Go's built-in dataflow analysis engine.

## Getting started

```
# Clone/build the repo
git clone https://github.com/blackoutjack/golze.git
cd golze && go build

# Install the test framework
python -m venv ./.venv/golze
. .venv/golze/bin/activate
pip install -r requirements.txt

# Run the tests
python -m test
```

## Current support

This is a work in progress. Only the following endpoint registration methods are supported:
- gorilla/mux.HandleFunc
Parameter discovery is based on observation of the following methods.
- encoding/json.Decoder.Decode
Stay tuned for more.
