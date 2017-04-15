BINARY=BreizhCamp2017programBuilder

BUILD=`date +%Y-%m-%d/%H:%M:%S`
LDFLAGS=-ldflags "-w -s -X main.Build=${BUILD}"

build:
	go build ${LDFLAGS} -o ${BINARY}

test: build
	go test -v

testrun: clean build test 
	./${BINARY}

install: test
	go install ${LDFLAGS}

clean:
	if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi

.PHONY: clean install
