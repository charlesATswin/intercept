/**
* Used to intercept traffic to and from the 
* local container and ask another sidecar if the traffic should be allowed.
*
* Based of ratelimited example.
*/
package intercept


import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"


	"github.com/valyala/fasthttp"

	"github.com/dapr/kit/logger"
	"github.com/dapr/components-contrib/middleware"
)

// Metadata is the ratelimit middleware config.
type interceptMetadata struct {
	DefaultSidecarPort int `json:"defaultSidecarPort"`
}


const (
	defaultSidecarPortKey = "defaultSidecarPort"

	// Defaults.
	defaultSidecarPort = 6969
)

// NewRateLimitMiddleware returns a new ratelimit middleware.
func NewMiddleware(logger logger.Logger) *Middleware {
	return &Middleware{logger: logger}
}

// Middleware is an intercept middleware.
type Middleware struct {
	logger logger.Logger
}

// GetHandler returns the HTTP handler provided by the middleware.
func (m *Middleware) GetHandler(metadata middleware.Metadata) (func(h fasthttp.RequestHandler) fasthttp.RequestHandler, error) {
	meta, err := m.getNativeMetadata(metadata)
	if err != nil {
		return nil, err
	}

	return func(h fasthttp.RequestHandler) fasthttp.RequestHandler {

		return func(ctx *fasthttp.RequestCtx) {
			postBody, _ := json.Marshal(map[string]string{
				"requestURI": string(ctx.Request.Header.RequestURI()),
				"header":     string(ctx.Request.Header.Header()),
				"body":       string(ctx.Request.Body()),
			})
			//  responseBody := bytes.NewBuffer(postBody)
		
			resp, err := http.Post("http://localhost:" + strconv.Itoa(meta.DefaultSidecarPort) + "/domain/messages/test",
				"application/json", bytes.NewBuffer(postBody))
			if err != nil {
				log.Fatalf("An Error Occured %v", err)
				h(ctx)
			} else {
				defer resp.Body.Close()
				//Read the response body
				body, _ := ioutil.ReadAll(resp.Body)
				if string(body) == "valid" {
					h(ctx)
				} else {
					log.Print("Nope, not going to let that through!")
					// todo: Is this how we reject the message?
					return
				}
			}
			// wrappedHandler(ctx)
		}
	}, nil
}


func (m *Middleware) getNativeMetadata(metadata middleware.Metadata) (*interceptMetadata, error) {
	var middlewareMetadata interceptMetadata

	middlewareMetadata.DefaultSidecarPort = defaultSidecarPort
	if val, ok := metadata.Properties[defaultSidecarPortKey]; ok {
		f, err := strconv.Atoi(val)
		if err != nil {
			return nil, fmt.Errorf("error parsing intercept middleware property %s: %+v", defaultSidecarPortKey, err)
		}
		middlewareMetadata.DefaultSidecarPort = f
	}

	return &middlewareMetadata, nil
}