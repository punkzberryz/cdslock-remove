build-linux:
	set GOOS=linux& set GOARCH=amd64& go build -o ./bin/cdslock-remove
build-linux-old: 
	#Run this if you are building the file in WSL-Ubuntu.
	#This will fource Go to use pure Go implementations instead of linking to C libraries, because
	#some old linux machine may be using mismatch library
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/cdslock-remove
