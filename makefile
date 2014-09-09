amd64:
	go build -o coordnode_amd64

arm:
	docker run -v /home/vagrant/go:/root/go -w="/root/go/src/github.com/herrfz/coordnode" \
	-e PATH=$$PATH:/opt/go/bin -e GOPATH=/root/go -e CC=arm-linux-gnueabi-gcc -e GOARM=7 \
	-e GOARCH=arm -e GOOS=linux -e CGO_ENABLED=1 herrfz/armbuilder go build -o coordnode_arm

i386:
	docker run -v /home/vagrant/go:/root/go -w="/root/go/src/github.com/herrfz/coordnode" \
	-e PATH=$$PATH:/opt/go/bin -e GOPATH=/root/go -e GOARCH=386 -e GOOS=linux \
	-e CGO_ENABLED=1 herrfz/i386builder go build -o coordnode_i386

all: amd64 arm i386

install:
	go install

clean:
	go clean -i ; rm -f coordnode_*
