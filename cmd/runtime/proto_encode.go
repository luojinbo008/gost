package runtime

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type Path string

func (path *Path) Replace(name, value string) {
	if path == nil {
		return
	}
	tmp := strings.Replace(string(*path), fmt.Sprintf(":%s", name), value, 1)
	*path = Path(tmp)
}

func (path *Path) String() string {
	if path == nil {
		return ""
	}
	return string(*path)
}

type RequestBody struct {
	query  url.Values
	header http.Header
	path   *Path
}

func (req *RequestBody) GetQuery() url.Values {
	return req.query
}

func (req *RequestBody) GetHeader() http.Header {
	return req.header
}

func (req *RequestBody) GetPath() string {
	return req.path.String()
}

var timeType = reflect.TypeOf(time.Time{})

var encoderType = reflect.TypeOf(new(Encoder)).Elem()

type Encoder interface {
	EncodeHeaders(key string, v *http.Header) error
	EncodeQuery(key string, v *url.Values) error

	EncodePath(key string, v *Path) error
}

func Values(path Path, v interface{}) (*RequestBody, error) {
	request := &RequestBody{
		query:  make(url.Values),
		header: make(http.Header),
		path:   &path,
	}

	val := reflect.ValueOf(v)
	for val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return request, nil
		}
		val = val.Elem()
	}

	if v == nil {
		return request, nil
	}

	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("query: Values() expects struct input. Got %v", val.Kind())
	}

	if err := reflectPath(request.path, val, ""); err != nil {
		return request, err
	}

	if err := reflectQuery(request.query, val, ""); err != nil {
		return request, err
	}

	if err := reflectHeader(request.header, val, ""); err != nil {
		return request, err
	}

	return request, nil
}

func reflectPath(value *Path, val reflect.Value, scope string) error {
	var embedded []reflect.Value
	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		sf := typ.Field(i)
		if sf.PkgPath != "" && !sf.Anonymous { // unexported
			continue
		}
		sv := val.Field(i)
		tag := sf.Tag.Get("param")
		if tag == "-" || tag == "" {
			continue
		}
		name, opts := parseTag(tag)

		if name == "" {
			if sf.Anonymous {
				v := reflect.Indirect(sv)
				if v.IsValid() && v.Kind() == reflect.Struct {
					// save embedded struct for later processing
					embedded = append(embedded, v)
					continue
				}
			}
			name = sf.Name
		}

		if scope != "" {
			name = scope + "[" + name + "]"
		}

		if opts.Contains("omitempty") && isEmptyValue(sv) {
			continue
		}

		if sv.Type().Implements(encoderType) {
			// if sv is a nil pointer and the custom encoder is defined on a non-pointer
			// method receiver, set sv to the zero value of the underlying type
			if !reflect.Indirect(sv).IsValid() && sv.Type().Elem().Implements(encoderType) {
				sv = reflect.New(sv.Type().Elem())
			}

			m := sv.Interface().(Encoder)
			if err := m.EncodePath(name, value); err != nil {
				return err
			}
			continue
		}

		// recursively dereference pointers. break on nil pointers
		for sv.Kind() == reflect.Ptr {
			if sv.IsNil() {
				break
			}
			sv = sv.Elem()
		}

		if sv.Kind() == reflect.Slice || sv.Kind() == reflect.Array {
			if sv.Len() == 0 {
				// skip if slice or array is empty
				continue
			}

			var del string
			if opts.Contains("comma") {
				del = ","
			} else if opts.Contains("space") {
				del = " "
			} else if opts.Contains("semicolon") {
				del = ";"
			} else if opts.Contains("brackets") {
				name = name + "[]"
			} else {
				del = sf.Tag.Get("del")
			}

			if del != "" {
				s := new(bytes.Buffer)
				first := true
				for i := 0; i < sv.Len(); i++ {
					if first {
						first = false
					} else {
						s.WriteString(del)
					}
					s.WriteString(valueString(sv.Index(i), opts, sf))
				}
				value.Replace(name, s.String())
			} else {
				for i := 0; i < sv.Len(); i++ {
					k := name
					if opts.Contains("numbered") {
						k = fmt.Sprintf("%s%d", name, i)
					}
					value.Replace(k, valueString(sv.Index(i), opts, sf))
				}
			}
			continue
		}

		if sv.Type() == timeType {
			value.Replace(name, valueString(sv, opts, sf))
			continue
		}

		if sv.Kind() == reflect.Struct {
			if err := reflectPath(value, sv, name); err != nil {
				return err
			}
			continue
		}
		value.Replace(name, valueString(sv, opts, sf))
	}

	for _, f := range embedded {
		if err := reflectPath(value, f, scope); err != nil {
			return err
		}
	}

	return nil
}

func reflectQuery(values url.Values, val reflect.Value, scope string) error {
	var embedded []reflect.Value
	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		sf := typ.Field(i)
		if sf.PkgPath != "" && !sf.Anonymous { // unexported
			continue
		}
		sv := val.Field(i)
		tag := sf.Tag.Get("query")
		if tag == "-" || tag == "" {
			continue
		}
		name, opts := parseTag(tag)
		if name == "" {
			if sf.Anonymous {
				v := reflect.Indirect(sv)
				if v.IsValid() && v.Kind() == reflect.Struct {
					// save embedded struct for later processing
					embedded = append(embedded, v)
					continue
				}
			}
			name = sf.Name
		}

		if scope != "" {
			name = scope + "[" + name + "]"
		}

		if opts.Contains("omitempty") && isEmptyValue(sv) {
			continue
		}

		// todo
		if sv.Type().Implements(encoderType) {
			// if sv is a nil pointer and the custom encoder is defined on a non-pointer
			// method receiver, set sv to the zero value of the underlying type
			if !reflect.Indirect(sv).IsValid() && sv.Type().Elem().Implements(encoderType) {
				sv = reflect.New(sv.Type().Elem())
			}

			m := sv.Interface().(Encoder)
			if err := m.EncodeQuery(name, &values); err != nil {
				return err
			}
			continue
		}

		// recursively dereference pointers. break on nil pointers
		for sv.Kind() == reflect.Ptr {
			if sv.IsNil() {
				break
			}
			sv = sv.Elem()
		}

		if sv.Kind() == reflect.Slice || sv.Kind() == reflect.Array {
			if sv.Len() == 0 {
				// skip if slice or array is empty
				continue
			}

			var del string
			if opts.Contains("comma") {
				del = ","
			} else if opts.Contains("space") {
				del = " "
			} else if opts.Contains("semicolon") {
				del = ";"
			} else if opts.Contains("brackets") {
				name = name + "[]"
			} else {
				del = sf.Tag.Get("del")
			}

			if del != "" {
				s := new(bytes.Buffer)
				first := true
				for i := 0; i < sv.Len(); i++ {
					if first {
						first = false
					} else {
						s.WriteString(del)
					}
					s.WriteString(valueString(sv.Index(i), opts, sf))
				}
				values.Add(name, s.String())
			} else {
				for i := 0; i < sv.Len(); i++ {
					k := name
					if opts.Contains("numbered") {
						k = fmt.Sprintf("%s%d", name, i)
					}
					values.Add(k, valueString(sv.Index(i), opts, sf))
				}
			}
			continue
		}

		if sv.Type() == timeType {
			values.Add(name, valueString(sv, opts, sf))
			continue
		}

		if sv.Kind() == reflect.Struct {
			if err := reflectQuery(values, sv, name); err != nil {
				return err
			}
			continue
		}

		values.Add(name, valueString(sv, opts, sf))
	}

	for _, f := range embedded {
		if err := reflectQuery(values, f, scope); err != nil {
			return err
		}
	}

	return nil
}

func reflectHeader(values http.Header, val reflect.Value, scope string) error {
	var embedded []reflect.Value
	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		sf := typ.Field(i)
		if sf.PkgPath != "" && !sf.Anonymous { // unexported
			continue
		}
		sv := val.Field(i)
		tag := sf.Tag.Get("header")
		if tag == "-" || tag == "" {
			continue
		}
		name, opts := parseTag(tag)
		if name == "" {
			if sf.Anonymous {
				v := reflect.Indirect(sv)
				if v.IsValid() && v.Kind() == reflect.Struct {
					// save embedded struct for later processing
					embedded = append(embedded, v)
					continue
				}
			}
			name = sf.Name
		}

		if scope != "" {
			name = scope + "[" + name + "]"
		}

		if opts.Contains("omitempty") && isEmptyValue(sv) {
			continue
		}

		if sv.Type().Implements(encoderType) {
			// if sv is a nil pointer and the custom encoder is defined on a non-pointer
			// method receiver, set sv to the zero value of the underlying type
			if !reflect.Indirect(sv).IsValid() && sv.Type().Elem().Implements(encoderType) {
				sv = reflect.New(sv.Type().Elem())
			}

			m := sv.Interface().(Encoder)
			if err := m.EncodeHeaders(name, &values); err != nil {
				return err
			}
			continue
		}

		// recursively dereference pointers. break on nil pointers
		for sv.Kind() == reflect.Ptr {
			if sv.IsNil() {
				break
			}
			sv = sv.Elem()
		}

		if sv.Kind() == reflect.Slice || sv.Kind() == reflect.Array {
			if sv.Len() == 0 {
				// skip if slice or array is empty
				continue
			}

			var del string
			if opts.Contains("comma") {
				del = ","
			} else if opts.Contains("space") {
				del = " "
			} else if opts.Contains("semicolon") {
				del = ";"
			} else if opts.Contains("brackets") {
				name = name + "[]"
			} else {
				del = sf.Tag.Get("del")
			}

			if del != "" {
				s := new(bytes.Buffer)
				first := true
				for i := 0; i < sv.Len(); i++ {
					if first {
						first = false
					} else {
						s.WriteString(del)
					}
					s.WriteString(valueString(sv.Index(i), opts, sf))
				}
				values.Add(name, s.String())
			} else {
				for i := 0; i < sv.Len(); i++ {
					k := name
					if opts.Contains("numbered") {
						k = fmt.Sprintf("%s%d", name, i)
					}
					values.Add(k, valueString(sv.Index(i), opts, sf))
				}
			}
			continue
		}

		if sv.Type() == timeType {
			values.Add(name, valueString(sv, opts, sf))
			continue
		}

		if sv.Kind() == reflect.Struct {
			if err := reflectHeader(values, sv, name); err != nil {
				return err
			}
			continue
		}

		values.Add(name, valueString(sv, opts, sf))
	}

	for _, f := range embedded {
		if err := reflectHeader(values, f, scope); err != nil {
			return err
		}
	}

	return nil
}

// func reflectUel(request *RequestBody, val reflect.Value, scope string) error {
// 	var embedded []reflect.Value
// 	typ := val.Type()
// 	for i := 0; i < typ.NumField(); i++ {
// 		sf := typ.Field(i)
// 		if sf.PkgPath != "" && !sf.Anonymous { // unexported
// 			continue
// 		}
// 		sv := val.Field(i)

// 		// 解析 query
// 		// 解析 header

// 		tag := sf.Tag.Get("query")

// 		if tag == "-" {
// 			continue
// 		}

// 		values := request.query
// 		name, opts := parseTag(tag)
// 		//
// 		if name == "" {
// 			if sf.Anonymous {
// 				v := reflect.Indirect(sv)
// 				if v.IsValid() && v.Kind() == reflect.Struct {
// 					// save embedded struct for later processing
// 					embedded = append(embedded, v)
// 					continue
// 				}
// 			}
// 			name = sf.Name
// 		}

// 		if scope != "" {
// 			name = scope + "[" + name + "]"
// 		}

// 		if opts.Contains("omitempty") && isEmptyValue(sv) {
// 			continue
// 		}

// 		if sv.Type().Implements(encoderType) {
// 			// if sv is a nil pointer and the custom encoder is defined on a non-pointer
// 			// method receiver, set sv to the zero value of the underlying type
// 			if !reflect.Indirect(sv).IsValid() && sv.Type().Elem().Implements(encoderType) {
// 				sv = reflect.New(sv.Type().Elem())
// 			}

// 			m := sv.Interface().(Encoder)
// 			if err := m.EncodeValues(name, &values); err != nil {
// 				return err
// 			}
// 			continue
// 		}

// 		// recursively dereference pointers. break on nil pointers
// 		for sv.Kind() == reflect.Ptr {
// 			if sv.IsNil() {
// 				break
// 			}
// 			sv = sv.Elem()
// 		}

// 		if sv.Kind() == reflect.Slice || sv.Kind() == reflect.Array {
// 			if sv.Len() == 0 {
// 				// skip if slice or array is empty
// 				continue
// 			}

// 			var del string
// 			if opts.Contains("comma") {
// 				del = ","
// 			} else if opts.Contains("space") {
// 				del = " "
// 			} else if opts.Contains("semicolon") {
// 				del = ";"
// 			} else if opts.Contains("brackets") {
// 				name = name + "[]"
// 			} else {
// 				del = sf.Tag.Get("del")
// 			}

// 			if del != "" {
// 				s := new(bytes.Buffer)
// 				first := true
// 				for i := 0; i < sv.Len(); i++ {
// 					if first {
// 						first = false
// 					} else {
// 						s.WriteString(del)
// 					}
// 					s.WriteString(valueString(sv.Index(i), opts, sf))
// 				}
// 				values.Add(name, s.String())
// 			} else {
// 				for i := 0; i < sv.Len(); i++ {
// 					k := name
// 					if opts.Contains("numbered") {
// 						k = fmt.Sprintf("%s%d", name, i)
// 					}
// 					values.Add(k, valueString(sv.Index(i), opts, sf))
// 				}
// 			}
// 			continue
// 		}

// 		if sv.Type() == timeType {
// 			values.Add(name, valueString(sv, opts, sf))
// 			continue
// 		}

// 		if sv.Kind() == reflect.Struct {
// 			if err := reflectValue(request, sv, name); err != nil {
// 				return err
// 			}
// 			continue
// 		}

// 		values.Add(name, valueString(sv, opts, sf))
// 	}

// 	for _, f := range embedded {
// 		if err := reflectValue(request, f, scope); err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

// valueString returns the string representation of a value.
func valueString(v reflect.Value, opts tagOptions, sf reflect.StructField) string {
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return ""
		}
		v = v.Elem()
	}

	if v.Kind() == reflect.Bool && opts.Contains("int") {
		if v.Bool() {
			return "1"
		}
		return "0"
	}

	if v.Type() == timeType {
		t := v.Interface().(time.Time)
		if opts.Contains("unix") {
			return strconv.FormatInt(t.Unix(), 10)
		}
		if opts.Contains("unixmilli") {
			return strconv.FormatInt((t.UnixNano() / 1e6), 10)
		}
		if opts.Contains("unixnano") {
			return strconv.FormatInt(t.UnixNano(), 10)
		}
		if layout := sf.Tag.Get("layout"); layout != "" {
			return t.Format(layout)
		}
		return t.Format(time.RFC3339)
	}

	return fmt.Sprint(v.Interface())
}

// isEmptyValue checks if a value should be considered empty for the purposes
// of omitting fields with the "omitempty" option.
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}

	type zeroable interface {
		IsZero() bool
	}

	if z, ok := v.Interface().(zeroable); ok {
		return z.IsZero()
	}

	return false
}

// tagOptions is the string following a comma in a struct field's "url" tag, or
// the empty string. It does not include the leading comma.
type tagOptions []string

// parseTag splits a struct field's url tag into its name and comma-separated
// options.
func parseTag(tag string) (string, tagOptions) {
	s := strings.Split(tag, ",")
	return s[0], s[1:]
}

// Contains checks whether the tagOptions contains the specified option.
func (o tagOptions) Contains(option string) bool {
	for _, s := range o {
		if s == option {
			return true
		}
	}
	return false
}
