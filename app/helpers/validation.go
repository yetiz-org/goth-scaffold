package helpers

import (
	"regexp"
	"unicode/utf8"

	"golang.org/x/text/language"
)

// ValidatePhoneNumber validates phone number format
// Accepts international format with optional country code
// Examples: +886912345678, 0912345678, +1-555-123-4567, (02)1234-5678
func ValidatePhoneNumber(phone string) bool {
	if phone == "" {
		return false
	}

	// Phone number pattern:
	// - Optional leading + for country code
	// - Can contain digits, spaces, hyphens, parentheses
	// - Must have at least 7 digits (minimum valid phone number)
	// - Maximum 20 characters total (including formatting)
	phonePattern := `^[\+]?[(]?[0-9]{1,4}[)]?[-\s\.]?[(]?[0-9]{1,4}[)]?[-\s\.]?[0-9]{1,4}[-\s\.]?[0-9]{0,9}$`

	matched, _ := regexp.MatchString(phonePattern, phone)
	if !matched {
		return false
	}

	// Count actual digits (must be between 7 and 15)
	digitCount := 0
	for _, c := range phone {
		if c >= '0' && c <= '9' {
			digitCount++
		}
	}

	return digitCount >= 7 && digitCount <= 15
}

// ValidateNameLength validates name field length (max 254 runes, DB limit 255 - 1)
func ValidateNameLength(name string) bool {
	return utf8.RuneCountInString(name) <= 254
}

// ValidateCountryCode validates ISO 3166-1 alpha-3 country code
// Accepts uppercase three-letter country codes (e.g., "TWN", "USA", "JPN")
// Returns true if valid
func ValidateCountryCode(code string) bool {
	if len(code) != 3 {
		return false
	}

	// ISO 3166-1 alpha-3 country codes (complete list)
	validCodes := map[string]bool{
		"AND": true, "ARE": true, "AFG": true, "ATG": true, "AIA": true, "ALB": true, "ARM": true, "AGO": true,
		"ATA": true, "ARG": true, "ASM": true, "AUT": true, "AUS": true, "ABW": true, "ALA": true, "AZE": true,
		"BIH": true, "BRB": true, "BGD": true, "BEL": true, "BFA": true, "BGR": true, "BHR": true, "BDI": true,
		"BEN": true, "BLM": true, "BMU": true, "BRN": true, "BOL": true, "BES": true, "BRA": true, "BHS": true,
		"BTN": true, "BVT": true, "BWA": true, "BLR": true, "BLZ": true, "CAN": true, "CCK": true, "COD": true,
		"CAF": true, "COG": true, "CHE": true, "CIV": true, "COK": true, "CHL": true, "CMR": true, "CHN": true,
		"COL": true, "CRI": true, "CUB": true, "CPV": true, "CUW": true, "CXR": true, "CYP": true, "CZE": true,
		"DEU": true, "DJI": true, "DNK": true, "DMA": true, "DOM": true, "DZA": true, "ECU": true, "EST": true,
		"EGY": true, "ESH": true, "ERI": true, "ESP": true, "ETH": true, "FIN": true, "FJI": true, "FLK": true,
		"FSM": true, "FRO": true, "FRA": true, "GAB": true, "GBR": true, "GRD": true, "GEO": true, "GUF": true,
		"GGY": true, "GHA": true, "GIB": true, "GRL": true, "GMB": true, "GIN": true, "GLP": true, "GNQ": true,
		"GRC": true, "SGS": true, "GTM": true, "GUM": true, "GNB": true, "GUY": true, "HKG": true, "HMD": true,
		"HND": true, "HRV": true, "HTI": true, "HUN": true, "IDN": true, "IRL": true, "ISR": true, "IMN": true,
		"IND": true, "IOT": true, "IRQ": true, "IRN": true, "ISL": true, "ITA": true, "JEY": true, "JAM": true,
		"JOR": true, "JPN": true, "KEN": true, "KGZ": true, "KHM": true, "KIR": true, "COM": true, "KNA": true,
		"PRK": true, "KOR": true, "KWT": true, "CYM": true, "KAZ": true, "LAO": true, "LBN": true, "LCA": true,
		"LIE": true, "LKA": true, "LBR": true, "LSO": true, "LTU": true, "LUX": true, "LVA": true, "LBY": true,
		"MAR": true, "MCO": true, "MDA": true, "MNE": true, "MAF": true, "MDG": true, "MHL": true, "MKD": true,
		"MLI": true, "MMR": true, "MNG": true, "MAC": true, "MNP": true, "MTQ": true, "MRT": true, "MSR": true,
		"MLT": true, "MUS": true, "MDV": true, "MWI": true, "MEX": true, "MYS": true, "MOZ": true, "NAM": true,
		"NCL": true, "NER": true, "NFK": true, "NGA": true, "NIC": true, "NLD": true, "NOR": true, "NPL": true,
		"NRU": true, "NIU": true, "NZL": true, "OMN": true, "PAN": true, "PER": true, "PYF": true, "PNG": true,
		"PHL": true, "PAK": true, "POL": true, "SPM": true, "PCN": true, "PRI": true, "PSE": true, "PRT": true,
		"PLW": true, "PRY": true, "QAT": true, "REU": true, "ROU": true, "SRB": true, "RUS": true, "RWA": true,
		"SAU": true, "SLB": true, "SYC": true, "SDN": true, "SWE": true, "SGP": true, "SHN": true, "SVN": true,
		"SJM": true, "SVK": true, "SLE": true, "SMR": true, "SEN": true, "SOM": true, "SUR": true, "SSD": true,
		"STP": true, "SLV": true, "SXM": true, "SYR": true, "SWZ": true, "TCA": true, "TCD": true, "ATF": true,
		"TGO": true, "THA": true, "TJK": true, "TKL": true, "TLS": true, "TKM": true, "TUN": true, "TON": true,
		"TUR": true, "TTO": true, "TUV": true, "TWN": true, "TZA": true, "UKR": true, "UGA": true, "UMI": true,
		"USA": true, "URY": true, "UZB": true, "VAT": true, "VCT": true, "VEN": true, "VGB": true, "VIR": true,
		"VNM": true, "VUT": true, "WLF": true, "WSM": true, "YEM": true, "MYT": true, "ZAF": true, "ZMB": true,
		"ZWE": true,
	}

	return validCodes[code]
}

// countryCodeAlpha3ToAlpha2 maps ISO 3166-1 alpha-3 to alpha-2 country codes
var countryCodeAlpha3ToAlpha2 = map[string]string{
	"AFG": "AF", "ALA": "AX", "ALB": "AL", "DZA": "DZ", "ASM": "AS", "AND": "AD", "AGO": "AO", "AIA": "AI",
	"ATA": "AQ", "ATG": "AG", "ARG": "AR", "ARM": "AM", "ABW": "AW", "AUS": "AU", "AUT": "AT", "AZE": "AZ",
	"BHS": "BS", "BHR": "BH", "BGD": "BD", "BRB": "BB", "BLR": "BY", "BEL": "BE", "BLZ": "BZ", "BEN": "BJ",
	"BMU": "BM", "BTN": "BT", "BOL": "BO", "BES": "BQ", "BIH": "BA", "BWA": "BW", "BVT": "BV", "BRA": "BR",
	"IOT": "IO", "BRN": "BN", "BGR": "BG", "BFA": "BF", "BDI": "BI", "CPV": "CV", "KHM": "KH", "CMR": "CM",
	"CAN": "CA", "CYM": "KY", "CAF": "CF", "TCD": "TD", "CHL": "CL", "CHN": "CN", "CXR": "CX", "CCK": "CC",
	"COL": "CO", "COM": "KM", "COG": "CG", "COD": "CD", "COK": "CK", "CRI": "CR", "CIV": "CI", "HRV": "HR",
	"CUB": "CU", "CUW": "CW", "CYP": "CY", "CZE": "CZ", "DNK": "DK", "DJI": "DJ", "DMA": "DM", "DOM": "DO",
	"ECU": "EC", "EGY": "EG", "SLV": "SV", "GNQ": "GQ", "ERI": "ER", "EST": "EE", "SWZ": "SZ", "ETH": "ET",
	"FLK": "FK", "FRO": "FO", "FJI": "FJ", "FIN": "FI", "FRA": "FR", "GUF": "GF", "PYF": "PF", "ATF": "TF",
	"GAB": "GA", "GMB": "GM", "GEO": "GE", "DEU": "DE", "GHA": "GH", "GIB": "GI", "GRC": "GR", "GRL": "GL",
	"GRD": "GD", "GLP": "GP", "GUM": "GU", "GTM": "GT", "GGY": "GG", "GIN": "GN", "GNB": "GW", "GUY": "GY",
	"HTI": "HT", "HMD": "HM", "VAT": "VA", "HND": "HN", "HKG": "HK", "HUN": "HU", "ISL": "IS", "IND": "IN",
	"IDN": "ID", "IRN": "IR", "IRQ": "IQ", "IRL": "IE", "IMN": "IM", "ISR": "IL", "ITA": "IT", "JAM": "JM",
	"JPN": "JP", "JEY": "JE", "JOR": "JO", "KAZ": "KZ", "KEN": "KE", "KIR": "KI", "PRK": "KP", "KOR": "KR",
	"KWT": "KW", "KGZ": "KG", "LAO": "LA", "LVA": "LV", "LBN": "LB", "LSO": "LS", "LBR": "LR", "LBY": "LY",
	"LIE": "LI", "LTU": "LT", "LUX": "LU", "MAC": "MO", "MDG": "MG", "MWI": "MW", "MYS": "MY", "MDV": "MV",
	"MLI": "ML", "MLT": "MT", "MHL": "MH", "MTQ": "MQ", "MRT": "MR", "MUS": "MU", "MYT": "YT", "MEX": "MX",
	"FSM": "FM", "MDA": "MD", "MCO": "MC", "MNG": "MN", "MNE": "ME", "MSR": "MS", "MAR": "MA", "MOZ": "MZ",
	"MMR": "MM", "NAM": "NA", "NRU": "NR", "NPL": "NP", "NLD": "NL", "NCL": "NC", "NZL": "NZ", "NIC": "NI",
	"NER": "NE", "NGA": "NG", "NIU": "NU", "NFK": "NF", "MKD": "MK", "MNP": "MP", "NOR": "NO", "OMN": "OM",
	"PAK": "PK", "PLW": "PW", "PSE": "PS", "PAN": "PA", "PNG": "PG", "PRY": "PY", "PER": "PE", "PHL": "PH",
	"PCN": "PN", "POL": "PL", "PRT": "PT", "PRI": "PR", "QAT": "QA", "REU": "RE", "ROU": "RO", "RUS": "RU",
	"RWA": "RW", "BLM": "BL", "SHN": "SH", "KNA": "KN", "LCA": "LC", "MAF": "MF", "SPM": "PM", "VCT": "VC",
	"WSM": "WS", "SMR": "SM", "STP": "ST", "SAU": "SA", "SEN": "SN", "SRB": "RS", "SYC": "SC", "SLE": "SL",
	"SGP": "SG", "SXM": "SX", "SVK": "SK", "SVN": "SI", "SLB": "SB", "SOM": "SO", "ZAF": "ZA", "SGS": "GS",
	"SSD": "SS", "ESP": "ES", "LKA": "LK", "SDN": "SD", "SUR": "SR", "SJM": "SJ", "SWE": "SE", "CHE": "CH",
	"SYR": "SY", "TWN": "TW", "TJK": "TJ", "TZA": "TZ", "THA": "TH", "TLS": "TL", "TGO": "TG", "TKL": "TK",
	"TON": "TO", "TTO": "TT", "TUN": "TN", "TUR": "TR", "TKM": "TM", "TCA": "TC", "TUV": "TV", "UGA": "UG",
	"UKR": "UA", "ARE": "AE", "GBR": "GB", "USA": "US", "UMI": "UM", "URY": "UY", "UZB": "UZ", "VUT": "VU",
	"VEN": "VE", "VNM": "VN", "VGB": "VG", "VIR": "VI", "WLF": "WF", "ESH": "EH", "YEM": "YE", "ZMB": "ZM",
	"ZWE": "ZW",
}

// countryCodeAlpha2ToAlpha3 maps ISO 3166-1 alpha-2 to alpha-3 country codes
var countryCodeAlpha2ToAlpha3 = map[string]string{
	"AF": "AFG", "AX": "ALA", "AL": "ALB", "DZ": "DZA", "AS": "ASM", "AD": "AND", "AO": "AGO", "AI": "AIA",
	"AQ": "ATA", "AG": "ATG", "AR": "ARG", "AM": "ARM", "AW": "ABW", "AU": "AUS", "AT": "AUT", "AZ": "AZE",
	"BS": "BHS", "BH": "BHR", "BD": "BGD", "BB": "BRB", "BY": "BLR", "BE": "BEL", "BZ": "BLZ", "BJ": "BEN",
	"BM": "BMU", "BT": "BTN", "BO": "BOL", "BQ": "BES", "BA": "BIH", "BW": "BWA", "BV": "BVT", "BR": "BRA",
	"IO": "IOT", "BN": "BRN", "BG": "BGR", "BF": "BFA", "BI": "BDI", "CV": "CPV", "KH": "KHM", "CM": "CMR",
	"CA": "CAN", "KY": "CYM", "CF": "CAF", "TD": "TCD", "CL": "CHL", "CN": "CHN", "CX": "CXR", "CC": "CCK",
	"CO": "COL", "KM": "COM", "CG": "COG", "CD": "COD", "CK": "COK", "CR": "CRI", "CI": "CIV", "HR": "HRV",
	"CU": "CUB", "CW": "CUW", "CY": "CYP", "CZ": "CZE", "DK": "DNK", "DJ": "DJI", "DM": "DMA", "DO": "DOM",
	"EC": "ECU", "EG": "EGY", "SV": "SLV", "GQ": "GNQ", "ER": "ERI", "EE": "EST", "SZ": "SWZ", "ET": "ETH",
	"FK": "FLK", "FO": "FRO", "FJ": "FJI", "FI": "FIN", "FR": "FRA", "GF": "GUF", "PF": "PYF", "TF": "ATF",
	"GA": "GAB", "GM": "GMB", "GE": "GEO", "DE": "DEU", "GH": "GHA", "GI": "GIB", "GR": "GRC", "GL": "GRL",
	"GD": "GRD", "GP": "GLP", "GU": "GUM", "GT": "GTM", "GG": "GGY", "GN": "GIN", "GW": "GNB", "GY": "GUY",
	"HT": "HTI", "HM": "HMD", "VA": "VAT", "HN": "HND", "HK": "HKG", "HU": "HUN", "IS": "ISL", "IN": "IND",
	"ID": "IDN", "IR": "IRN", "IQ": "IRQ", "IE": "IRL", "IM": "IMN", "IL": "ISR", "IT": "ITA", "JM": "JAM",
	"JP": "JPN", "JE": "JEY", "JO": "JOR", "KZ": "KAZ", "KE": "KEN", "KI": "KIR", "KP": "PRK", "KR": "KOR",
	"KW": "KWT", "KG": "KGZ", "LA": "LAO", "LV": "LVA", "LB": "LBN", "LS": "LSO", "LR": "LBR", "LY": "LBY",
	"LI": "LIE", "LT": "LTU", "LU": "LUX", "MO": "MAC", "MG": "MDG", "MW": "MWI", "MY": "MYS", "MV": "MDV",
	"ML": "MLI", "MT": "MLT", "MH": "MHL", "MQ": "MTQ", "MR": "MRT", "MU": "MUS", "YT": "MYT", "MX": "MEX",
	"FM": "FSM", "MD": "MDA", "MC": "MCO", "MN": "MNG", "ME": "MNE", "MS": "MSR", "MA": "MAR", "MZ": "MOZ",
	"MM": "MMR", "NA": "NAM", "NR": "NRU", "NP": "NPL", "NL": "NLD", "NC": "NCL", "NZ": "NZL", "NI": "NIC",
	"NE": "NER", "NG": "NGA", "NU": "NIU", "NF": "NFK", "MK": "MKD", "MP": "MNP", "NO": "NOR", "OM": "OMN",
	"PK": "PAK", "PW": "PLW", "PS": "PSE", "PA": "PAN", "PG": "PNG", "PY": "PRY", "PE": "PER", "PH": "PHL",
	"PN": "PCN", "PL": "POL", "PT": "PRT", "PR": "PRI", "QA": "QAT", "RE": "REU", "RO": "ROU", "RU": "RUS",
	"RW": "RWA", "BL": "BLM", "SH": "SHN", "KN": "KNA", "LC": "LCA", "MF": "MAF", "PM": "SPM", "VC": "VCT",
	"WS": "WSM", "SM": "SMR", "ST": "STP", "SA": "SAU", "SN": "SEN", "RS": "SRB", "SC": "SYC", "SL": "SLE",
	"SG": "SGP", "SX": "SXM", "SK": "SVK", "SI": "SVN", "SB": "SLB", "SO": "SOM", "ZA": "ZAF", "GS": "SGS",
	"SS": "SSD", "ES": "ESP", "LK": "LKA", "SD": "SDN", "SR": "SUR", "SJ": "SJM", "SE": "SWE", "CH": "CHE",
	"SY": "SYR", "TW": "TWN", "TJ": "TJK", "TZ": "TZA", "TH": "THA", "TL": "TLS", "TG": "TGO", "TK": "TKL",
	"TO": "TON", "TT": "TTO", "TN": "TUN", "TR": "TUR", "TM": "TKM", "TC": "TCA", "TV": "TUV", "UG": "UGA",
	"UA": "UKR", "AE": "ARE", "GB": "GBR", "US": "USA", "UM": "UMI", "UY": "URY", "UZ": "UZB", "VU": "VUT",
	"VE": "VEN", "VN": "VNM", "VG": "VGB", "VI": "VIR", "WF": "WLF", "EH": "ESH", "YE": "YEM", "ZM": "ZMB",
	"ZW": "ZWE",
}

// CountryCodeAlpha3ToAlpha2 converts ISO 3166-1 alpha-3 to alpha-2 country code
// Returns empty string if code is invalid
func CountryCodeAlpha3ToAlpha2(alpha3 string) string {
	if len(alpha3) != 3 {
		return ""
	}
	return countryCodeAlpha3ToAlpha2[alpha3]
}

// CountryCodeAlpha2ToAlpha3 converts ISO 3166-1 alpha-2 to alpha-3 country code
// Returns empty string if code is invalid
func CountryCodeAlpha2ToAlpha3(alpha2 string) string {
	if len(alpha2) != 2 {
		return ""
	}
	return countryCodeAlpha2ToAlpha3[alpha2]
}

// ValidateLocale validates locale string using BCP 47 language tag format
// Accepts formats: language (e.g., "en", "zh"), language-region (e.g., "zh-TW", "en-US")
// Returns normalized locale string and validation result
// Examples of valid locales: "en", "zh-TW", "ja-JP", "zh-Hant-TW"
func ValidateLocale(locale string) (normalizedLocale string, valid bool) {
	// Empty locale is invalid
	if locale == "" {
		return "", false
	}

	// Check max length (database VARCHAR(85) constraint)
	if len(locale) > 85 {
		return "", false
	}

	// Parse using BCP 47 standard (golang.org/x/text/language)
	tag, err := language.Parse(locale)
	if err != nil {
		return "", false
	}

	// Get normalized form (handles case normalization and canonicalization)
	normalized := tag.String()

	// Verify normalized form doesn't exceed database limit
	if len(normalized) > 85 {
		return "", false
	}

	return normalized, true
}
