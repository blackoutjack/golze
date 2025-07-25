
run_1 = ["test/e2e/basic", "-vettool=../../../golze", "./..."]

err_1 = """
# basic/handlers
===== beginning discovery: endpoints@basic/handlers =====
handlers/content.go:21:46: inferred parameters for json.Decoder.Decode: [json:"kind" json:"records"]
# basic
# [basic]
===== beginning discovery: endpoints@basic =====
./main.go:17:17: gorilla/mux endpoint detected: "/":string
./main.go:18:17: gorilla/mux endpoint detected: "/home":string
./main.go:19:17: gorilla/mux endpoint detected: "/content":string
"""

code_1 = 1
