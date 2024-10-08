// Code generated by goa v3.17.2, DO NOT EDIT.
//
// transaction HTTP client encoders and decoders
//
// Command:
// $ goa gen wallet/design

package client

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	transaction "wallet/gen/transaction"

	goahttp "goa.design/goa/v3/http"
)

// BuildHealthcheckRequest instantiates a HTTP request object with method and
// path set to call the "transaction" service "healthcheck" endpoint
func (c *Client) BuildHealthcheckRequest(ctx context.Context, v any) (*http.Request, error) {
	u := &url.URL{Scheme: c.scheme, Host: c.host, Path: HealthcheckTransactionPath()}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, goahttp.ErrInvalidURL("transaction", "healthcheck", u.String(), err)
	}
	if ctx != nil {
		req = req.WithContext(ctx)
	}

	return req, nil
}

// DecodeHealthcheckResponse returns a decoder for responses returned by the
// transaction healthcheck endpoint. restoreBody controls whether the response
// body should be restored after having been read.
func DecodeHealthcheckResponse(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (any, error) {
	return func(resp *http.Response) (any, error) {
		if restoreBody {
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			resp.Body = io.NopCloser(bytes.NewBuffer(b))
			defer func() {
				resp.Body = io.NopCloser(bytes.NewBuffer(b))
			}()
		} else {
			defer resp.Body.Close()
		}
		switch resp.StatusCode {
		case http.StatusOK:
			var (
				body HealthcheckResponseBody
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("transaction", "healthcheck", err)
			}
			err = ValidateHealthcheckResponseBody(&body)
			if err != nil {
				return nil, goahttp.ErrValidationError("transaction", "healthcheck", err)
			}
			res := NewHealthcheckResultOK(&body)
			return res, nil
		default:
			body, _ := io.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse("transaction", "healthcheck", resp.StatusCode, string(body))
		}
	}
}

// BuildCreateRequest instantiates a HTTP request object with method and path
// set to call the "transaction" service "create" endpoint
func (c *Client) BuildCreateRequest(ctx context.Context, v any) (*http.Request, error) {
	u := &url.URL{Scheme: c.scheme, Host: c.host, Path: CreateTransactionPath()}
	req, err := http.NewRequest("POST", u.String(), nil)
	if err != nil {
		return nil, goahttp.ErrInvalidURL("transaction", "create", u.String(), err)
	}
	if ctx != nil {
		req = req.WithContext(ctx)
	}

	return req, nil
}

// EncodeCreateRequest returns an encoder for requests sent to the transaction
// create server.
func EncodeCreateRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, any) error {
	return func(req *http.Request, v any) error {
		p, ok := v.(*transaction.CreatePayload)
		if !ok {
			return goahttp.ErrInvalidType("transaction", "create", "*transaction.CreatePayload", v)
		}
		{
			head := p.SourceType
			req.Header.Set("Source-Type", head)
		}
		body := NewCreateRequestBody(p)
		if err := encoder(req).Encode(&body); err != nil {
			return goahttp.ErrEncodingError("transaction", "create", err)
		}
		return nil
	}
}

// DecodeCreateResponse returns a decoder for responses returned by the
// transaction create endpoint. restoreBody controls whether the response body
// should be restored after having been read.
func DecodeCreateResponse(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (any, error) {
	return func(resp *http.Response) (any, error) {
		if restoreBody {
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			resp.Body = io.NopCloser(bytes.NewBuffer(b))
			defer func() {
				resp.Body = io.NopCloser(bytes.NewBuffer(b))
			}()
		} else {
			defer resp.Body.Close()
		}
		switch resp.StatusCode {
		case http.StatusAccepted:
			return nil, nil
		default:
			body, _ := io.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse("transaction", "create", resp.StatusCode, string(body))
		}
	}
}
