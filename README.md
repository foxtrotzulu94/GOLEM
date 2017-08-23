# golem
## The Golang Ordered List Executive Manager

## What's this?
This is just a small tool written in Go that I used both to keep track of my evergrowing lists of lists and also learn the language (my commit history will show I took some bad decisions along the way, but hopefully corrected most of them by now)

## How to Use?
`golem` attempts to be simple to use and extend. There are several list types built-in (anime, books, games) which you can run actions on. Typical usage would look something like:

    $ golem scan books
    $ golem list books 15

The above two commands will scan the current directory for a books list ("Books.txt"), add all entries (one per line) in the database and then list the top 15 books in your collection

Actions are simply what you can do with a given list. The most common are:

* Scan
* List
* Detail
* Finish
* Remove

## Why the name though?
The namesake of this program is this [pokemon](). It was good enough because it had *go* in the name :)