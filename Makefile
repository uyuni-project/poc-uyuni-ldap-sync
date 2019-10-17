ldapsync_SOURCES = \
	*.go

all: ldapsync

ldapsync: $(ldapsync_SOURCES)
	go build -x -o uyuni-ldapsync $(ldapsync_SOURCES)

clean:
	go clean -x -i

.PHONY: all install clean
