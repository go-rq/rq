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

#### Scripting API

##### setEnv(key string, value string)

The runtime environment variables can be modified from either pre or post
request script. No return value.

```javascript
setEnv('host', 'http://localhost:8000');
```

##### getEnv(key string)

Returns the value of the environment variable `key`.

```javascript
getEnv('host'); // returns 'http://localhost:8000'
```

##### assert(condition boolean, message string)

Some assertion which resolves to a boolean value can be made for the
condition with a corresponding message. Each assertion is extracted
from the scripting runtime and added to the Response object.

```javascript
assert(response.status === 200, 'response code is 200');
```

##### log(message string)

Appends a log message to `Request.Logs` property allowing the go
runtime to access debug information from the scripting runtime.

```javascript
log(`the response status code was ${response.statusCode}`);
```

##### `request` and `response` objects

The `response` object is available in the post-request script and the `request`
object is available in both the pre and post-request scripts. Changes made to
these objects are not reflected in the actual request or response, rather they
are used for assertions, control flow, and setting up the environment. There
are a few exceptions to this rule, such as the `skip` property on `request`,
see the next section for more information.

```javascript
request = {
    name: 'Create A User',
    method: 'POST',
    url: 'http://localhost:8000/users',
    headers: {
      'Content-Type': 'application/json',
    }, 
    body: `{
        name: 'r2d2'
    }`,
}
```

```javascript
response = {
    status: 'OK',
    statusCode: 200,
    headers: {
        'Content-Type': 'application/json',
    },
    body: `{
        id: '1234',
        name: 'r2d2'
    }`,

    // `json` is the parsed json body that is only available if the response Content-Type
    // header contains `application/json`
    json: {
      id: '1234',
      name: 'r2d2'
    } 
}
```

##### Skipping Requests (pre-request script)

You can prevent the request from being made from pre-request scripts being
setting the `skip` property on the `request` object to `true`

#### Examples

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
    assertTrue(response.status === 200, 'response code is 200');
%}
```

### Running .http Files from Go Tests

The package `treqs`, short for `Testing Requests`, provides functions
for running `.http` files from Go tests.

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
