package format

import (
	"strings"
	"time"
)

// RelativeTimeFormat holds locale-specific relative time phrases.
type RelativeTimeFormat struct {
	JustNow    string
	SecondsAgo string
	MinuteAgo  string
	MinutesAgo string
	HourAgo    string
	HoursAgo   string
	Yesterday  string
	DaysAgo    string
	WeekAgo    string
	WeeksAgo   string
	MonthAgo   string
	MonthsAgo  string
	YearAgo    string
	YearsAgo   string

	InSeconds string
	InMinute  string
	InMinutes string
	InHour    string
	InHours   string
	Tomorrow  string
	InDays    string
	InWeek    string
	InWeeks   string
	InMonth   string
	InMonths  string
	InYear    string
	InYears   string
}

// localeRelativeTimeFormats contains relative time phrases for various locales.
var localeRelativeTimeFormats = map[string]RelativeTimeFormat{
	"en": {
		JustNow:    "just now",
		SecondsAgo: "%d seconds ago",
		MinuteAgo:  "1 minute ago",
		MinutesAgo: "%d minutes ago",
		HourAgo:    "1 hour ago",
		HoursAgo:   "%d hours ago",
		Yesterday:  "yesterday",
		DaysAgo:    "%d days ago",
		WeekAgo:    "last week",
		WeeksAgo:   "%d weeks ago",
		MonthAgo:   "last month",
		MonthsAgo:  "%d months ago",
		YearAgo:    "last year",
		YearsAgo:   "%d years ago",

		InSeconds: "in %d seconds",
		InMinute:  "in 1 minute",
		InMinutes: "in %d minutes",
		InHour:    "in 1 hour",
		InHours:   "in %d hours",
		Tomorrow:  "tomorrow",
		InDays:    "in %d days",
		InWeek:    "next week",
		InWeeks:   "in %d weeks",
		InMonth:   "next month",
		InMonths:  "in %d months",
		InYear:    "next year",
		InYears:   "in %d years",
	},
	"es": {
		JustNow:    "ahora mismo",
		SecondsAgo: "hace %d segundos",
		MinuteAgo:  "hace 1 minuto",
		MinutesAgo: "hace %d minutos",
		HourAgo:    "hace 1 hora",
		HoursAgo:   "hace %d horas",
		Yesterday:  "ayer",
		DaysAgo:    "hace %d días",
		WeekAgo:    "la semana pasada",
		WeeksAgo:   "hace %d semanas",
		MonthAgo:   "el mes pasado",
		MonthsAgo:  "hace %d meses",
		YearAgo:    "el año pasado",
		YearsAgo:   "hace %d años",

		InSeconds: "en %d segundos",
		InMinute:  "en 1 minuto",
		InMinutes: "en %d minutos",
		InHour:    "en 1 hora",
		InHours:   "en %d horas",
		Tomorrow:  "mañana",
		InDays:    "en %d días",
		InWeek:    "la próxima semana",
		InWeeks:   "en %d semanas",
		InMonth:   "el próximo mes",
		InMonths:  "en %d meses",
		InYear:    "el próximo año",
		InYears:   "en %d años",
	},
	"de": {
		JustNow:    "gerade eben",
		SecondsAgo: "vor %d Sekunden",
		MinuteAgo:  "vor 1 Minute",
		MinutesAgo: "vor %d Minuten",
		HourAgo:    "vor 1 Stunde",
		HoursAgo:   "vor %d Stunden",
		Yesterday:  "gestern",
		DaysAgo:    "vor %d Tagen",
		WeekAgo:    "letzte Woche",
		WeeksAgo:   "vor %d Wochen",
		MonthAgo:   "letzten Monat",
		MonthsAgo:  "vor %d Monaten",
		YearAgo:    "letztes Jahr",
		YearsAgo:   "vor %d Jahren",

		InSeconds: "in %d Sekunden",
		InMinute:  "in 1 Minute",
		InMinutes: "in %d Minuten",
		InHour:    "in 1 Stunde",
		InHours:   "in %d Stunden",
		Tomorrow:  "morgen",
		InDays:    "in %d Tagen",
		InWeek:    "nächste Woche",
		InWeeks:   "in %d Wochen",
		InMonth:   "nächsten Monat",
		InMonths:  "in %d Monaten",
		InYear:    "nächstes Jahr",
		InYears:   "in %d Jahren",
	},
	"fr": {
		JustNow:    "à l'instant",
		SecondsAgo: "il y a %d secondes",
		MinuteAgo:  "il y a 1 minute",
		MinutesAgo: "il y a %d minutes",
		HourAgo:    "il y a 1 heure",
		HoursAgo:   "il y a %d heures",
		Yesterday:  "hier",
		DaysAgo:    "il y a %d jours",
		WeekAgo:    "la semaine dernière",
		WeeksAgo:   "il y a %d semaines",
		MonthAgo:   "le mois dernier",
		MonthsAgo:  "il y a %d mois",
		YearAgo:    "l'année dernière",
		YearsAgo:   "il y a %d ans",

		InSeconds: "dans %d secondes",
		InMinute:  "dans 1 minute",
		InMinutes: "dans %d minutes",
		InHour:    "dans 1 heure",
		InHours:   "dans %d heures",
		Tomorrow:  "demain",
		InDays:    "dans %d jours",
		InWeek:    "la semaine prochaine",
		InWeeks:   "dans %d semaines",
		InMonth:   "le mois prochain",
		InMonths:  "dans %d mois",
		InYear:    "l'année prochaine",
		InYears:   "dans %d ans",
	},
	"ja": {
		JustNow:    "たった今",
		SecondsAgo: "%d秒前",
		MinuteAgo:  "1分前",
		MinutesAgo: "%d分前",
		HourAgo:    "1時間前",
		HoursAgo:   "%d時間前",
		Yesterday:  "昨日",
		DaysAgo:    "%d日前",
		WeekAgo:    "先週",
		WeeksAgo:   "%d週間前",
		MonthAgo:   "先月",
		MonthsAgo:  "%dヶ月前",
		YearAgo:    "去年",
		YearsAgo:   "%d年前",

		InSeconds: "%d秒後",
		InMinute:  "1分後",
		InMinutes: "%d分後",
		InHour:    "1時間後",
		InHours:   "%d時間後",
		Tomorrow:  "明日",
		InDays:    "%d日後",
		InWeek:    "来週",
		InWeeks:   "%d週間後",
		InMonth:   "来月",
		InMonths:  "%dヶ月後",
		InYear:    "来年",
		InYears:   "%d年後",
	},
	"zh": {
		JustNow:    "刚刚",
		SecondsAgo: "%d秒前",
		MinuteAgo:  "1分钟前",
		MinutesAgo: "%d分钟前",
		HourAgo:    "1小时前",
		HoursAgo:   "%d小时前",
		Yesterday:  "昨天",
		DaysAgo:    "%d天前",
		WeekAgo:    "上周",
		WeeksAgo:   "%d周前",
		MonthAgo:   "上个月",
		MonthsAgo:  "%d个月前",
		YearAgo:    "去年",
		YearsAgo:   "%d年前",

		InSeconds: "%d秒后",
		InMinute:  "1分钟后",
		InMinutes: "%d分钟后",
		InHour:    "1小时后",
		InHours:   "%d小时后",
		Tomorrow:  "明天",
		InDays:    "%d天后",
		InWeek:    "下周",
		InWeeks:   "%d周后",
		InMonth:   "下个月",
		InMonths:  "%d个月后",
		InYear:    "明年",
		InYears:   "%d年后",
	},
	"ru": {
		JustNow:    "только что",
		SecondsAgo: "%d секунд назад",
		MinuteAgo:  "1 минуту назад",
		MinutesAgo: "%d минут назад",
		HourAgo:    "1 час назад",
		HoursAgo:   "%d часов назад",
		Yesterday:  "вчера",
		DaysAgo:    "%d дней назад",
		WeekAgo:    "на прошлой неделе",
		WeeksAgo:   "%d недель назад",
		MonthAgo:   "в прошлом месяце",
		MonthsAgo:  "%d месяцев назад",
		YearAgo:    "в прошлом году",
		YearsAgo:   "%d лет назад",

		InSeconds: "через %d секунд",
		InMinute:  "через 1 минуту",
		InMinutes: "через %d минут",
		InHour:    "через 1 час",
		InHours:   "через %d часов",
		Tomorrow:  "завтра",
		InDays:    "через %d дней",
		InWeek:    "на следующей неделе",
		InWeeks:   "через %d недель",
		InMonth:   "в следующем месяце",
		InMonths:  "через %d месяцев",
		InYear:    "в следующем году",
		InYears:   "через %d лет",
	},
	"pt": {
		JustNow:    "agora mesmo",
		SecondsAgo: "há %d segundos",
		MinuteAgo:  "há 1 minuto",
		MinutesAgo: "há %d minutos",
		HourAgo:    "há 1 hora",
		HoursAgo:   "há %d horas",
		Yesterday:  "ontem",
		DaysAgo:    "há %d dias",
		WeekAgo:    "semana passada",
		WeeksAgo:   "há %d semanas",
		MonthAgo:   "mês passado",
		MonthsAgo:  "há %d meses",
		YearAgo:    "ano passado",
		YearsAgo:   "há %d anos",

		InSeconds: "em %d segundos",
		InMinute:  "em 1 minuto",
		InMinutes: "em %d minutos",
		InHour:    "em 1 hora",
		InHours:   "em %d horas",
		Tomorrow:  "amanhã",
		InDays:    "em %d dias",
		InWeek:    "próxima semana",
		InWeeks:   "em %d semanas",
		InMonth:   "próximo mês",
		InMonths:  "em %d meses",
		InYear:    "próximo ano",
		InYears:   "em %d anos",
	},
	"ko": {
		JustNow:    "방금",
		SecondsAgo: "%d초 전",
		MinuteAgo:  "1분 전",
		MinutesAgo: "%d분 전",
		HourAgo:    "1시간 전",
		HoursAgo:   "%d시간 전",
		Yesterday:  "어제",
		DaysAgo:    "%d일 전",
		WeekAgo:    "지난주",
		WeeksAgo:   "%d주 전",
		MonthAgo:   "지난달",
		MonthsAgo:  "%d개월 전",
		YearAgo:    "작년",
		YearsAgo:   "%d년 전",

		InSeconds: "%d초 후",
		InMinute:  "1분 후",
		InMinutes: "%d분 후",
		InHour:    "1시간 후",
		InHours:   "%d시간 후",
		Tomorrow:  "내일",
		InDays:    "%d일 후",
		InWeek:    "다음주",
		InWeeks:   "%d주 후",
		InMonth:   "다음달",
		InMonths:  "%d개월 후",
		InYear:    "내년",
		InYears:   "%d년 후",
	},
	"ar": {
		JustNow:    "الآن",
		SecondsAgo: "منذ %d ثانية",
		MinuteAgo:  "منذ دقيقة",
		MinutesAgo: "منذ %d دقائق",
		HourAgo:    "منذ ساعة",
		HoursAgo:   "منذ %d ساعات",
		Yesterday:  "أمس",
		DaysAgo:    "منذ %d أيام",
		WeekAgo:    "الأسبوع الماضي",
		WeeksAgo:   "منذ %d أسابيع",
		MonthAgo:   "الشهر الماضي",
		MonthsAgo:  "منذ %d أشهر",
		YearAgo:    "العام الماضي",
		YearsAgo:   "منذ %d سنوات",

		InSeconds: "خلال %d ثانية",
		InMinute:  "خلال دقيقة",
		InMinutes: "خلال %d دقائق",
		InHour:    "خلال ساعة",
		InHours:   "خلال %d ساعات",
		Tomorrow:  "غدا",
		InDays:    "خلال %d أيام",
		InWeek:    "الأسبوع القادم",
		InWeeks:   "خلال %d أسابيع",
		InMonth:   "الشهر القادم",
		InMonths:  "خلال %d أشهر",
		InYear:    "العام القادم",
		InYears:   "خلال %d سنوات",
	},
}

// GetRelativeTimeFormat returns the relative time format for a locale.
func GetRelativeTimeFormat(locale string) RelativeTimeFormat {
	// Try exact match
	if fmt, ok := localeRelativeTimeFormats[locale]; ok {
		return fmt
	}

	// Try language only
	if idx := strings.Index(locale, "-"); idx != -1 {
		lang := locale[:idx]
		if fmt, ok := localeRelativeTimeFormats[lang]; ok {
			return fmt
		}
	}

	// Default to English
	return localeRelativeTimeFormats["en"]
}

// FormatRelativeTime formats a time as relative to now (e.g., "2 hours ago").
func FormatRelativeTime(locale string, t time.Time) string {
	return FormatRelativeTimeFrom(locale, t, time.Now())
}

// FormatRelativeTimeFrom formats a time relative to a reference time.
func FormatRelativeTimeFrom(locale string, t, ref time.Time) string {
	rtf := GetRelativeTimeFormat(locale)
	diff := ref.Sub(t)

	// Handle future times
	if diff < 0 {
		return formatFutureTime(rtf, -diff)
	}

	return formatPastTime(rtf, diff)
}

// formatPastTime formats a duration in the past.
func formatPastTime(rtf RelativeTimeFormat, diff time.Duration) string {
	seconds := int(diff.Seconds())
	minutes := int(diff.Minutes())
	hours := int(diff.Hours())
	days := hours / 24
	weeks := days / 7
	months := days / 30
	years := days / 365

	switch {
	case seconds < 10:
		return rtf.JustNow
	case seconds < 60:
		return formatWithNumber(rtf.SecondsAgo, seconds)
	case minutes == 1:
		return rtf.MinuteAgo
	case minutes < 60:
		return formatWithNumber(rtf.MinutesAgo, minutes)
	case hours == 1:
		return rtf.HourAgo
	case hours < 24:
		return formatWithNumber(rtf.HoursAgo, hours)
	case days == 1:
		return rtf.Yesterday
	case days < 7:
		return formatWithNumber(rtf.DaysAgo, days)
	case weeks == 1:
		return rtf.WeekAgo
	case weeks < 4:
		return formatWithNumber(rtf.WeeksAgo, weeks)
	case months == 1:
		return rtf.MonthAgo
	case months < 12:
		return formatWithNumber(rtf.MonthsAgo, months)
	case years == 1:
		return rtf.YearAgo
	default:
		return formatWithNumber(rtf.YearsAgo, years)
	}
}

// formatFutureTime formats a duration in the future.
func formatFutureTime(rtf RelativeTimeFormat, diff time.Duration) string {
	seconds := int(diff.Seconds())
	minutes := int(diff.Minutes())
	hours := int(diff.Hours())
	days := hours / 24
	weeks := days / 7
	months := days / 30
	years := days / 365

	switch {
	case seconds < 10:
		return rtf.JustNow
	case seconds < 60:
		return formatWithNumber(rtf.InSeconds, seconds)
	case minutes == 1:
		return rtf.InMinute
	case minutes < 60:
		return formatWithNumber(rtf.InMinutes, minutes)
	case hours == 1:
		return rtf.InHour
	case hours < 24:
		return formatWithNumber(rtf.InHours, hours)
	case days == 1:
		return rtf.Tomorrow
	case days < 7:
		return formatWithNumber(rtf.InDays, days)
	case weeks == 1:
		return rtf.InWeek
	case weeks < 4:
		return formatWithNumber(rtf.InWeeks, weeks)
	case months == 1:
		return rtf.InMonth
	case months < 12:
		return formatWithNumber(rtf.InMonths, months)
	case years == 1:
		return rtf.InYear
	default:
		return formatWithNumber(rtf.InYears, years)
	}
}

// formatWithNumber replaces %d with the given number in the format string.
func formatWithNumber(format string, n int) string {
	return strings.Replace(format, "%d", itoa(n), 1)
}

// itoa converts an integer to a string without importing strconv.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}

	negative := n < 0
	if negative {
		n = -n
	}

	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}

	if negative {
		digits = append([]byte{'-'}, digits...)
	}

	return string(digits)
}
