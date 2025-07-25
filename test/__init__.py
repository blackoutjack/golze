
from rjtools.util.testing import run_modules

COMMAND_PREFIX = [
    "go", "vet", "-C",
]

def run():
    from . import analysis

    return run_modules("golze", locals(), COMMAND_PREFIX)

