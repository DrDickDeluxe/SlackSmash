PROJECT_NAME := ss

SRCDIR := src
BINDIR := bin

SRCS := $(wildcard $(SRCDIR)/*.go)
CC := "/usr/local/go/bin/go"

TARGET := $(BINDIR)/$(PROJECT_NAME)

all: $(TARGET) $(GOPATH)/pkg/linux_amd64/golang.org/x/net/proxy.a

clean:
	rm -f $(TARGET)

$(GOPATH):
	$(error "GOPATH not defined in environment")

$(GOPATH)/pkg/linux_amd64/golang.org/x/net/proxy.a: $(GOPATH)
	@echo "You don't have Go proxy. Fetching automagically"
	$(CC) get golang.org/x/net/proxy

$(TARGET): $(SRCS)
	@mkdir -p $(shell dirname $@)
	$(CC) build -o $@ $^

.PHONY: all clean
