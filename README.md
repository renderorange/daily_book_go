# daily_book_go

rewrite of [daily_book](https://github.com/renderorange/daily_book) in Go.

`daily_book_go` finds quotes in free ebooks at [project gutenberg](https://www.gutenberg.org) then tweets them.

## USAGE

```bash
$ ./quote --help
usage: ./quote [-d] [-m] <book number> [-t]

options:
  -d	print more information during the run
  -m int
    	manually specify the book number
  -t	post the quote to twitter
```

### Find a quote for a random book

```bash
$ ./quote

title: Tom Slade's Double Dare
author: Percy Keese Fitzhugh

"I think one of our boys from camp ought to do that," said one of the other scoutmasters. "How about you, Roy?"  https://gutenberg.org/ebooks/19590
```

### Find a quote from a specific book

```bash
$ ./quote -m 220

title: The Secret Sharer
author: Joseph Conrad

"Bless my soul! Do you mean, sir, in the dark amongst the lot of all them islands and reefs and shoals?"  https://gutenberg.org/ebooks/220
```

### Error and debug output is sent to STDERR

To print more information during the run, `-d` can be defined to send additional output to `STDERR`.

```bash
$ ./quote -m 220 -d 2> stderr.o > stdout.o
$ cat stderr.o
[debug] page link: https://gutenberg.org/ebooks/220
[debug] book link: https://gutenberg.pglaf.org/2/2/220/220.txt
[debug] parser is in head
[debug] parser is in body
[debug] parser is in footer
[debug] title: The Secret Sharer
[debug] author: Joseph Conrad
[debug] paragraphs found: 356
[debug] quote was found: "Hadn't the slightest idea. I am the mate of her--" He paused and corrected himself. "I should say I _was_." 
[debug] quote was found: "Unless you manage to recover him before tomorrow," I assented, dispassionately.... "I mean, alive." 
[debug] quote was found: "Bless my soul! Do you mean, sir, in the dark amongst the lot of all them islands and reefs and shoals?" 
[debug] quote was found: "I won't be there to see you go," I began with an effort. "The rest ... I only hope I have understood, too." 
```

Non debug information and error output is also sent to `STDERR`.

```bash
$ ./quote 2> stderr.o > stdout.o
$ cat stderr.o
[info] quote was not found - 3470
[info] quote was not found - 6760
[info] quote was not found - 35385
[info] quote was not found - 5042
[info] quote was not found - 27706
[info] quote was not found - 16136
$ cat stdout.o

title: Phebe, Her Profession
author: Anna Chapin Ray

"I'm glad," she said slowly, as she rose. "No; don't come, Cousin Ted. I want to think it over."  https://gutenberg.org/ebooks/12584
```

```bash
$ ./quote -m 52751 2> stderr.o
$ cat stderr.o
[error] download response was 404 - 52751
```

```bash
$ ./quote -m 46392 2> stderr.o
$ cat stderr.o
[info] quote was not found - 46392
```

## COPYRIGHT AND LICENSE

`daily_book_go` is Copyright (c) 2022 Blaine Motsinger under the MIT license.

## AUTHOR

Blaine Motsinger `blaine@renderorange.com`
