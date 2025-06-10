package encoding

import (
	"errors"
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"strings"

	"github.com/thinkgos/encoding/codec"
	"github.com/thinkgos/encoding/form"
	"github.com/thinkgos/encoding/json"
)

const defaultMemory = 32 << 20

// Content-Type MIME of the most common data formats.
const (
	// MIMEURI is special form query.
	Mime_Query = "__MIME__/QUERY"
	// Mime_Uri is special form uri.
	Mime_Uri = "__MIME__/URI"
	// Mime_Wildcard is the fallback special MIME type used for requests which do not match
	// a registered MIME type.
	Mime_Wildcard = "*"

	Mime_JSON              = "application/json"
	Mime_HTML              = "text/html"
	Mime_XML               = "application/xml"
	Mime_XML2              = "text/xml"
	Mime_Plain             = "text/plain"
	Mime_PostForm          = "application/x-www-form-urlencoded"
	Mime_MultipartPostForm = "multipart/form-data"
	Mime_PROTOBUF          = "application/x-protobuf"
	Mime_MSGPACK           = "application/x-msgpack"
	Mime_MSGPACK2          = "application/msgpack"
	Mime_YAML              = "application/x-yaml"
	Mime_TOML              = "application/toml"
)

var (
	acceptHeader      = http.CanonicalHeaderKey("Accept")
	contentTypeHeader = http.CanonicalHeaderKey("Content-Type")
)

// Encoding is a mapping from MIME types to Marshalers.
type Encoding struct {
	mimeMap      map[string]codec.Marshaler
	mimeQuery    codec.FormMarshaler
	mimeUri      codec.UriMarshaler
	mimeWildcard codec.Marshaler
}

// New encoding with default Marshalers
// Default:
//
//	Mime_PostForm: form.Codec
//	Mime_MultipartPostForm: form.MultipartCodec
//	Mime_JSON: json.Codec
//	mime_Query: form.QueryCodec
//	mime_Uri:   form.UriCodec
//	mime_Wildcard: json.Codec
//
// you can manually register your custom Marshaler.
//
//	Mime_PROTOBUF: proto.Codec
//	Mime_XML:      xml.Codec
//	Mime_XML2:     xml.Codec
//	Mime_MSGPACK:  msgpack.Codec
//	Mime_MSGPACK2: msgpack.Codec
//	Mime_YAML:     yaml.Codec
//	Mime_TOML:    toml.Codec
func New() *Encoding {
	return &Encoding{
		mimeMap: map[string]codec.Marshaler{
			Mime_PostForm:          form.New("json"),
			Mime_MultipartPostForm: &form.MultipartCodec{Codec: form.New("json")},
			Mime_JSON:              &json.Codec{UseNumber: true, DisallowUnknownFields: false},
		},
		mimeQuery:    &form.QueryCodec{Codec: form.New("json")},
		mimeUri:      &form.UriCodec{Codec: form.New("json")},
		mimeWildcard: &json.Codec{UseNumber: true, DisallowUnknownFields: true},
	}
}

// Register a marshaler for a case-sensitive MIME type string
// ("*" to match any MIME type).
// you can override default marshaler with same MIME type
func (r *Encoding) Register(mime string, marshaler codec.Marshaler) error {
	if len(mime) == 0 {
		return errors.New("encoding: empty MIME type")
	}
	if marshaler == nil {
		return errors.New("encoding: marshaller should be not nil")
	}
	switch mime {
	case Mime_Query:
		m, ok := marshaler.(codec.FormMarshaler)
		if !ok {
			return errors.New("encoding: marshaller should be implement codec.FormMarshaler")
		}
		r.mimeQuery = m
	case Mime_Uri:
		m, ok := marshaler.(codec.UriMarshaler)
		if !ok {
			return errors.New("encoding: marshaller should be implement codec.UriMarshaler")
		}
		r.mimeUri = m
	case Mime_Wildcard:
		r.mimeWildcard = marshaler
	default:
		r.mimeMap[mime] = marshaler
	}
	return nil
}

// Get returns the marshalers with a case-sensitive MIME type string
// It checks the MIME type on the Encoding.
// Otherwise, it follows the above logic for "*" Marshaler.
func (r *Encoding) Get(mime string) codec.Marshaler {
	switch mime {
	case Mime_Query:
		return r.mimeQuery
	case Mime_Uri:
		return r.mimeUri
	case Mime_Wildcard:
		return r.mimeWildcard
	default:
		m := r.mimeMap[mime]
		if m == nil {
			m = r.mimeWildcard
		}
		return m
	}
}

// Delete remove the MIME type marshaler.
// MIMEWildcard, MIMEQuery, MIMEURI should be always exist and valid.
func (r *Encoding) Delete(mime string) error {
	if mime == Mime_Wildcard ||
		mime == Mime_Query ||
		mime == Mime_Uri {
		return fmt.Errorf("encoding: MIME(%s) can't delete, but you can override it", mime)
	}
	delete(r.mimeMap, mime)
	return nil
}

// InboundForRequest returns the inbound `Content-Type` and marshalers for this request.
// It checks the registry on the Encoding for the MIME type set by the `Content-Type` header.
// If it isn't set (or the request `Content-Type` is empty), checks for "*".
// If there are multiple `Content-Type` headers set, choose the first one that it can
// exactly match in the registry.
// Otherwise, it follows the above logic for "*" Marshaler.
func (r *Encoding) InboundForRequest(req *http.Request) (string, codec.Marshaler) {
	return r.marshalerFromHeaderContentType(req.Header[contentTypeHeader])
}

// OutboundForRequest returns the marshalers for this request.
// It checks the registry on the Encoding for the MIME type set by the `Accept` header.
// If it isn't set (or the request `Accept` is empty), checks for "*".
// If there are multiple `Accept` headers set, choose the first one that it can
// exactly match in the registry.
// Otherwise, it follows the above logic for "*" Marshaler.
func (r *Encoding) OutboundForRequest(req *http.Request) codec.Marshaler {
	return r.marshalerFromHeaderAccept(req.Header[acceptHeader])
}

// Bind checks the Method and Content-Type to select codec.Marshaler automatically,
// Depending on the "Content-Type" header different bind are used, for example:
//
//	"application/json" --> JSON codec.Marshaler
//	"application/xml"  --> XML codec.Marshaler
//
// It parses the request's body as JSON if Content-Type == "application/json" using JSON or XML as a JSON input.
// It decodes the json payload into the struct specified as a pointer.
func (r *Encoding) Bind(req *http.Request, v any) error {
	if req.Method == http.MethodGet {
		return r.BindQuery(req, v)
	}
	contentType, marshaller := r.InboundForRequest(req)
	if contentType == Mime_MultipartPostForm {
		m, ok := marshaller.(codec.FormCodec)
		if !ok {
			return fmt.Errorf("encoding: not supported marshaller(%v)", contentType)
		}
		if err := req.ParseMultipartForm(defaultMemory); err != nil {
			return err
		}
		return m.Decode(req.MultipartForm.Value, v)
	}
	return marshaller.NewDecoder(req.Body).
		Decode(v)
}

// BindQuery binds the passed struct pointer using the query codec.Marshaler.
func (r *Encoding) BindQuery(req *http.Request, v any) error {
	return r.mimeQuery.Decode(req.URL.Query(), v)
}

// BindUri binds the passed struct pointer using the uri codec.Marshaler.
func (r *Encoding) BindUri(raws url.Values, v any) error {
	return r.mimeUri.Decode(raws, v)
}

// Render writes the response headers and calls the outbound marshalers for this request.
// It checks the registry on the Encoding for the MIME type set by the Accept header.
// If it isn't set (or the request Accept is empty), checks for "*". for example:
//
//	"application/json" --> JSON codec.Marshaler
//	"application/xml"  --> XML codec.Marshaler
//
// If there are multiple Accept headers set, choose the first one that it can
// exactly match in the registry.
// Otherwise, it follows the above logic for "*" Marshaler.
func (r *Encoding) Render(w http.ResponseWriter, req *http.Request, v any) error {
	if v == nil {
		return nil
	}
	marshaller := r.OutboundForRequest(req)
	data, err := marshaller.Marshal(v)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", marshaller.ContentType(v))
	_, err = w.Write(data)
	return err
}

func parseAcceptHeader(header string) []string {
	// TODO: cache header maps to avoid parse again?
	values := strings.Split(header, ",")
	for i := 0; i < len(values); i++ {
		values[i] = strings.TrimSpace(values[i])
	}
	return values
}

// InboundForResponse returns the inbound marshaler for this response.
// It checks the registry on the Encoding for the MIME type set by the `Content-Type` header.
// If it isn't set (or the response `Content-Type` is empty), checks for "*".
// If there are multiple `Content-Type` headers set, choose the first one that it can
// exactly match in the registry.
// Otherwise, it follows the above logic for "*" Marshaler.
func (r *Encoding) InboundForResponse(resp *http.Response) codec.Marshaler {
	_, marshaler := r.marshalerFromHeaderContentType(resp.Header[contentTypeHeader])
	return marshaler
}

// Encode encode v use contentType
func (r *Encoding) Encode(contentType string, v any) ([]byte, error) {
	return r.Get(contentType).Marshal(v)
}

// EncodeQuery encode v to the query url.Values.
func (r *Encoding) EncodeQuery(v any) (url.Values, error) {
	return r.mimeQuery.Encode(v)
}

// EncodeUrl encode msg to url path.
// pathTemplate is a template of url path like http://helloworld.dev/{name}/sub/{sub.name},
func (r *Encoding) EncodeUrl(athTemplate string, msg any, needQuery bool) string {
	return r.mimeUri.EncodeUrl(athTemplate, msg, needQuery)
}

// marshalerFromHeaderContentType returns the `Content-Type` and marshaler from `Content-Type` header.
// It checks the registry on the Encoding for the MIME type set by the `Content-Type` header.
// If it isn't set (or the `Content-Type` is empty), checks for "*".
// If there are multiple `Content-Type` headers set, choose the first one that it can
// exactly match in the registry.
// Otherwise, it follows the above logic for "*" Marshaler.
func (r *Encoding) marshalerFromHeaderContentType(values []string) (string, codec.Marshaler) {
	var err error
	var marshaler codec.Marshaler
	var contentType string

	for _, contentTypeVal := range values {
		contentType, _, err = mime.ParseMediaType(contentTypeVal)
		if err != nil {
			continue
		}
		if m, ok := r.mimeMap[contentType]; ok {
			marshaler = m
			break
		}
	}
	if marshaler == nil {
		contentType = Mime_Wildcard
		marshaler = r.mimeWildcard
	}
	return contentType, marshaler
}

// marshalerFromHeaderAccept returns the marshalers from `Accept` header.
// It checks the registry on the Encoding for the MIME type set by the `Accept` header.
// If it isn't set (or the `Accept` is empty), checks for "*".
// If there are multiple `Accept` headers set, choose the first one that it can
// exactly match in the registry.
// Otherwise, it follows the above logic for "*" Marshaler.
func (r *Encoding) marshalerFromHeaderAccept(values []string) codec.Marshaler {
	var marshaler codec.Marshaler

	for _, acceptVal := range values {
		headerValues := parseAcceptHeader(acceptVal)
		for _, value := range headerValues {
			if m, ok := r.mimeMap[value]; ok {
				marshaler = m
				break
			}
		}
	}
	if marshaler == nil {
		marshaler = r.mimeWildcard
	}
	return marshaler
}
