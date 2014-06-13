#!/bin/sh
../../netroute -verbose -ssl -delim "\r\n" irc.offblast.org 6697 &
go run ./irc_client.go &
