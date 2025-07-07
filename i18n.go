package redgiant

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/pmeier/redgiant/internal/errors"
	"github.com/pmeier/redgiant/internal/utils"
)

type Language uint8

const (
	NoLanguage Language = iota
	ChineseLanguage
	EnglishLanguage
	GermanLanguage
	DutchLanguage
	PolishLanguage
)

func (l Language) String() string {
	switch l {
	case NoLanguage:
		return ""
	case ChineseLanguage:
		return "ch_CN"
	case EnglishLanguage:
		return "en_US"
	case GermanLanguage:
		return "de_DE"
	case DutchLanguage:
		return "nl_NL"
	case PolishLanguage:
		return "pl_PL"
	}
	return strconv.Itoa(int(l))
}

func ParseLanguage(langStr string) (Language, error) {
	for _, lang := range []Language{
		NoLanguage,
		ChineseLanguage,
		EnglishLanguage,
		GermanLanguage,
		DutchLanguage,
		PolishLanguage,
	} {
		if strings.EqualFold(langStr, lang.String()) {
			return lang, nil
		}
	}
	return NoLanguage, errors.New(
		"unknown language",
		errors.WithContext(errors.Context{"language": langStr}),
		errors.WithHTTPCode(http.StatusUnprocessableEntity),
		errors.WithHTTPDetail(errors.ContextHTTPDetail),
	)
}

func (l *Language) UnmarshalParam(param string) error {
	lang, err := ParseLanguage(param)
	if err != nil {
		return err
	}
	*l = lang
	return nil
}

type Localizer interface {
	Localize(i18nCode string, lang Language) (string, error)
}

type SungrowLocalizer struct {
	host string
	lm   map[Language]map[string]string
	re   *regexp.Regexp
}

func NewSungrowLocalizer(host string) *SungrowLocalizer {
	return &SungrowLocalizer{
		host: host,
		lm:   map[Language]map[string]string{},
		re:   regexp.MustCompile(`{\d+}`),
	}
}

func (l *SungrowLocalizer) getCodeMap(lang Language) (map[string]string, error) {
	cm, ok := l.lm[lang]
	if ok {
		return cm, nil
	}

	c := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	u := fmt.Sprintf("https://%s/i18n/%s.properties", l.host, lang)
	r, err := c.Get(u)
	if err != nil {
		return nil, errors.Wrap(err, errors.WithContext(errors.Context{"url": u}))
	}
	defer r.Body.Close()

	s := bufio.NewScanner(r.Body)
	cm = map[string]string{}
	for s.Scan() {
		l := s.Text()
		parts := strings.SplitN(l, "=", 2)
		if len(parts) != 2 {
			return nil, errors.New("unknown line format", errors.WithContext(errors.Context{"line": l}))
		}
		i18nCode, name := parts[0], parts[1]
		cm[i18nCode] = name
	}
	l.lm[lang] = cm
	return cm, nil
}

func (l *SungrowLocalizer) Localize(i18nCode string, lang Language) (string, error) {
	if lang == NoLanguage {
		return i18nCode, nil
	}

	cm, err := l.getCodeMap(lang)
	if err != nil {
		return "", err
	}

	// TODO: this likely needs to be adapted
	parts := strings.Split(i18nCode, "%@")
	i18nCode, args := parts[0], parts[1:]

	v, ok := cm[i18nCode]
	if !ok {
		return "", errors.New(
			"unknown i18n code",
			errors.WithContext(errors.Context{"i18nCode": i18nCode, "language": lang.String()}),
			errors.WithHTTPCode(http.StatusUnprocessableEntity),
			errors.WithHTTPDetail(errors.ContextHTTPDetail),
		)
	}

	if len(args) > 0 {
		v = l.re.ReplaceAllStringFunc(v, func(a string) string {
			// FIXME use proper group replace instead of indexing
			return args[utils.Must(strconv.Atoi(a[1:len(a)-1]))]
		})
	}

	return v, nil
}
