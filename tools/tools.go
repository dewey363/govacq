package tools

import (
	"github.com/araddon/dateparse"
	"github.com/pkg/errors"
	"golang.org/x/text/encoding/simplifiedchinese"
	"net/url"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

const SplitStr = "★"

var GBKTrans = simplifiedchinese.GBK

func CleanHref(oriUrl string, href string, addSlash bool) *url.URL {
	root, _ := url.Parse(oriUrl)
	if strings.Contains(strings.ToLower(href), "javascript") {
		return nil
	}
	ref, _ := root.Parse(href)
	reference := root.ResolveReference(ref)
	reference.Fragment = ""
	reference.Path, _ = url.PathUnescape(ref.Path)
	if !addSlash || len(reference.RawQuery) > 0 {
		goto exit
	}
	if strings.ContainsRune(reference.Path, '.') || strings.HasSuffix(reference.Path, "/") {
		goto exit
	}
	reference.Path += "/"
exit:
	return reference
}

type URLCounterMatcher struct {
	Count   int
	Expr    string
	Example string
}

func SplitURL0(urls []string, mapSize int) ([]URLCounterMatcher, error) {
	var (
		pattern         *regexp.Regexp
		unExpectedCount = 0
		expr            string
		matcher         URLCounterMatcher
	)
	for _, str := range urls {
		URL, _ := url.Parse(str)
		base := path.Base(URL.Path)
		if pattern != nil {
			if !pattern.MatchString(base) {
				unExpectedCount++
			} else {
				if len(matcher.Example) == 0 {
					matcher.Example = str
				}
			}
			break
		}
		var suffix string

		if index := strings.IndexRune(base, '.'); index > 0 {
			suffix = base[index+1:]
			if index = strings.IndexFunc(suffix, func(r rune) bool {
				return !unicode.IsLetter(r)
			}); index > 0 {
				suffix = suffix[:index]
			}
			suffix = `\.` + suffix
		}

		for _, r := range base {
			if unicode.IsDigit(r) {
				break
			}
			expr += string(r)
		}
		expr = "^" + expr + ".*" + suffix
		pattern = regexp.MustCompile(expr)
	}
	if unExpectedCount > mapSize+50 {
		return nil, errors.New("too many unexpected url! ")
	}

	matcher.Expr = "/" + strings.TrimLeft(expr, "^")
	matcher.Count = len(urls) - unExpectedCount
	return []URLCounterMatcher{matcher}, nil
}

func SplitURL1(urls []string) []URLCounterMatcher {
	deepMatchers := func(matchers []URLCounterMatcher, delta int) []URLCounterMatcher {
		m := make(map[string]URLCounterMatcher, len(matchers))

		for _, matcher := range matchers {
			splitStrs := strings.FieldsFunc(matcher.Expr, func(r rune) bool {
				return r == '/'
			})
			if len(splitStrs) > 3 {
				var suffix string
				if strings.HasSuffix(matcher.Expr, "/") {
					suffix = "/"
				}
				index := len(splitStrs) - 1 + delta
				tmp := splitStrs[index]
				splitStrs[index] = `.*`
				if pointIndex := strings.IndexRune(tmp, '.'); pointIndex > 0 {
					splitStrs[index] += tmp[pointIndex:]
				}
				newExpr := strings.Join(splitStrs, "/") + suffix
				if v, ok := m[newExpr]; !ok {
					m[newExpr] = matcher
				} else {
					v.Count += matcher.Count
					m[newExpr] = v
				}
			} else {
				m[matcher.Expr] = matcher
			}
		}

		matchers = make([]URLCounterMatcher, 0, len(m))
		for k, v := range m {
			v.Expr = k
			matchers = append(matchers, v)
		}

		sort.Slice(matchers, func(i, j int) bool {
			return matchers[i].Count > matchers[j].Count
		})
		return matchers
	}

	deepMatchers2 := func(matchers []URLCounterMatcher) []URLCounterMatcher {
		index := 0
		for index < len(matchers) {
			reg := regexp.MustCompile(matchers[index].Expr)

			for i := index + 1; i < len(matchers); i++ {
				if reg.MatchString(matchers[i].Expr) {
					matchers[index].Count++
					matchers = append(matchers[0:i], matchers[i+1:]...)
					i--
				}
			}
			index++
		}
		return matchers

	}

	setMatcherExpr := func(matcher *URLCounterMatcher, URL *url.URL) {
		urlRegex := URL.Path
		matcher.Count++
		var suffix string
		if index := strings.IndexRune(urlRegex, '.'); index > 0 {
			suffix = urlRegex[index:]
			if _index := strings.IndexFunc(suffix[1:], func(r rune) bool {
				return !unicode.IsLetter(r)
			}); _index > 0 {
				suffix = suffix[0 : _index+1]
			}

			urlRegex = urlRegex[0:index]
		}
		base := path.Base(urlRegex)
		var newBase string
		if len(base) > 9 {
			newBase = `.*`
			if strings.HasSuffix(URL.Path, "/") {
				newBase += "/"
			}
		} else {
			replacers := make([]string, 0)
			tmp := ""
			for _, b := range base {
				if unicode.IsDigit(b) {
					tmp += string(b)
				} else {
					if len(tmp) > 0 {
						replacers = append(replacers, tmp, `\d+`)
						tmp = ""
					}
				}
			}
			if len(tmp) > 0 {
				replacers = append(replacers, tmp, `\d+`)
			}
			newBase = strings.NewReplacer(replacers...).Replace(base)
		}
		matcher.Expr = URL.Hostname() + strings.Replace(urlRegex, base, newBase, 1) + suffix
	}

	sort.Strings(urls)
	var matchers []URLCounterMatcher
A:
	for i := 0; i < len(urls); i++ {
		example := urls[i]
		var matcher = URLCounterMatcher{Example: example}
		URL, _ := url.Parse(example)
		if len(URL.Path) == 0 || URL.Path == "/" {
			continue A
		}
		isMatch := false
		var matcherIndex int
		for j, v := range matchers {
			reg := regexp.MustCompile(v.Expr)
			if reg.MatchString(example) {
				if strings.ContainsAny(v.Expr, "+*") {
					isMatch = true
					matcherIndex = j
					break
				} else if v.Count > 100 {
					isMatch = true
					matcherIndex = j
					v.Example = example
					matchers[j] = v
					break
				} else {
					setMatcherExpr(&v, URL)
					v.Example = example
					matchers[j] = v
					continue A
				}
			}
		}

		if !isMatch {
			setMatcherExpr(&matcher, URL)
			matchers = append(matchers, matcher)
		} else {
			matchers[matcherIndex].Count++
		}
	}

	sort.Slice(matchers, func(i, j int) bool {
		return matchers[i].Count > matchers[j].Count
	})

	matchers = deepMatchers(matchers, -1)
	matchers = deepMatchers2(matchers)

	return matchers
}

type ErrStatusCode int

func (e ErrStatusCode) Error() string {
	return "错误的状态码! code:" + strconv.Itoa(int(e))
}

var (
	datePattern   = regexp.MustCompile(`\d{4}.\d+.\d+`)
	StandardBirth = "2006-01-02"
)

func FormatTime(s string) string {
	if datePattern.MatchString(s) {
		s = datePattern.FindString(s)
	}
	t, err := dateparse.ParseStrict(s)
	if err != nil {
		return ""
	}
	return t.Format(StandardBirth)
}
