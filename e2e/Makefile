ictest-poa:
	go clean -testcache && go test -race -v -run TestPOABase .

ictest-jail:
	go clean -testcache && go test -race -v -run TestPOAJailing .

ictest-val-add:
	go clean -testcache && go test -race -v -run TestPOAAddValidator .

ictest-poa-to-pos:
	go clean -testcache && go test -race -v -run TestPoAToPoSUpgrade .

ictest-val-remove:
	go clean -testcache && go test -race -v -run TestPOARemoval .
