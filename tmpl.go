package main

import (
	"bytes"
	htmpl "html/template"
	"net/http"
	"net/url"
)

const TemplatesDirName = "./templates/"

var Tmpls *htmpl.Template

func ParseTemplates() {
	Tmpls = htmpl.Must(htmpl.New("index.tmpl").ParseGlob(TemplatesDirName + "*.tmpl"))
}

func WriteTemplate(w http.ResponseWriter, tmpl string, respCode int, payload url.Values, err error) error {
	response := new(bytes.Buffer)
	if e := Tmpls.ExecuteTemplate(response, tmpl, struct {
		Error   error
		Payload url.Values
	}{err, payload}); e != nil {
		return WrapErrorWithTrace(e)
	}

	w.WriteHeader(respCode)
	if _, err := w.Write(response.Bytes()); err != nil {
		return WrapErrorWithTrace(err)
	}
	return nil
}

func WriteTemplateAny[T any](w http.ResponseWriter, tmpl string, respCode int, s T) error {
	response := new(bytes.Buffer)
	if e := Tmpls.ExecuteTemplate(response, tmpl, s); e != nil {
		return WrapErrorWithTraceSkip(e, 2)
	}
	w.WriteHeader(respCode)
	if _, err := w.Write(response.Bytes()); err != nil {
		return WrapErrorWithTraceSkip(err, 2)
	}
	return nil
}
