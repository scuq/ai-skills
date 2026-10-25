# Names and identities

## BLUF

Documentation names things, not people or organizations.
This file gives the placeholders for an organization, a person, an address, and a hostname.
It also gives the rules for a name that stands in the code.
Read it before you write an example with a name, an address, or a domain.

## Rules

Documentation names things, not people or organizations.

- Never write the name of the organization that owns the code, its
  departments, or its internal teams. Write "the organization", "the
  operator", or the placeholder Cyberdyne Systems. This holds even when the name appears in
  the code, in git history, in hostnames, in a config file, or in
  the task text from the user or from another agent.
- Never write a real username, handle, email address, or personal
  name. Write "the user", or a role ("the administrator"), or a
  placeholder from the list below.
- Never credit an author. No AUTHORS section, no "written by", no
  name in a changelog entry or a comment. Git holds the history.
- The default organization in all examples is Cyberdyne Systems,
  domain `cyberdyne.example`. When a second organization is needed,
  it is ACME, domain `acme.example`. Use them in that order. Do not
  invent a third. The `.example` top-level domain is reserved (RFC
  2606) and never resolves. Do not use a live top-level domain for a
  placeholder, because someone can register it.
- People: Alice, Bob, Carol, Dave in that order. For security text:
  Eve (eavesdropper), Mallory (attacker), Trent (trusted third
  party). Addresses: `alice@cyberdyne.example`.
- Addresses: `192.0.2.0/24`, `198.51.100.0/24`, `203.0.113.0/24`
  (RFC 5737), `2001:db8::/32` (RFC 3849), MAC `00-53-00-xx-xx-xx`
  (RFC 7042). Hostnames: `fw01.cyberdyne.example`.
- Never a real hostname, domain, address, or organization name from
  the code, unless the user asks for the real value in that document.
- In certificate and ACME-protocol documents, do not use ACME as an
  organization. Use Cyberdyne Systems only.
- Vendor and product names are technical names and stay: Cisco,
  Palo Alto Networks, Debian, PostgreSQL. The line is: the name of a
  thing you can buy or download stays. The name of a person or of
  the organization that runs the code does not.
- When an identifier itself contains a person's name or the
  organization's name (a config key, a function name, a path), you
  can write the identifier verbatim in a code span, because the code
  is the truth. Do not repeat the name in prose around it.
