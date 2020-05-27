default:
	go vet .
	for i in *; do \
	  test ! -d $$i || (echo $$i && cd $$i && go vet . && cd ..); \
	done;
