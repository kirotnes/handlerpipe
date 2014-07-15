package handlerpipe

import (
	"net/http"
)

type responseWriterWrapper struct {
	wrappedWriter http.ResponseWriter
	written       bool
}

func (w *responseWriterWrapper) Header() http.Header {
	return w.wrappedWriter.Header()
}

func (w *responseWriterWrapper) Write(bytes []byte) (int, error) {
	w.written = true
	return w.wrappedWriter.Write(bytes)
}

func (w *responseWriterWrapper) WriteHeader(statusCode int) {
	w.wrappedWriter.WriteHeader(statusCode)
}

func New() *handleChain {

	return &handleChain{nil}
}

type handleChain struct {
	funcs []http.HandlerFunc
}

func (hc *handleChain) AddFuncs(handlerfunc ...http.HandlerFunc) *handleChain {
	hc.funcs = append(hc.funcs, handlerfunc...)
	return hc
}

func (hc *handleChain) AddHandlers(handler ...http.Handler) *handleChain {

	for _, handler := range handler {
		handlerFunc := func(w http.ResponseWriter, req *http.Request) {
			handler.ServeHTTP(w, req)
		}

		hc.funcs = append(hc.funcs, handlerFunc)
	}
	return hc
}

func (hc *handleChain) UnwrapHandlerFunc() http.HandlerFunc {

	return func(w http.ResponseWriter, req *http.Request) {

		var customResponseWriter = &responseWriterWrapper{w, false}

		defer func() {

			if err := recover(); err != nil {
				w.WriteHeader(500)
				w.Write([]byte("Internal server error"))
			}
		}()

		for _, handler := range hc.funcs {
			handler(customResponseWriter, req)

			if customResponseWriter.written {
				break
			}
		}
	}
}

func (hc *handleChain) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	hc.UnwrapHandlerFunc()(w, req)
}

type handlerListTemplate struct {
	funcs []http.HandlerFunc
}

func NewTemplate() *handlerListTemplate {

	return &handlerListTemplate{}
}

func (template *handlerListTemplate) AddFuncs(handlerfunc ...http.HandlerFunc) *handlerListTemplate {
	template.funcs = append(template.funcs, handlerfunc...)
	return template
}

func (template *handlerListTemplate) AddHandlers(handler ...http.Handler) *handlerListTemplate {

	for _, handler := range handler {
		handlerFunc := func(w http.ResponseWriter, req *http.Request) {
			handler.ServeHTTP(w, req)
		}

		template.funcs = append(template.funcs, handlerFunc)
	}
	return template
}

func (template *handlerListTemplate) ToChain(handler ...http.Handler) *handleChain {

	return New().AddFuncs(template.funcs...)
}
