# Language Server Challenge
Author: Caleb Doxsey

This is a language server I wrote for a programming challenge for a job I was subsequently rejected for without explanation.

## Summary

I opted to implement a plugin for sublime text for the Go programming language.

It consists of a python library (`languageserver-challenge.py`) and a Go http server
that implements the source code analysis.

There are 3 functions:

1. findreferences: which finds references for whatever's underneath the cursor
2. gotodefinition: which goes to the definition of whatever's underneath the cursor
3. hover: which implements documentation when you hover over something

Analysis is done using the go parsing libraries:

- `go/ast`
- `go/build`
- `go/parser`
- `go/token`
- `go/types`

Communication between the applications is done via JSON over HTTP.

## Installation

1. Install sublime text. You can just use the free trial.
2. Put this directory in the package directory for sublime text. On my machine it's something like this:

    /Users/caleb/Library/Application Support/Sublime Text 3/Packages/languageserver-challenge
    † tree
    .
    ├── README.md
    ├── server
    │   ├── analyzer.go
    │   ├── analyzer_test.go
    │   ├── api.go
    │   ├── api_test.go
    │   ├── main.go
    │   └── srcimporter.go
    ├── languageserver-challenge.py
    ├── languageserver-challenge.sublime-commands
    └── vendor

3. Run the go server. It listens on port 5000. (you'll need a recent version of Go that includes `go/types`)

    (cd server && go build -o /tmp/server . && /tmp/server)

4. Open sublime text. `findreferences` and `gotodefinition` can be performed by opening
   the command palette (cmd-shift-p on mac I think) and looking for "Find First Reference"
   and "Go To Definition". `hover` is done by just hovering over some text for a few seconds.

## Caveats

I decided to use the built-in parsing framework because writing a language parser takes a
significant amount of time, so I didn't think I'd have time to complete that and do much
of anything else. I was somewhat familiar with parsing Go code, so I knew this was going
to be difficult. There are a ton of edge cases and properly handling types and references
can be really challenging. (I saw Alan Donovan give a talk about this at GothamGo a while back)

The `go/types` library is currently crippled because it uses `.a` files and not the source code
to do type analysis. Because of this it won't include the filepaths for external libraries
out of the box. After getting nowhere with it for about an hour, I had to copy the importer
from `srcimporter` (it was internal so I couldn't use it directly) which allowed me to pass
in a `FileSet` so I could resolve imports.

I tested a few things, but I suspect a lot won't work. I don't think it handles partial code blocks (I
think the code needs to be a valid Go program to provide any assistance), and although I started
with a function signature that could take the contents of the file, I didn't implement that.
(This is important to provide assistance for things you haven't saved yet) I also doubt it works
with vendor directories. (You'll find warnings about this littered throughout the docs of
these libraries)

I only had time to implement one test. It dumps a go file to a temp directory and tries to do
analysis on it. A lot more could be done there. Also there's a lot of copy-pasta I could clean up.

The UI bits for `hover` are pretty rudimentary. I also didn't get a chance to provide the actual
comments from the source (ala `godoc`), though in principle that ought to be possible.
`findreferences` only works with the current project.

Also all of this is really slow. It reparses the code every time and there's an O(n) map traversal
so `findreferences` isn't very practical. `gotodefinition` seems to perform better. (FWIW in actual
Go editors finding references is also very slow) In principle it ought to be possible to build an
index of this information so that subsequent searches are much faster. The challenge will be in
keeping all of it in sync.

Anyway this is all I managed to complete in the allotted time.
