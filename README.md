# rq
Go library for making http requests from .http files.

## Installation

```shell
go get github.com/go-rq/rq
```

## .http File Syntax

```http request
### <name of request>
< {% <pre-request
javascipt> %} 

<method> <url>
<header>:<value>
...

<request body>

< {% <post-request 
javascript> %}
```

### Scripts

Scripts can be embedded in the `.http` request directly or loaded from

```http
### <name of request>
< path/to/script.js

<method> <url>
<header>:<value>
...

<request body>

< path/to/script.js
```

### Examples

```http request
### Create a User
< {% 
    setEnv('host', 'http://localhost:8000');
    setEnv('name', 'r2d2');
%}
POST {{host}}/users
Content-Type: application/json

{
    "name": "{{name}}"
}

< {% 
    assertTrue("response code is 200", response.status === 200);
%}
```

### Running .http Files from Go Tests

The package `treqs`, short for `Testing Requests`, provides functions for running `.http` files from Go tests.

```go
package test

import (
  "testing"

  "github.com/go-rq/rq/treqs"
)

func TestFoo(t *testing.T) {
    ctx := rq.WithEnvironment(context.Background(), 
    map[string]string{"host": "http://localhost:8080"}) 

    // run all requests in testdata/foo.http
    treqs.RunFile(t, ctx, "testdata/foo.http")
  
    // run all requests in all .http files in the testdata directory
    treqs.RunDir(t, ctx, "testdata")
}
```
