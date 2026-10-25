# tool - list and count the entries of a directory

tool is a small command-line utility that lists directory entries.

## Usage

    tool list DIR
    tool count DIR

## Description

`tool list` prints one entry per line, sorted.
`tool count` prints the number of entries.
`tool export DIR` writes the entries to a CSV file.

## Exit status

The exit status is 0 on success, and 1 when the directory cannot be read.
