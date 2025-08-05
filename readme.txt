mdconf
==========

license: public domain

config file format in markdown syntax.  it's like ini but it's
markdown.

installation
-----------------

```
go get github.com/zedhead037/mdconf
```

test: `git clone`, and then `go test -v`.

syntax
----------

basically like this:

```
# section 1
+ key1.1: value
+ key1.2: value2
+ key1.3: value 3

## section 1.1
+ key1.1.1: value
+ key1.1.2: value2
+ key1.1.3: value 3

## section 1.2
+ key1.2.1: value
+ key1.2.2: value2
+ key1.2.3: value 3

# section 2

## section 2.1
+ key2.1.1: value
+ key2.1.2: value
```

keys are always treated as if its whitespaces are trimmed from both
sides. e.g. this:

```
+ key1: value1
```

is equivalent to this:

```
+         key1    : value1
```

the same goes for values.

any character after the backward slash `\` is to be kept.  this is the
same everywhere.  for example, this:

```
+       key1:value1
```

has a key of `key1`, but this:

```
+  \    key1:value1
```

has a key of `    key1`. and this:

```
###  blah
```

is a level-3 header w/ a section name of `blah`, but this:

```
##\#  blah
```

is a level-2 header w/ a section name of `#  blah`.

any empty lines and any spaces at the beginning and the end of a line
is to be ignored.

any lines that starts with two slashes `//` (after skipping the
whitespaces at the beginning) is to be considered as a comment line,
which is also to be ignored.

multi-line value is supported.  at the end of the first line write one
(1) single slash `\` with absolutely no character after to signifies
the value continues to the next line.  all continuing lines is NOT
trimmed.  for example:

```
+ key1: value\
blahblahblah\
    blabasdfd \
blahblah
```

has the key `key1` and the value:

```
value
blahblahblah
    blabasdfd 
blahblah
```

multi-line key and section header are NOT supported.

