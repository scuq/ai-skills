# Word choice without the dictionary

## BLUF

The scraibe standard does not contain the ASD-STE100 dictionary.
This file gives 12 patterns for the choice of a word that section 2 of the standard does not name.
They are a direction, not a test. Read this file when you must judge a word.

## Patterns

This standard does not contain the dictionary (`reference/attribution.md`). When you
must judge a word that is not in section 2, use these patterns. They
come from an analysis of the 2198 entries of the Issue 9 dictionary.
They are a direction, not a test (section 1a).

1. Keep the part of speech. A verb becomes a verb, an adjective an
   adjective. Four of five replacements in the dictionary do this.
2. Choose the concrete word for this one meaning. The dictionary
   spreads its replacements over more than 600 approved words. It
   has no small set of general words. A word with several meanings
   gets one replacement per meaning: "perform a test" becomes "do a
   test", "perform a measurement" becomes "measure". Ask what <!-- ste-lint: disable-line nominalization -->
   physically happens: put, move, remove, stop, start, attach, hold, <!-- ste-lint: disable-line synonym-rotation -->
   turn, show, tell, give, find.
3. An action noun becomes its verb: "the absorption of moisture by
   the gel" becomes "the gel absorbs the moisture". Nouns in -tion,
   -ment, and -ance are the usual case.
4. A tool used as a verb becomes a general verb and the tool: "bolt
   the bracket" becomes "attach the bracket with bolts". A technical
   noun is never a verb (STE rule 1.7).
5. When no one word fits, build a phrase from approved words:
   "exceed" becomes "more than", "excessive" becomes "too much",
   "simultaneously" becomes "at the same time".
6. Absence and negation: "not", "no", "none", or the approved
   negative word. "absent" becomes "missing" or "not installed",
   "faulty" becomes "unsatisfactory", "abnormal" becomes "unusual".
7. Degree: "absolutely", "completely", "entirely", and "exactly"
   become "fully". "critical" becomes "very important".
8. One member of a word family is approved. When the noun is not,
   its verb usually is. Write the sentence around the approved
   member.
9. A word with the prefix "re-" becomes "again" or "back" with the
   basic word.
10. A Latin suffix is a signal, not a verdict. About four of five
    dictionary words in -ate, -ize, -ify, -ity, -ence, -ous, and
    -able are not approved. The suffixes -ly and -al say nothing.
11. A vague verb has no one replacement: act, affect, handle,
    process. Name the specific action the code does.
12. Technical nouns and technical verbs are outside the dictionary
    (STE rules 1.5 to 1.13). The name of a part, a product, or a
    process in your field stays. Use the shortest one, and always
    the same one.

## Preferred terms and hyphenation

Preferred terms (man-pages): filename, filesystem, hostname, pathname,
timestamp, timezone, username, superuser, symbolic link, nonzero,
lowercase, uppercase, user space, run time (noun), run-time (adjective).

Avoid: "OS" → "operating system", "tty" → "terminal",
"non-root" → "unprivileged user", "manpage" → "man page".

Hyphenate a compound before a noun: "command-line argument",
"read-only file". No hyphen after multi, non, pre, re, sub:
nonblocking, subdirectory, reinitialize. Keep the hyphen before a
capital: non-ASCII.
