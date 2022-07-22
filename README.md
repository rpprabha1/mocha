<h1 id="mocha-top" align="center">Mocha</h1>

<div align="center">
    <a href="#"><img src="logo.png" width="120px" alt="Mocha Logo"></a>
    <p align="center">
        HTTP Mocking Tool for Go
        <br />
    </p>
    <div>
      <a href="https://github.com/vitorsalgado/mocha/actions/workflows/ci.yml">
        <img src="https://github.com/vitorsalgado/mocha/actions/workflows/ci.yml/badge.svg" alt="CI Status" />
      </a>
      <a href="https://codecov.io/gh/vitorsalgado/mocha">
        <img src="https://codecov.io/gh/vitorsalgado/mocha/branch/main/graph/badge.svg?token=XOFUV52P31" alt="Coverage"/>
      </a>
      <a href="#">
        <img alt="GitHub go.mod Go version" src="https://img.shields.io/github/go-mod/go-version/vitorsalgado/mocha">
      </a>
      <a href="https://pkg.go.dev/github.com/vitorsalgado/mocha">
        <img src="https://pkg.go.dev/badge/github.com/vitorsalgado/mocha.svg" alt="Go Reference">
      </a>
    </div>
</div>

## Overview

HTTP server mocking tool for Go.  
**Mocha** creates a real HTTP server and lets you configure response stubs for HTTP Requests when it matches a set of
matchers.
It provides a functional like API that allows you to match any part of a request against a set of matching
functions that can be composed.

Inspired by [WireMock](https://github.com/wiremock/wiremock) and [Nock](https://github.com/nock/nock).

## Installation

```bash
go get github.com/vitorsalgado/mocha
```

## Features

- Configure HTTP response stubs for specific requests based on a criteria set.
- Matches request URL, headers, queries, body.
- Stateful matches to create scenarios, mocks for a specific number of calls.
- Response body template.
- Response delays.
- Run in your automated tests.

## How It Works

**Mocha** works by creating a real HTTP Server that you can configure response stubs for HTTP requests when they match a
set request matchers. Mock definitions are stored in memory in the server and response will continue to be served as
long as the requests keep passing the configured matchers.  
The basic is workflow for a request is:

- run configured middlewares
- mocha parses the request body based on:
  - custom `RequestBodyParser` configured
  - request content-type
- mock http handler tries to find a mock for the incoming request were all matchers evaluates to true
  - if a mock is found, it will run **post matchers**.
  - if all matchers passes, it will use mock reply implementation to build a response
  - if no mock is found, **it returns an HTTP Status Code 418 (teapot)**.
- after serving a mock response, it will run any `core.PostAction` configured.

## Getting Started

Usage typically looks like the example below:

```
func Test_Example(t *testing.T) {
	m := mocha.New(t)
	m.Start()

	scoped := m.AddMocks(mocha.Get(expect.URLPath("/test")).
		Header("test", expect.ToEqual("hello")).
		Query("filter", expect.ToEqual("all")).
		Reply(reply.Created().BodyString("hello world")))

	req, _ := http.NewRequest(http.MethodGet, m.URL()+"/test?filter=all", nil)
	req.Header.Add("test", "hello")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	body, err := ioutil.ReadAll(res.Body)

	assert.Nil(t, err)
	assert.True(t, scoped.Called())
	assert.Equal(t, 201, res.StatusCode)
	assert.Equal(t, string(body), "hello world")
}
```

## Configuration

Mocha has two ways to create an instance: `mocha.New()` and `mocha.NewSimple()`.  
`mocha.NewSimple()` creates a new instance with default values for everything.  
`mocha.New(t, ...config)` needs a `core.T` implementation and allows to configure the mock server.
You use `testing.T` implementation. Mocha will use this to log useful information for each request match attempt.
Use `mocha.Configure()` or provide a `mocha.Config` to configure the mock server.

## Request Matching

Matchers can be applied to any part of a Request and **Mocha** provides a fluent API to make your life easier.  
See usage examples below:

### Method and URL

```
m := mocha.New(t)
m.AddMocks(mocha.Request().Method(http.MethodGet).URL(expect.URLPath("/test"))
```

### Header

```
m := mocha.New(t)
m.AddMocks(mocha.Get(expect.URLPath("/test")).
    Header("test", expect.ToEqual("hello")))
```

### Query

```
m := mocha.New(t)
m.AddMocks(mocha.Get(expect.URLPath("/test")).
    Query("filter", expect.ToEqual("all")))
```

### Body

**Matching JSON Fields**

```
m := mocha.New(t)
m.AddMocks(mocha.Post(expect.URLPath("/test")).
    Body(
        expect.JSONPath("name", expect.ToEqual("dev")), expect.JSONPath("ok", expect.ToEqual(true))).
    Reply(reply.OK()))
```

### Form URL Encoded Fields

```
m.AddMocks(mocha.Post(expect.URLPath("/test")).
    FormField("field1", expect.ToEqual("dev")).
    FormField("field2", expect.ToContain("qa")).
    Reply(reply.OK()))
```

## Replies

You can define a response that should be served once a request is matched.  
**Mocha** provides several ways to configure a reply. The built-in replies are:

- single reply
- random replies
- sequence replies
- reply from a function
- reply from a proxied request

Replies are based on the `core.Reply` interface.  
It's also possible to configure response bodies from templates. **Mocha** uses Go Templates.
Replies usage examples:

### Basic Reply

```
m := mocha.New(t)
m.AddMocks(mocha.Get(expect.URLPath("/test")).
    Reply(reply.OK())
```

### Sequence

```
m := mocha.New(t)
m.AddMocks(mocha.Get(expect.URLPath("/test")).
    Reply(reply.Seq().
	    Add(InternalServerError(), BadRequest(), OK(), NotFound())))
```

### Random Replies

```
m := mocha.New(t)
m.AddMocks(mocha.Get(expect.URLPath("/test")).
    Reply(reply.Rand().
		Add(BadRequest(), OK(), Created(), InternalServerError())))
```

### Reply Function

```
m := mocha.New(t)
m.AddMocks(mocha.Get(expect.URLPath("/test")).
    ReplyFunction(func(r *http.Request, m *core.Mock, p parameters.Params) (*core.Response, error) {
        return &core.Response{Status: http.StatusAccepted}, nil
    }))
```

### Proxied From

**reply.From** will forward the request to the given destination and serve the response from the forwarded server.  
It`s possible to add extra headers to the request and the response and also remove unwanted headers.

```
m := mocha.New(t)
m.AddMocks(mocha.Get(expect.URLPath("/test")).
    Reply(reply.From("http://example.org").
		ProxyHeader("x-proxy", "proxied").
		RemoveProxyHeader("x-to-be-removed").
		Header("x-res", "response"))
```

### Body Template

**Mocha** comes with a built-in template parser based on Go Templates.  
To serve a response body from a template, follow the example below:

```
templateFile, _ := os.Open("template.tmpl"))
content, _ := ioutil.ReadAll(templateFile)

m := mocha.New(t)
m.AddMocks(mocha.Get(expect.URLPath("/test")).
    Reply(reply.
        OK().
        BodyTemplate(reply.NewTextTemplate().
            FuncMap(template.FuncMap{"trim": strings.TrimSpace}).
            Template(string(content))).
        Model(data))
```

### Specifying Headers

```
m := mocha.New(t)
m.AddMocks(mocha.Get(expect.URLPath("/test")).
    Reply(reply.OK().Header("test", "test-value"))
```

## Delay Responses

You can configure a delay to responses to simulate timeouts, slow requests and any other timing related scenarios.  
See the example below:

```
delay := time.Duration(1250) * time.Millisecond

m.AddMocks(Get(expect.URLPath("/test")).
    Reply(reply.
        OK().
        Delay(delay)))
```

## Assertions

### Mocha Instance

Mocha instance provides methods to assert if associated mocks were called or not, how many times they were called,
allows you to enable/disable then and so on.  
The available assertion methods on mocha instance are:

- AssertCalled: asserts that all associated mocks were called at least once.
- AssertNotCalled: asserts that associated mocks were **not** called.
- AssertHits: asserts that the sum of calls is equal to the expected value.

### Scope

Mocha instance method `AddMocks` returns a `Scoped` instance that holds all mocks created.  
`Scopes` allows you control related mocks, enabling/disabling, checking if they were called or not. Scoped instance also
provides **assertions** to facility **tests** verification.
See below the available test assertions:

- AssertCalled: asserts that all associated mocks were called at least once.
- AssertNotCalled: asserts that associated mocks were **not** called.

### Matchers

Mocha provides several matcher functions to facilitate request matching and verification.
See the package `expect` for more details.  
You can create custom matchers using these two approaches:

- create a `expect.Matcher` struct
- use the function `expect.Func` providing a function with the following
  signature: `func(v any, a expect.Args) (bool, error)`

---

## Future Plans

- [ ] Configure mocks with JSON/YAML files
- [ ] CLI
- [ ] Docker
- [ ] Proxy and Record

## Contributing

Check our [Contributing](CONTRIBUTING.md) guide for more details.

## License

[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fvitorsalgado%2Fmocha.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fvitorsalgado%2Fmocha?ref=badge_shield)

This project is [MIT Licensed](LICENSE).

<p align="center"><a href="#mocha-top">back to the top</a></p>
