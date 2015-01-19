# gomodel

[![Build Status](https://travis-ci.org/clarkf/gomodel.svg?branch=master)](https://travis-ci.org/clarkf/gomodel)
[![Coverage Status](https://coveralls.io/repos/clarkf/gomodel/badge.svg?branch=master)](https://coveralls.io/r/clarkf/gomodel?branch=master)

The state of ORMs in the go world isn't great (as with most strongly typed
languages), and writing your own SQL still seems to be the best way to get
stuff done. gomodel helps with this task by unmarshaling your data into usable
structs.

Under the hood, gomarshal uses the `reflect` package to examine your structs
and tries to correlate column names with struct fields.
## Installation

Install with a go get:

    $ go get -u github.com/clarkf/gomodel

Or clone it yourself:

    $ git clone https://github.com/clarkf/gomodel.git $GOPATH/github.com/clarkf/gomodel

Then, in your code:

```go
import "github.com/clarkf/gomodel"
```

## Usage

gomodel was designed to compliment the `database/sql` package (though it doesn't
necessarily depend on it).  Assuming that you already have an
established `*sql.DB` connection, `db`:

```go
type User stuct {
    ID           int64
    Username     string
    PasswordHash string
    Email        string `sql:"email_address"`
}

// Query the database and create the *sql.Rows row set
rows, err := db.Query("SELECT * FROM users")
if err != nil {
    // handle query error
}

// Scan the database rows into an array of users
var users []User
if err := gomodel.ScanRows(rows, &users); err != nil {
    // handle scan error
}

// users is now populated!
```

gomodel requires that you pass a pointer to a slice of structs (or a
slice of pointers to structs) as the second argument to scan rows.  For
example:

```go
var models []ModelStruct  // GOOD: Slice of structs
var models []*ModelStruct // GOOD: Slice of pointers to structs
var models []string       // BAD: Not a struct
var models ModelStruct    // BAD: Not a slice (see Scan below)

// And most importantly, be sure to pass a pointer to slice to ScanRows:
gomodel.ScanRows(_, &models)
```

### Scanning a single result

The `database/sql` package provides a facility for fetching one row from
a database: `*db.QueryRow`.  gomodel can fill a single model for you
provided you know the order of the columns:
```go
row := db.Query("SELECT id, username FROM users WHERE id = 1234 LIMIT 1")
columns := []string{"id", "username"}
user := &User{}

if err := gomodel.Scan(row.Scan, columns, user); err != nil {
    // Handle scanning and database errors here.  sql.Row defers
    // database errors to this point.
}
```

## Contributing

All contributions and issues are welcome.  If you intend on contributing,
please:

1. Fork the repository and create a descriptive feature branch
2. Write your code (and accompanying tests)
3. Ensure that the test suite passes and test coverage has not fallen by
   running `go test -v -cover` before committing.
4. Push to your feature branch and submit a pull request.

## License

The MIT License (MIT)

Copyright (c) 2015 Clark Fischer

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
