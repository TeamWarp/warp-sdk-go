# Warp

This library provides convenient access to the Warp REST API from Go.

The REST API documentation can be found on [docs.warp.co](https://docs.warp.co). The full API of this library can be found in [api.md](./api.md).

<br />

## Contents

- [Installation](#installation)
- [Usage](#usage)
- [API Reference](./api.md)
- [Authentication](#authentication)
- [Errors](#errors)
- [Client Options](#client-options)
- [Request Options](#request-options)
- [Retries and Timeouts](#retries-and-timeouts)
- [Webhooks](#webhooks)
- [Helpers](#helpers)
- [Logging](#logging)
- [Requirements](#requirements)

<br />

## Installation

```sh
go get github.com/TeamWarp/warp-go-sdk
```

<br />

## Usage

```go
package main

import (
	"context"
	"fmt"
	"os"

	sdk "github.com/TeamWarp/warp-go-sdk"
	"github.com/TeamWarp/warp-go-sdk/option"
)

func main() {
	client := sdk.NewClient(
		option.WithAPIKey(os.Getenv("WARP_API_KEY")),
	)

	healthPlan, err := client.Benefits.HealthPlans.List(context.Background(), sdk.BenefitHealthPlanListParams{
		Limit:    sdk.F[string]("limit"),
		Statuses: sdk.F[[]sdk.PublicHealthPlanStatus]([]sdk.PublicHealthPlanStatus{"active"}),
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(healthPlan)
}
```

The examples in the following sections assume a `client` configured as shown above.

See the [API reference](./api.md) for every available operation.

<br />

## Authentication

Pass credentials to the generated client constructor. Environment variables are read automatically when supported by the target runtime.

| Option | Type | Default | Description |
| --- | --- | --- | --- |
| `option.WithAPIKey` | `func(string) option.RequestOption` | - | The API key for header authorization. Defaults to WARP_API_KEY. |

Declared schemes:

- `apiKey` API key in header `x-api-key`

<br />

## Errors

Non-success responses return generated API errors. Error objects expose status, headers, response body, and request metadata where the target runtime supports it.

```go
healthPlan, err := client.Benefits.HealthPlans.List(context.Background(), sdk.BenefitHealthPlanListParams{
	Limit:    sdk.F[string]("limit"),
	Statuses: sdk.F[[]sdk.PublicHealthPlanStatus]([]sdk.PublicHealthPlanStatus{"active"}),
})
if err != nil {
	var apiErr *sdk.Error
	if errors.As(err, &apiErr) {
		fmt.Println(apiErr.StatusCode, apiErr.JSON.RawJSON())
	}
	panic(err)
}

// imports: "context", "errors", "fmt", sdk "github.com/TeamWarp/warp-go-sdk"
```

Documented error statuses: `400`, `401`, `403`, `404`, `409`, `422`, `429`, `500`.

<br />

## Client Options

Configure the generated client by setting any of these options when you create it.

```go
client := sdk.NewClient(
	option.WithBaseURL("https://api.example.com"),
	option.WithMaxRetries(2),
	option.WithRequestTimeout(60*time.Second),
)

// imports: sdk "github.com/TeamWarp/warp-go-sdk", "github.com/TeamWarp/warp-go-sdk/option", "time"
```

| Option | Type | Default | Description |
| --- | --- | --- | --- |
| `option.WithAPIKey` | `func(string) option.RequestOption` | `os.Getenv("WARP_API_KEY")` | The API key for header authorization. |
| `option.WithEnvironmentProduction` | `func() option.RequestOption` | - | Select the production API environment. |
| `option.WithBaseURL` | `func(string) option.RequestOption` | `os.Getenv("WARP_BASE_URL")` | Override the default API base URL. |
| `option.WithRequestTimeout` | `func(time.Duration) option.RequestOption` | - | Maximum time to wait for each request attempt. |
| `option.WithMaxRetries` | `func(int) option.RequestOption` | `2` | Number of retries for temporary failures. |
| `option.WithHTTPClient` | `func(option.HTTPClient) option.RequestOption` | - | Custom HTTP client or transport implementation. |

<br />

## Request Options

| Option | Type | Default | Description |
| --- | --- | --- | --- |
| `option.WithHeader` | `func(string, string) option.RequestOption` | - | Set a per-request header. |
| `option.WithQuery` | `func(string, string) option.RequestOption` | - | Set a per-request query parameter. |
| `option.WithRequestBody` | `func(string, any) option.RequestOption` | - | Override the serialized request body and content type. |
| `option.WithResponseInto` | `func(**http.Response) option.RequestOption` | - | Capture the raw HTTP response. |
| `option.WithResponseBodyInto` | `func(any) option.RequestOption` | - | Override the response deserialization target. |

<br />

## Retries and Timeouts

Generated clients support request timeouts and retry temporary failures such as network errors, 408, 409, 429, and 5xx responses. Retry delays honor `Retry-After` headers when present. Tune the retry and timeout client options shown above, or override them per request.

<br />

## Webhooks

The webhook helper verifies the signature of each delivery your server receives, then decodes its body into a typed event. Pass it the request body exactly as it arrived, before any JSON parsing, together with the request headers: the signature covers the raw body, so one that was parsed and re-serialized fails verification. A delivery whose signature does not verify is rejected before it is decoded.

```go
// `payload` is the raw request body; `r` is the *http.Request it arrived on.
event, err := client.Webhooks.Unwrap(payload, r.Header)
if err != nil {
	// The signature did not verify, or the payload did not decode.
}
```

The signing secret comes from the client's `option.WithWebhookSecret` option, which defaults to the `WARP_WEBHOOK_SECRET` environment variable. Pass `option.WithWebhookSecret(secret)` to use a different secret for one call.

The same resource also has:

- `client.Webhooks.VerifySignature(payload, r.Header, secret, time.Now())` checks a delivery's signature, and fails when it does not verify.
- `client.Webhooks.ParseUnsafe(payload)` decodes a body into a typed event without checking its signature. Use it only for a body verified elsewhere, or in tests.

<br />

## Helpers

- Pass `option.WithResponseInto(&raw)` to capture the underlying `*http.Response` for a request.
- Use the generated `F`, `Null`, `Raw`, `Int`, `String`, `Float`, `Bool`, and `FileParam` helpers when setting optional params.

<br />

## Logging

- Wrap the HTTP client with `option.WithMiddleware(...)` to add request logging or tracing.

<br />

## Requirements

- Go 1.22 or newer

Powered by Scalar.
