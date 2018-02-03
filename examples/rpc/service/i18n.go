package hello

var (
	AmericanEnglish    = "en-US"
	BritishEnglish     = "en-GB"
	EuropeanSpanish    = "es-ES"
	Finnish            = "fi"
	French             = "fr"
	Japanese           = "ja"
	SimplifiedChinese  = "zh-Hans"
	SupportedLanguages = []string{
		AmericanEnglish,
		BritishEnglish,
		EuropeanSpanish,
		Finnish,
		French,
		Japanese,
		SimplifiedChinese,
	}
	greetings = map[string]string{
		AmericanEnglish:   "Hello",
		BritishEnglish:    "Cheerio",
		EuropeanSpanish:   "Hola",
		Finnish:           "Hei",
		French:            "Bonjour",
		Japanese:          "こんにちは",
		SimplifiedChinese: "你好",
	}
)
