# tool - list, count, and export the entries of a directory

## BLUF

tool lists and counts the entries of a directory.
The list and count commands work.
Export to a CSV file comes next.

## Usage

    tool list DIR
    tool count DIR

## Description

`tool list` prints one entry per line, sorted.
`tool count` prints the number of entries.
`tool export` writes the entries with their sizes to a CSV file.

## Options

`--help`
: Print the usage and exit.

## Exit status

The exit status is 0 on success, and 1 when the directory cannot be read.
