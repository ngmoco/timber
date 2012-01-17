include $(GOROOT)/src/Make.inc

TARG=timber
GOFILES=\
	timber.go\
	pattern_formatter.go\
	console_writer.go\
	buffered_writer.go\
	socket_writer.go\
	file_writer.go\
	xml_config.go\

include $(GOROOT)/src/Make.pkg

