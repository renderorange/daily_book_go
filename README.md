# daily_book_go

rewrite of [daily_book](https://github.com/renderorange/daily_book) in Go.

## USAGE

```bash
$ ./quote --help
usage: ./quote [-d] [-m] <book number>

options:
  -d	print more information during the run
  -m int
    	manually specify the book number
```

```bash
./quote -m 220

title: The Secret Sharer
author: Joseph Conrad

"Bless my soul! Do you mean, sir, in the dark amongst the lot of all them islands and reefs and shoals?"  https://gutenberg.org/ebooks/220
```

To print more information during the run, `-d` can be defined to send debug output to `STDERR`.

```bash
$ ./quote -m 220 -d 2> debug.o

title: The Secret Sharer
author: Joseph Conrad

"Unless you manage to recover him before tomorrow," I assented, dispassionately.... "I mean, alive."  https://gutenberg.org/ebooks/220

$ cat debug.o 
[debug] page_link: https://gutenberg.org/ebooks/220
[debug] book_link: https://gutenberg.pglaf.org/2/2/220/220.txt
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

## COPYRIGHT AND LICENSE

`daily_book_go` is Copyright (c) 2022 Blaine Motsinger under the MIT license.

## AUTHOR

Blaine Motsinger `blaine@renderorange.com`
