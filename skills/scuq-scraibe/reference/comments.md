# Code comments by language

## BLUF

This file gives the comment rules for Go, Python, Bash, and .NET.
The common rules are in section 6 of the standard, and they apply here too.
Read the part for the language before you write or change a comment.

## Go

Follow go.dev/doc/comment. Run `gofmt` after every change.

- Every exported name has a doc comment. The comment starts with the
  name: "Package netbox ...", "A Client ...", "Fetch returns ...". <!-- ste-lint: disable-line synonym-rotation -->
- Command: the package comment starts with the program name and
  describes the behavior. Then a `Usage:` block, indented.
- Func: say what it returns, or what it does when called for the
  side effect. Boolean result: "reports whether". Name the results
  when the comment refers to them.
- Type: say what one instance represents. State concurrency safety
  if it is stronger than "one goroutine at a time". State the zero
  value meaning if it is not obvious. Document every exported field,
  in the type comment or per field.
- Special cases: list them explicitly, one per line, indented.
- Do not describe the algorithm in the doc comment. Put that inside
  the function body.
- Deprecation: a paragraph `Deprecated: <reason>. Use <X> instead.`
- Links: `[Name]`, `[pkg.Name]`, `[*pkg.Type]`. URL targets go at the
  end as `[Text]: URL`.
- Syntax: blank `//` line between paragraphs. `# Heading` unindented
  with blank lines around it. Lists indented with `-` or `1.`. Code
  blocks indented one tab. No nested lists. Do not indent a wrapped
  prose line.

## Python

- Docstring on every public module, class, and function.
- First line: one sentence, imperative, ends with a period.
  Blank line. Then details.
- Sections `Args:`, `Returns:`, `Raises:` when the function has them.

## Bash

- Header block: name, one-line purpose, usage line, exit codes.
- One comment line before each function: what it does.

## .NET

- `<summary>` is one sentence. Add `<param>`, `<returns>`,
  `<exception>` when they exist.
