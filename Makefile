all: build

build:
	go build

# For wsc, we need to build wcs.exe before testing because it
# will be run during the integration tests.
test: build
	go test
	
clean:
	go clean
	rm -f test_output.txt

demo: build
	wsc --dir=test_www ruby test.rb


