CC=gcc
CFLAGS= -Wall

all: main

main: main.o scanner.o
	$(CC) $(CFLAGS) main.c scanner.c -o main