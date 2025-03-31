package redgiant

import (
	"bufio"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type Language uint8

const (
	Chinese Language = iota
	English
	German
	Dutch
	Polish
)

func (l Language) String() string {
	switch l {
	case Chinese:
		return "ch_CN"
	case English:
		return "en_US"
	case German:
		return "de_DE"
	case Dutch:
		return "nl_NL"
	case Polish:
		return "pl_PL"
	}
	return strconv.Itoa(int(l))
}

type Localizer interface {
	Localize(i18nCode string, lang Language) (string, error)
}

type SungrowLocalizer struct {
	host string
	lm   map[Language]map[string]string
}

func NewSungrowLocalizer(host string) *SungrowLocalizer {
	return &SungrowLocalizer{host: host, lm: map[Language]map[string]string{}}
}

func (l *SungrowLocalizer) getCodeMap(lang Language) (map[string]string, error) {
	cm, ok := l.lm[lang]
	if ok {
		return cm, nil
	}

	r, err := http.Get(fmt.Sprintf("https://%s/i18n/%s.properties", l.host, lang))
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	s := bufio.NewScanner(r.Body)
	cm = map[string]string{}
	for s.Scan() {
		parts := strings.Split(s.Text(), "=")
		if len(parts) != 2 {
			return nil, errors.New("unknown line format")
		}
		i18nCode, name := parts[0], parts[1]
		cm[i18nCode] = name
	}
	return cm, nil
}

func (l *SungrowLocalizer) Localize(i18nCode string, lang Language) (string, error) {
	cm, err := l.getCodeMap(lang)
	if err != nil {
		return "", err
	}

	name, ok := cm[i18nCode]
	if !ok {
		return "", errors.New("unknown i18n code")
	}

	return name, nil
}
