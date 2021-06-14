# This file includes a shim that will execute your task code.

import importlib.util as util
import json
import sys
import os

def run(args):
	if len(args) != 3:
		raise Exception("shim: expected 3 arguments, got {}".format(args))

	spec = util.spec_from_file_location("mod.main", "{{ .Entrypoint }}")
	mod = util.module_from_spec(spec)
	spec.loader.exec_module(mod)

	try:
		mod.main(json.loads(args[2]))
	except Exception as e:
		raise Exception("shim: executing {{.Entrypoint}}") from e

if __name__ == "__main__":
	run(sys.argv)
