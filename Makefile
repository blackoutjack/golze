.PHONY: test clean

VENV := .venv
PYTHON := $(VENV)/bin/python
PIP := $(VENV)/bin/pip
TESTREQ := $(VENV)/.installed

golze: *.go
	go build

$(PIP):
	python3 -m venv $(VENV)

$(TESTREQ): $(PIP) requirements.txt
	$(PIP) install -r requirements.txt
	touch $(TESTREQ)

test: $(TESTREQ) golze
	$(VENV)/bin/python -m test

clean:
	rm -rf $(VENV) golze

