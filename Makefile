ldapsync_SOURCES = \
	*.go

all: ldapsync

ldapsync: $(ldapsync_SOURCES)
	go build -x -o mgr-ldapsync $(ldapsync_SOURCES)

clean:
	go clean -x -i

.PHONY: all install clean
