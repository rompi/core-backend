package format

import (
	"strings"
	"time"
)

// DateStyle defines the style for date formatting.
type DateStyle int

const (
	// DateStyleShort formats as 1/15/24 (locale dependent).
	DateStyleShort DateStyle = iota
	// DateStyleMedium formats as Jan 15, 2024.
	DateStyleMedium
	// DateStyleLong formats as January 15, 2024.
	DateStyleLong
	// DateStyleFull formats as Monday, January 15, 2024.
	DateStyleFull
)

// TimeStyle defines the style for time formatting.
type TimeStyle int

const (
	// TimeStyleShort formats as 3:04 PM (locale dependent).
	TimeStyleShort TimeStyle = iota
	// TimeStyleMedium formats as 3:04:05 PM.
	TimeStyleMedium
	// TimeStyleLong formats as 3:04:05 PM EST.
	TimeStyleLong
)

// DateTimeFormat holds locale-specific date/time formatting patterns.
type DateTimeFormat struct {
	// Date formats
	DateShort  string
	DateMedium string
	DateLong   string
	DateFull   string

	// Time formats
	TimeShort  string
	TimeMedium string
	TimeLong   string

	// DateTime separator
	DateTimeSep string

	// Month names
	MonthsShort []string
	MonthsLong  []string

	// Weekday names
	WeekdaysShort []string
	WeekdaysLong  []string

	// AM/PM
	AM string
	PM string

	// Use 24-hour clock
	Use24Hour bool
}

// localeDateTimeFormats contains date/time formatting rules for various locales.
var localeDateTimeFormats = map[string]DateTimeFormat{
	"en": {
		DateShort:     "1/2/06",
		DateMedium:    "Jan 2, 2006",
		DateLong:      "January 2, 2006",
		DateFull:      "Monday, January 2, 2006",
		TimeShort:     "3:04 PM",
		TimeMedium:    "3:04:05 PM",
		TimeLong:      "3:04:05 PM MST",
		DateTimeSep:   ", ",
		MonthsShort:   []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"},
		MonthsLong:    []string{"January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"},
		WeekdaysShort: []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"},
		WeekdaysLong:  []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"},
		AM:            "AM",
		PM:            "PM",
		Use24Hour:     false,
	},
	"de": {
		DateShort:     "02.01.06",
		DateMedium:    "2. Jan. 2006",
		DateLong:      "2. January 2006",
		DateFull:      "Monday, 2. January 2006",
		TimeShort:     "15:04",
		TimeMedium:    "15:04:05",
		TimeLong:      "15:04:05 MST",
		DateTimeSep:   ", ",
		MonthsShort:   []string{"Jan", "Feb", "Mär", "Apr", "Mai", "Jun", "Jul", "Aug", "Sep", "Okt", "Nov", "Dez"},
		MonthsLong:    []string{"Januar", "Februar", "März", "April", "Mai", "Juni", "Juli", "August", "September", "Oktober", "November", "Dezember"},
		WeekdaysShort: []string{"So", "Mo", "Di", "Mi", "Do", "Fr", "Sa"},
		WeekdaysLong:  []string{"Sonntag", "Montag", "Dienstag", "Mittwoch", "Donnerstag", "Freitag", "Samstag"},
		Use24Hour:     true,
	},
	"fr": {
		DateShort:     "02/01/06",
		DateMedium:    "2 janv. 2006",
		DateLong:      "2 January 2006",
		DateFull:      "Monday 2 January 2006",
		TimeShort:     "15:04",
		TimeMedium:    "15:04:05",
		TimeLong:      "15:04:05 MST",
		DateTimeSep:   " à ",
		MonthsShort:   []string{"janv.", "févr.", "mars", "avr.", "mai", "juin", "juil.", "août", "sept.", "oct.", "nov.", "déc."},
		MonthsLong:    []string{"janvier", "février", "mars", "avril", "mai", "juin", "juillet", "août", "septembre", "octobre", "novembre", "décembre"},
		WeekdaysShort: []string{"dim.", "lun.", "mar.", "mer.", "jeu.", "ven.", "sam."},
		WeekdaysLong:  []string{"dimanche", "lundi", "mardi", "mercredi", "jeudi", "vendredi", "samedi"},
		Use24Hour:     true,
	},
	"es": {
		DateShort:     "2/1/06",
		DateMedium:    "2 ene 2006",
		DateLong:      "2 de January de 2006",
		DateFull:      "Monday, 2 de January de 2006",
		TimeShort:     "15:04",
		TimeMedium:    "15:04:05",
		TimeLong:      "15:04:05 MST",
		DateTimeSep:   ", ",
		MonthsShort:   []string{"ene", "feb", "mar", "abr", "may", "jun", "jul", "ago", "sep", "oct", "nov", "dic"},
		MonthsLong:    []string{"enero", "febrero", "marzo", "abril", "mayo", "junio", "julio", "agosto", "septiembre", "octubre", "noviembre", "diciembre"},
		WeekdaysShort: []string{"dom.", "lun.", "mar.", "mié.", "jue.", "vie.", "sáb."},
		WeekdaysLong:  []string{"domingo", "lunes", "martes", "miércoles", "jueves", "viernes", "sábado"},
		Use24Hour:     true,
	},
	"ja": {
		DateShort:     "06/01/02",
		DateMedium:    "2006年1月2日",
		DateLong:      "2006年1月2日",
		DateFull:      "2006年1月2日Monday",
		TimeShort:     "15:04",
		TimeMedium:    "15:04:05",
		TimeLong:      "15:04:05 MST",
		DateTimeSep:   " ",
		MonthsShort:   []string{"1月", "2月", "3月", "4月", "5月", "6月", "7月", "8月", "9月", "10月", "11月", "12月"},
		MonthsLong:    []string{"1月", "2月", "3月", "4月", "5月", "6月", "7月", "8月", "9月", "10月", "11月", "12月"},
		WeekdaysShort: []string{"日", "月", "火", "水", "木", "金", "土"},
		WeekdaysLong:  []string{"日曜日", "月曜日", "火曜日", "水曜日", "木曜日", "金曜日", "土曜日"},
		Use24Hour:     true,
	},
	"zh": {
		DateShort:     "06/1/2",
		DateMedium:    "2006年1月2日",
		DateLong:      "2006年1月2日",
		DateFull:      "2006年1月2日 Monday",
		TimeShort:     "15:04",
		TimeMedium:    "15:04:05",
		TimeLong:      "15:04:05 MST",
		DateTimeSep:   " ",
		MonthsShort:   []string{"1月", "2月", "3月", "4月", "5月", "6月", "7月", "8月", "9月", "10月", "11月", "12月"},
		MonthsLong:    []string{"一月", "二月", "三月", "四月", "五月", "六月", "七月", "八月", "九月", "十月", "十一月", "十二月"},
		WeekdaysShort: []string{"周日", "周一", "周二", "周三", "周四", "周五", "周六"},
		WeekdaysLong:  []string{"星期日", "星期一", "星期二", "星期三", "星期四", "星期五", "星期六"},
		Use24Hour:     true,
	},
	"ko": {
		DateShort:     "06. 1. 2.",
		DateMedium:    "2006년 1월 2일",
		DateLong:      "2006년 1월 2일",
		DateFull:      "2006년 1월 2일 Monday",
		TimeShort:     "3:04 PM",
		TimeMedium:    "3:04:05 PM",
		TimeLong:      "3:04:05 PM MST",
		DateTimeSep:   " ",
		MonthsShort:   []string{"1월", "2월", "3월", "4월", "5월", "6월", "7월", "8월", "9월", "10월", "11월", "12월"},
		MonthsLong:    []string{"1월", "2월", "3월", "4월", "5월", "6월", "7월", "8월", "9월", "10월", "11월", "12월"},
		WeekdaysShort: []string{"일", "월", "화", "수", "목", "금", "토"},
		WeekdaysLong:  []string{"일요일", "월요일", "화요일", "수요일", "목요일", "금요일", "토요일"},
		AM:            "오전",
		PM:            "오후",
		Use24Hour:     false,
	},
	"ru": {
		DateShort:     "02.01.06",
		DateMedium:    "2 янв. 2006 г.",
		DateLong:      "2 January 2006 г.",
		DateFull:      "Monday, 2 January 2006 г.",
		TimeShort:     "15:04",
		TimeMedium:    "15:04:05",
		TimeLong:      "15:04:05 MST",
		DateTimeSep:   ", ",
		MonthsShort:   []string{"янв.", "февр.", "мар.", "апр.", "мая", "июн.", "июл.", "авг.", "сент.", "окт.", "нояб.", "дек."},
		MonthsLong:    []string{"января", "февраля", "марта", "апреля", "мая", "июня", "июля", "августа", "сентября", "октября", "ноября", "декабря"},
		WeekdaysShort: []string{"вс", "пн", "вт", "ср", "чт", "пт", "сб"},
		WeekdaysLong:  []string{"воскресенье", "понедельник", "вторник", "среда", "четверг", "пятница", "суббота"},
		Use24Hour:     true,
	},
	"ar": {
		DateShort:     "2/1/06",
		DateMedium:    "2 يناير 2006",
		DateLong:      "2 January 2006",
		DateFull:      "Monday، 2 January 2006",
		TimeShort:     "3:04 م",
		TimeMedium:    "3:04:05 م",
		TimeLong:      "3:04:05 م MST",
		DateTimeSep:   "، ",
		MonthsShort:   []string{"يناير", "فبراير", "مارس", "أبريل", "مايو", "يونيو", "يوليو", "أغسطس", "سبتمبر", "أكتوبر", "نوفمبر", "ديسمبر"},
		MonthsLong:    []string{"يناير", "فبراير", "مارس", "أبريل", "مايو", "يونيو", "يوليو", "أغسطس", "سبتمبر", "أكتوبر", "نوفمبر", "ديسمبر"},
		WeekdaysShort: []string{"أحد", "إثنين", "ثلاثاء", "أربعاء", "خميس", "جمعة", "سبت"},
		WeekdaysLong:  []string{"الأحد", "الإثنين", "الثلاثاء", "الأربعاء", "الخميس", "الجمعة", "السبت"},
		AM:            "ص",
		PM:            "م",
		Use24Hour:     false,
	},
	"pt": {
		DateShort:     "02/01/06",
		DateMedium:    "2 de jan. de 2006",
		DateLong:      "2 de January de 2006",
		DateFull:      "Monday, 2 de January de 2006",
		TimeShort:     "15:04",
		TimeMedium:    "15:04:05",
		TimeLong:      "15:04:05 MST",
		DateTimeSep:   " ",
		MonthsShort:   []string{"jan.", "fev.", "mar.", "abr.", "mai.", "jun.", "jul.", "ago.", "set.", "out.", "nov.", "dez."},
		MonthsLong:    []string{"janeiro", "fevereiro", "março", "abril", "maio", "junho", "julho", "agosto", "setembro", "outubro", "novembro", "dezembro"},
		WeekdaysShort: []string{"dom.", "seg.", "ter.", "qua.", "qui.", "sex.", "sáb."},
		WeekdaysLong:  []string{"domingo", "segunda-feira", "terça-feira", "quarta-feira", "quinta-feira", "sexta-feira", "sábado"},
		Use24Hour:     true,
	},
}

// GetDateTimeFormat returns the date/time format for a locale.
func GetDateTimeFormat(locale string) DateTimeFormat {
	// Try exact match
	if fmt, ok := localeDateTimeFormats[locale]; ok {
		return fmt
	}

	// Try language only
	if idx := strings.Index(locale, "-"); idx != -1 {
		lang := locale[:idx]
		if fmt, ok := localeDateTimeFormats[lang]; ok {
			return fmt
		}
	}

	// Default to English
	return localeDateTimeFormats["en"]
}

// FormatDate formats a date according to locale conventions.
func FormatDate(locale string, t time.Time, style DateStyle) string {
	dtf := GetDateTimeFormat(locale)

	var pattern string
	switch style {
	case DateStyleShort:
		pattern = dtf.DateShort
	case DateStyleMedium:
		pattern = dtf.DateMedium
	case DateStyleLong:
		pattern = dtf.DateLong
	case DateStyleFull:
		pattern = dtf.DateFull
	default:
		pattern = dtf.DateMedium
	}

	return formatDateTime(t, pattern, dtf)
}

// FormatTime formats a time according to locale conventions.
func FormatTime(locale string, t time.Time, style TimeStyle) string {
	dtf := GetDateTimeFormat(locale)

	var pattern string
	switch style {
	case TimeStyleShort:
		pattern = dtf.TimeShort
	case TimeStyleMedium:
		pattern = dtf.TimeMedium
	case TimeStyleLong:
		pattern = dtf.TimeLong
	default:
		pattern = dtf.TimeShort
	}

	return formatDateTime(t, pattern, dtf)
}

// FormatDateTime formats a date and time according to locale conventions.
func FormatDateTime(locale string, t time.Time, dateStyle DateStyle, timeStyle TimeStyle) string {
	date := FormatDate(locale, t, dateStyle)
	timeStr := FormatTime(locale, t, timeStyle)
	dtf := GetDateTimeFormat(locale)

	return date + dtf.DateTimeSep + timeStr
}

// formatDateTime formats a time using the given pattern and locale format.
func formatDateTime(t time.Time, pattern string, dtf DateTimeFormat) string {
	// Replace month names
	result := t.Format(pattern)

	// Replace English month names with locale-specific ones
	for i, enMonth := range []string{"January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"} {
		if len(dtf.MonthsLong) > i {
			result = strings.ReplaceAll(result, enMonth, dtf.MonthsLong[i])
		}
	}
	for i, enMonth := range []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"} {
		if len(dtf.MonthsShort) > i {
			result = strings.ReplaceAll(result, enMonth, dtf.MonthsShort[i])
		}
	}

	// Replace weekday names
	for i, enDay := range []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"} {
		if len(dtf.WeekdaysLong) > i {
			result = strings.ReplaceAll(result, enDay, dtf.WeekdaysLong[i])
		}
	}
	for i, enDay := range []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"} {
		if len(dtf.WeekdaysShort) > i {
			result = strings.ReplaceAll(result, enDay, dtf.WeekdaysShort[i])
		}
	}

	// Replace AM/PM if defined
	if dtf.AM != "" {
		result = strings.ReplaceAll(result, "AM", dtf.AM)
	}
	if dtf.PM != "" {
		result = strings.ReplaceAll(result, "PM", dtf.PM)
	}

	return result
}
