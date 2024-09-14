package re2go

/*!rules:re2c:base_template
re2c:eof              = 0;
re2c:yyfill:enable    = 0;
re2c:posix-captures   = 0;
re2c:case-insensitive = 0;

re2c:define:YYCTYPE     = byte;
re2c:define:YYPEEK      = "input[cursor]";
re2c:define:YYSKIP      = "cursor++";
re2c:define:YYBACKUP    = "marker = cursor";
re2c:define:YYRESTORE   = "cursor = marker";
re2c:define:YYLESSTHAN  = "limit <= cursor";
re2c:define:YYSTAGP     = "@@{tag} = cursor";
re2c:define:YYSTAGN     = "@@{tag} = -1";
re2c:define:YYSHIFTSTAG = "@@{tag} += @@{shift}";
*/

// TODO: check "août"
// PYTHON NAME: LONG_TEXT_PATTERN
// Given the following pattern:
//
// - day: [0-3]?[0-9]
// - year: 199[0-9]|20[0-3][0-9]
// - month: January?|February?|March|A[pv]ril|Ma[iy]|Jun[ei]|Jul[iy]|August|September|O[ck]tober|November|De[csz]ember|Jan|Feb|M[aä]r|Apr|Jun|Jul|Aug|Sep|O[ck]t|Nov|De[cz]|Januari|Februari|Maret|Mei|Agustus|Jänner|Feber|März|janvier|février|mars|juin|juillet|aout|septembre|octobre|novembre|décembre|Ocak|Şubat|Mart|Nisan|Mayıs|Haziran|Temmuz|Ağustos|Eylül|Ekim|Kasım|Aralık|Oca|Şub|Mar|Nis|Haz|Tem|Ağu|Eyl|Eki|Kas|Ara
//
// It's combined into two patterns:
//
// - MDY = ({rxMonth})[\t\n\f\r ]({rxDay})(!st|nd|rd|th)?,?[\t\n\f\r ]({rxYear})
// - DMY = ({rxDay})(!st|nd|rd|th|[.])?[\t\n\f\r ](!of[\t\n\f\r ])?({rxMonth})[,.]?[\t\n\f\r ]({rxYear})
func FindLongTextPattern(input string) (year, month, day string, ok bool) {
	var cursor, marker int
	input += string(rune(0)) // add terminating null
	limit := len(input) - 1  // limit points at the terminating null

	// Capturing groups
	/*!maxnmatch:re2c*/
	yypmatch := make([]int, YYMAXNMATCH*2)
	var yynmatch int
	var yyt1, yyt2, yyt3, yyt4 int
	_ = yynmatch

	for { /*!use:re2c:base_template
		re2c:posix-captures   = 1;
		re2c:case-insensitive = 1;

		rxDay = [0-3]?[0-9];
		rxYear = 199[0-9]|20[0-3][0-9];
		rxMonth = January?|February?|March|A[pv]ril|Ma[iy]|Ju(!n[ei]|l[iy])|August|September|O[ck]tober|November|De[csz]ember|Jan|Feb|M[aä]r|Apr|Ju[ln]|Aug|Sep|O[ck]t|Nov|De[cz]|Januari|Februari|M(!aret|ei)|Agustus|Jänner|Feber|März|janvier|février|mars|jui(!n|llet)|aout|septembre|octobre|novembre|décembre|Ocak|Şubat|Mart|Nisan|Mayıs|Haziran|Temmuz|Ağustos|E(!ylül|kim)|Kasım|Aralık|Oca|Şub|Mar|Nis|Haz|Tem|Ağu|E(!yl|ki)|Kas|Ara;
		rxMDY = ({rxMonth})[\t\n\f\r ]({rxDay})(!st|nd|rd|th)?,?[\t\n\f\r ]({rxYear});
		rxDMY = ({rxDay})(!st|nd|rd|th|[.])?[\t\n\f\r ](!of[\t\n\f\r ])?({rxMonth})[,.]?[\t\n\f\r ]({rxYear});

		{rxMDY} {
			month = input[yypmatch[2]:yypmatch[3]]
			day = input[yypmatch[4]:yypmatch[5]]
			year = input[yypmatch[6]:yypmatch[7]]
			ok = true
			return
		}

		{rxDMY} {
			day = input[yypmatch[2]:yypmatch[3]]
			month = input[yypmatch[4]:yypmatch[5]]
			year = input[yypmatch[6]:yypmatch[7]]
			ok = true
			return
		}

		* { continue }
		$ { return }
		*/
	}
}

// PYTHON NAME: TEXT_DATE_PATTERN
// original pattern: [.:,_/ -]|^\d+$
func MatchTextDatePattern(input string) bool {
	var cursor int
	input += string(rune(0)) // add terminating null
	limit := len(input) - 1  // limit points at the terminating null

	// Capturing groups
	/*!maxnmatch:re2c*/
	yypmatch := make([]int, YYMAXNMATCH*2)
	var yynmatch int
	var yyt1 int
	_ = yynmatch

	for { /*!use:re2c:base_template
		re2c:posix-captures   = 1;

		[ ,-/:_] {
			// Handle [.:,_/ -]
			return true
		}

		[0-9]+ {
			// Handle ^\d+$
			if yypmatch[0] == 0 && yypmatch[1] == limit {
				return true
			}
			continue
		}

		* { continue }
		$ { return false }
		*/
	}
}
