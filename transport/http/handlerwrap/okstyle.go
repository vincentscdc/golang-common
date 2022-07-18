package handlerwrap

import (
	"net/http"
)

// OKStyle wraps an ok styled response that uses a special boolean field `ok` to indicate
// whether the request has been successfully responded or not. Specifically,
// On success, the response body will be transformed to
// ```json
// {
//		"ok": true,
//		<OTHER BODY PROPERTIES>
// }
// ```
//
// On failure, the response body will be
// ```json
// {
//		"ok": false,
//		"error_code": <ERROR CODE FROM APPLICATION>
//		"error_message": <ERROR MESSAGE FROM APPLICATION>
// }
// ```
func OKStyle(bodyName string, handler TypedHandler) TypedHandler {
	return func(r *http.Request) (*Response, *ErrorResponse) {
		resp, err := handler(r)
		if err != nil {
			return &Response{
				Body: map[string]any{
					"ok":            false,
					"error":         err.Error,
					"error_message": err.ErrorMessage,
				},
				StatusCode: http.StatusOK,
				Headers:    resp.Headers,
			}, err
		}

		return &Response{
			Body: map[string]any{
				"ok":     true,
				bodyName: resp.Body,
			},
			StatusCode: http.StatusOK,
			Headers:    resp.Headers,
		}, nil
	}
}
