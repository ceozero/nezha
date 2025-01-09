package i18n

import (
	"embed"
	"fmt"
	"sync/atomic"

	"github.com/chai2010/gettext-go"
)

//go:embed translations
var Translations embed.FS

var Languages = map[string]string{
	"zh_CN": "简体中文",
	"zh_TW": "繁體中文",
	"en_US": "English",
	"es_ES": "Español",
	"de_DE": "Deutsch",
}

type Localizer struct {
	intlMap atomic.Pointer[map[string]gettext.Gettexter]
	lang    atomic.Pointer[string]
}

func NewLocalizer(lang, domain, path string, data any) *Localizer {
	intl := gettext.New(domain, path, data)
	intl.SetLanguage(lang)

	intlMap := make(map[string]gettext.Gettexter)
	intlMap[lang] = intl

	l := new(Localizer)
	l.intlMap.Store(&intlMap)
	l.lang.Store(&lang)

	return l
}

func (l *Localizer) SetLanguage(lang string) {
	l.lang.Store(&lang)
}

func (l *Localizer) Exists(lang string) bool {
	m := *l.intlMap.Load()
	_, ok := m[lang]
	return ok
}

func (l *Localizer) AppendIntl(lang, domain, path string, data any) {
	intl := gettext.New(domain, path, data)
	intl.SetLanguage(lang)

	m := *l.intlMap.Load()
	newMap := make(map[string]gettext.Gettexter, len(m)+1)
	for k, v := range m {
		newMap[k] = v
	}
	newMap[lang] = intl

	l.intlMap.Store(&newMap)
}

// Modified from k8s.io/kubectl/pkg/util/i18n

func (l *Localizer) T(orig string) string {
	m := *l.intlMap.Load()
	intl, ok := m[*l.lang.Load()]
	if !ok {
		return orig
	}

	return intl.PGettext("", orig)
}

// N translates a string, possibly substituting arguments into it along
// the way. If len(args) is > 0, args1 is assumed to be the plural value
// and plural translation is used.
func (l *Localizer) N(orig string, args ...int) string {
	m := *l.intlMap.Load()
	intl, ok := m[*l.lang.Load()]
	if !ok {
		return orig
	}

	if len(args) == 0 {
		return intl.PGettext("", orig)
	}
	return fmt.Sprintf(intl.PNGettext("", orig, orig+".plural", args[0]),
		args[0])
}

// ErrorT produces an error with a translated error string.
// Substitution is performed via the `T` function above, following
// the same rules.
func (l *Localizer) ErrorT(defaultValue string, args ...any) error {
	return fmt.Errorf(l.T(defaultValue), args...)
}

func (l *Localizer) Tf(defaultValue string, args ...any) string {
	return fmt.Sprintf(l.T(defaultValue), args...)
}
