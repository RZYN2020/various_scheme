# Makefile for Scheme Interpreter

# Compiler and flags
CC = gcc
CFLAGS = -Wall -Wextra -g -I./cscheme

SOURCE_DIR = ./ss_c
SRCS = $(SOURCE_DIR)/main.c $(SOURCE_DIR)/eval.c $(SOURCE_DIR)/prim.c $(SOURCE_DIR)/parse.c $(SOURCE_DIR)/data.c

# Target executable
TARGET = $(SOURCE_DIR)/cscheme

# Default target
all: $(TARGET)

# Build the executable
$(TARGET): $(SRCS)
	$(CC) $(CFLAGS) -o $@ $^

# Clean up
clean:
	rm -f $(TARGET)

# Run the interpreter
run: $(TARGET)
	./$(TARGET)

.PHONY: all clean run