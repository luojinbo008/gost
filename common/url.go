package common

import (
	"bytes"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jinzhu/copier"
	"github.com/luojinbo008/gost/common/constant"
	gxset "github.com/luojinbo008/gost/utils/container/set"
	perrors "github.com/pkg/errors"
)

const (
	PROTOCOL = "protocol"
)

var (
	compareURLEqualFunc CompareURLEqualFunc
)

func init() {
	compareURLEqualFunc = defaultCompareURLEqual
}

// 基础url
type baseURL struct {
	Protocol string // 协议
	Location string // ip + port
	Ip       string // ip
	Port     string // port

	PrimitiveURL string // 原始url 预留
}

type noCopy struct{}

func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}

type URL struct {
	noCopy noCopy

	baseURL

	// url.Values is not safe map, add to avoid concurrent map read and map write error
	paramsLock sync.RWMutex
	params     url.Values

	Path string

	// special for registry
	SubURL *URL
}

type Option func(*URL)

func NewURL(urlString string, opts ...Option) (*URL, error) {
	s := URL{baseURL: baseURL{}}

	if urlString == "" {
		return &s, nil
	}

	rawURLString, err := url.QueryUnescape(urlString)
	if err != nil {
		return &s, perrors.Errorf("URL.QueryUnescape(%s),  error{%v}", urlString, err)
	}

	// rawURLString = "//" + rawURLString
	if !strings.Contains(rawURLString, "//") {
		t := URL{baseURL: baseURL{}}
		for _, opt := range opts {
			opt(&t)
		}
		rawURLString = t.Protocol + "://" + rawURLString
	}

	serviceURL, urlParseErr := url.Parse(rawURLString)
	if urlParseErr != nil {
		return &s, perrors.Errorf("URL.Parse(URL string{%s}),  error{%v}", rawURLString, err)
	}

	s.params, err = url.ParseQuery(serviceURL.RawQuery)
	if err != nil {
		return &s, perrors.Errorf("URL.ParseQuery(raw URL string{%s}),  error{%v}", serviceURL.RawQuery, err)
	}

	s.PrimitiveURL = urlString
	s.Protocol = serviceURL.Scheme
	s.Location = serviceURL.Host
	s.Path = serviceURL.Path
	for _, location := range strings.Split(s.Location, ",") {
		location = strings.Trim(location, " ")
		if strings.Contains(location, ":") {
			s.Ip, s.Port, err = net.SplitHostPort(location)
			if err != nil {
				return &s, perrors.Errorf("net.SplitHostPort(url.Host{%s}), error{%v}", s.Location, err)
			}
			break
		}
	}

	for _, opt := range opts {
		opt(&s)
	}
	return &s, nil
}

// CloneWithParams Copy URL based on the reserved parameter's keys.
func (c *URL) CloneWithParams(reserveParams []string) *URL {
	params := url.Values{}
	for _, reserveParam := range reserveParams {
		v := c.GetParam(reserveParam, "")
		if len(v) != 0 {
			params.Set(reserveParam, v)
		}
	}

	return NewURLWithOptions(
		WithProtocol(c.Protocol),
		WithIp(c.Ip),
		WithPort(c.Port),
		WithPath(c.Path),
		WithParams(params),
	)
}

// GetParams gets values
func (c *URL) GetParams() url.Values {
	return c.params
}

// GetParam gets value by key
func (c *URL) GetParam(s string, d string) string {
	c.paramsLock.RLock()
	defer c.paramsLock.RUnlock()

	var r string
	if len(c.params) > 0 {
		r = c.params.Get(s)
	}
	if len(r) == 0 {
		r = d
	}

	return r
}

// GetParamDuration get duration if param is invalid or missing will return 3s
func (c *URL) GetParamDuration(s string, d string) time.Duration {
	if t, err := time.ParseDuration(c.GetParam(s, d)); err == nil {
		return t
	}
	return 3 * time.Second
}

// GetParamInt gets int64 value by @key
func (c *URL) GetParamInt(key string, d int64) int64 {
	r, err := strconv.ParseInt(c.GetParam(key, ""), 10, 64)
	if err != nil {
		return d
	}
	return r
}

// GetParamByIntValue gets int value by @key
func (c *URL) GetParamByIntValue(key string, d int) int {
	r, err := strconv.ParseInt(c.GetParam(key, ""), 10, 0)
	if err != nil {
		return d
	}
	return int(r)
}

// GetParamBool judge whether @key exists or not
func (c *URL) GetParamBool(key string, d bool) bool {
	r, err := strconv.ParseBool(c.GetParam(key, ""))
	if err != nil {
		return d
	}
	return r
}

func (c *URL) SetParam(key string, value string) {
	c.paramsLock.Lock()
	defer c.paramsLock.Unlock()
	if c.params == nil {
		c.params = url.Values{}
	}
	c.params.Set(key, value)
}

// Clone will copy the URL
func (c *URL) Clone() *URL {
	newURL := &URL{}
	if err := copier.Copy(newURL, c); err != nil {
		// this is impossible
		return newURL
	}
	newURL.params = url.Values{}
	c.RangeParams(func(key, value string) bool {
		newURL.SetParam(key, value)
		return true
	})

	return newURL
}

// ToMap transfer URL to Map
func (c *URL) ToMap() map[string]string {
	paramsMap := make(map[string]string)

	c.RangeParams(func(key, value string) bool {
		paramsMap[key] = value
		return true
	})

	if c.Protocol != "" {
		paramsMap[PROTOCOL] = c.Protocol
	}

	if c.Location != "" {
		paramsMap["host"] = strings.Split(c.Location, ":")[0]
		var port string
		if strings.Contains(c.Location, ":") {
			port = strings.Split(c.Location, ":")[1]
		} else {
			port = "0"
		}
		paramsMap["port"] = port
	}
	if c.Protocol != "" {
		paramsMap[PROTOCOL] = c.Protocol
	}
	if c.Path != "" {
		paramsMap["path"] = c.Path
	}
	if len(paramsMap) == 0 {
		return nil
	}
	return paramsMap
}

// RangeParams will iterate the params
func (c *URL) RangeParams(f func(key, value string) bool) {
	c.paramsLock.RLock()
	defer c.paramsLock.RUnlock()
	for k, v := range c.params {
		if !f(k, v[0]) {
			break
		}
	}
}

// AddParam will add the key-value pair
func (c *URL) AddParam(key string, value string) {
	c.paramsLock.Lock()
	defer c.paramsLock.Unlock()
	if c.params == nil {
		c.params = url.Values{}
	}
	c.params.Add(key, value)
}

// todo: 规则要重新定义
// GetCacheInvokerMapKey get directory cacheInvokerMap key
func (c *URL) GetCacheInvokerMapKey() string {
	urlNew, _ := NewURL(c.PrimitiveURL)

	buildString := fmt.Sprintf("%s://%s:%s/?interface=%s&group=%s&version=%s&timestamp=%s",
		c.Protocol, c.Ip, c.Port, c.Service(), c.GetParam(constant.GroupKey, ""),
		c.GetParam(constant.VersionKey, ""), urlNew.GetParam(constant.TimestampKey, ""))
	return buildString
}

// 获得服务名 ｜ query 下标 constant.InterfaceKey ｜ path
func (c *URL) Service() string {
	service := c.GetParam(constant.InterfaceKey, strings.TrimPrefix(c.Path, "/"))
	if service != "" {
		return service
	} else if c.SubURL != nil { // 兼容注册url
		service = c.SubURL.GetParam(constant.InterfaceKey, strings.TrimPrefix(c.Path, "/"))
		if service != "" { // if URL.path is "" then return suburl's path, special for registry URL
			return service
		}
	}
	return ""
}

// ServiceKey gets a unique key of a service.
func (c *URL) ServiceKey() string {
	return ServiceKey(c.GetParam(constant.InterfaceKey, strings.TrimPrefix(c.Path, constant.PathSeparator)),
		c.GetParam(constant.GroupKey, ""), c.GetParam(constant.VersionKey, ""))
}

// Group get group
func (c *URL) Group() string {
	return c.GetParam(constant.GroupKey, "")
}

// Version get version
func (c *URL) Version() string {
	return c.GetParam(constant.VersionKey, "")
}

// URL to string
func (c *URL) String() string {
	c.paramsLock.Lock()
	defer c.paramsLock.Unlock()
	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("%s://%s:%s%s?", c.Protocol, c.Ip, c.Port, c.Path))
	buf.WriteString(c.params.Encode())
	return buf.String()
}

// 替换url parmas
func (c *URL) ReplaceParams(param url.Values) {
	c.paramsLock.Lock()
	defer c.paramsLock.Unlock()
	c.params = param
}

// Key gets key
func (c *URL) Key() string {
	buildString := fmt.Sprintf("%s://%s:%s/?interface=%s&group=%s&version=%s",
		c.Protocol, c.Ip, c.Port, c.Service(), c.GetParam(constant.GroupKey, ""), c.GetParam(constant.VersionKey, ""))
	return buildString
}

func (c *URL) CloneExceptParams(excludeParams *gxset.HashSet) *URL {
	newURL := &URL{}
	if err := copier.Copy(newURL, c); err != nil {
		// this is impossible
		return newURL
	}
	newURL.params = url.Values{}
	c.RangeParams(func(key, value string) bool {
		if !excludeParams.Contains(key) {
			newURL.SetParam(key, value)
		}
		return true
	})
	return newURL
}

// NewURLWithOptions will create a new URL with options
func NewURLWithOptions(opts ...Option) *URL {
	newURL := &URL{}
	for _, opt := range opts {
		opt(newURL)
	}
	newURL.Location = newURL.Ip + ":" + newURL.Port
	return newURL
}

// WithProtocol sets proto for URL
func WithProtocol(proto string) Option {
	return func(url *URL) {
		url.Protocol = proto
	}
}

// WithIp sets ip for URL
func WithIp(ip string) Option {
	return func(url *URL) {
		url.Ip = ip
	}
}

// WithPort sets port for URL
func WithPort(port string) Option {
	return func(url *URL) {
		url.Port = port
	}
}

// WithPath sets path for URL
func WithPath(path string) Option {
	return func(url *URL) {
		url.Path = "/" + strings.TrimPrefix(path, "/")
	}
}

func WithLocation(location string) Option {
	return func(url *URL) {
		url.Location = location
	}
}

func WithParams(params url.Values) Option {
	return func(url *URL) {
		url.params = params
	}
}

// WithParamsValue sets params field for URL
func WithParamsValue(key, val string) Option {
	return func(url *URL) {
		url.SetParam(key, val)
	}
}

// service key 拼接规则
func ServiceKey(name string, group string, version string) string {
	if name == "" {
		return ""
	}
	buf := &bytes.Buffer{}
	if group != "" {
		buf.WriteString(group)
		buf.WriteString("/")
	}

	buf.WriteString(name)

	if version != "" && version != "0.0.0" {
		buf.WriteString(":")
		buf.WriteString(version)
	}

	return buf.String()
}

type CompareURLEqualFunc func(l *URL, r *URL, excludeParam ...string) bool

func SetCompareURLEqualFunc(f CompareURLEqualFunc) {
	compareURLEqualFunc = f
}

func GetCompareURLEqualFunc() CompareURLEqualFunc {
	return compareURLEqualFunc
}

func defaultCompareURLEqual(l *URL, r *URL, excludeParam ...string) bool {
	return IsEquals(l, r, excludeParam...)
}

// IsEquals compares if two URLs equals with each other. Excludes are all parameter keys which should ignored.
func IsEquals(left *URL, right *URL, excludes ...string) bool {
	if (left == nil && right != nil) || (right == nil && left != nil) {
		return false
	}
	if left.Ip != right.Ip || left.Port != right.Port {
		return false
	}

	leftMap := left.ToMap()
	rightMap := right.ToMap()
	for _, exclude := range excludes {
		delete(leftMap, exclude)
		delete(rightMap, exclude)
	}

	if len(leftMap) != len(rightMap) {
		return false
	}

	for lk, lv := range leftMap {
		if rv, ok := rightMap[lk]; !ok {
			return false
		} else if lv != rv {
			return false
		}
	}

	return true
}
