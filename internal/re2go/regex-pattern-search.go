// Code generated by re2c 3.1, DO NOT EDIT.
package re2go

// PYTHON NAME: TIMESTAMP_PATTERN
// Given the following pattern:
//
// - day: [0-3]?[0-9]
// - month: [0-1]?[0-9]
// - year: 199[0-9]|20[0-3][0-9]
//
// Its original pattern is: (?i)((?:year)-(?:month)-(?:day)).[0-9]{2}:[0-9]{2}:[0-9]{2}
func TimestampPatternSubmatch(input string) ([]string, int) {
	var cursor, marker int
	input += string(rune(0)) // add terminating null
	limit := len(input) - 1  // limit points at the terminating null
	_ = marker

	// Variable for capturing parentheses (twice the number of groups).
	const YYMAXNMATCH = 2

	yypmatch := make([]int, YYMAXNMATCH*2)
	var yynmatch int
	_ = yynmatch

	// Autogenerated tag variables used by the lexer to track tag values.
	var yyt1 int
	_ = yyt1
	var yyt2 int
	_ = yyt2
	var yyt3 int
	_ = yyt3

	for {
		{
			var yych byte
			yych = input[cursor]
			switch yych {
			case '1':
				yyt1 = cursor
				goto yy3
			case '2':
				yyt1 = cursor
				goto yy4
			default:
				if limit <= cursor {
					goto yy34
				}
				goto yy1
			}
		yy1:
			cursor++
		yy2:
			{
				continue
			}
		yy3:
			cursor++
			marker = cursor
			yych = input[cursor]
			switch yych {
			case '9':
				goto yy5
			default:
				goto yy2
			}
		yy4:
			cursor++
			marker = cursor
			yych = input[cursor]
			switch yych {
			case '0':
				goto yy7
			default:
				goto yy2
			}
		yy5:
			cursor++
			yych = input[cursor]
			switch yych {
			case '9':
				goto yy8
			default:
				goto yy6
			}
		yy6:
			cursor = marker
			goto yy2
		yy7:
			cursor++
			yych = input[cursor]
			switch yych {
			case '0', '1', '2', '3':
				goto yy8
			default:
				goto yy6
			}
		yy8:
			cursor++
			yych = input[cursor]
			switch yych {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				goto yy9
			default:
				goto yy6
			}
		yy9:
			cursor++
			yych = input[cursor]
			switch yych {
			case '-':
				goto yy10
			default:
				goto yy6
			}
		yy10:
			cursor++
			yych = input[cursor]
			switch yych {
			case '0', '1':
				goto yy11
			case '2', '3', '4', '5', '6', '7', '8', '9':
				goto yy12
			default:
				goto yy6
			}
		yy11:
			cursor++
			yych = input[cursor]
			switch yych {
			case '-':
				goto yy13
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				goto yy12
			default:
				goto yy6
			}
		yy12:
			cursor++
			yych = input[cursor]
			switch yych {
			case '-':
				goto yy13
			default:
				goto yy6
			}
		yy13:
			cursor++
			yych = input[cursor]
			switch yych {
			case '0', '1', '2', '3':
				goto yy14
			case '4', '5', '6', '7', '8', '9':
				goto yy15
			default:
				goto yy6
			}
		yy14:
			cursor++
			yych = input[cursor]
			switch yych {
			case 0x00:
				fallthrough
			case 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, '\t':
				fallthrough
			case '\v', '\f', '\r', 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E, 0x1F, ' ', '!', '"', '#', '$', '%', '&', '\'', '(', ')', '*', '+', ',', '-', '.', '/':
				fallthrough
			case ':', ';', '<', '=', '>', '?', '@', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z', '[', '\\', ']', '^', '_', '`', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', '{', '|', '}', '~', 0x7F:
				if limit <= cursor {
					goto yy6
				}
				yyt2 = cursor
				goto yy16
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				yyt2 = cursor
				goto yy17
			case 0xC2, 0xC3, 0xC4, 0xC5, 0xC6, 0xC7, 0xC8, 0xC9, 0xCA, 0xCB, 0xCC, 0xCD, 0xCE, 0xCF, 0xD0, 0xD1, 0xD2, 0xD3, 0xD4, 0xD5, 0xD6, 0xD7, 0xD8, 0xD9, 0xDA, 0xDB, 0xDC, 0xDD, 0xDE, 0xDF:
				yyt2 = cursor
				goto yy18
			case 0xE0:
				yyt2 = cursor
				goto yy19
			case 0xE1, 0xE2, 0xE3, 0xE4, 0xE5, 0xE6, 0xE7, 0xE8, 0xE9, 0xEA, 0xEB, 0xEC, 0xED, 0xEE, 0xEF:
				yyt2 = cursor
				goto yy20
			case 0xF0:
				yyt2 = cursor
				goto yy21
			case 0xF1, 0xF2, 0xF3:
				yyt2 = cursor
				goto yy22
			case 0xF4:
				yyt2 = cursor
				goto yy23
			default:
				goto yy6
			}
		yy15:
			cursor++
			yych = input[cursor]
			switch yych {
			case 0x00:
				fallthrough
			case 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, '\t':
				fallthrough
			case '\v', '\f', '\r', 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E, 0x1F, ' ', '!', '"', '#', '$', '%', '&', '\'', '(', ')', '*', '+', ',', '-', '.', '/', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', ':', ';', '<', '=', '>', '?', '@', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z', '[', '\\', ']', '^', '_', '`', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', '{', '|', '}', '~', 0x7F:
				if limit <= cursor {
					goto yy6
				}
				yyt2 = cursor
				goto yy16
			case 0xC2, 0xC3, 0xC4, 0xC5, 0xC6, 0xC7, 0xC8, 0xC9, 0xCA, 0xCB, 0xCC, 0xCD, 0xCE, 0xCF, 0xD0, 0xD1, 0xD2, 0xD3, 0xD4, 0xD5, 0xD6, 0xD7, 0xD8, 0xD9, 0xDA, 0xDB, 0xDC, 0xDD, 0xDE, 0xDF:
				yyt2 = cursor
				goto yy18
			case 0xE0:
				yyt2 = cursor
				goto yy19
			case 0xE1, 0xE2, 0xE3, 0xE4, 0xE5, 0xE6, 0xE7, 0xE8, 0xE9, 0xEA, 0xEB, 0xEC, 0xED, 0xEE, 0xEF:
				yyt2 = cursor
				goto yy20
			case 0xF0:
				yyt2 = cursor
				goto yy21
			case 0xF1, 0xF2, 0xF3:
				yyt2 = cursor
				goto yy22
			case 0xF4:
				yyt2 = cursor
				goto yy23
			default:
				goto yy6
			}
		yy16:
			cursor++
			yych = input[cursor]
			switch yych {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				goto yy24
			default:
				goto yy6
			}
		yy17:
			cursor++
			yych = input[cursor]
			switch yych {
			case 0x00:
				fallthrough
			case 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, '\t':
				fallthrough
			case '\v', '\f', '\r', 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E, 0x1F, ' ', '!', '"', '#', '$', '%', '&', '\'', '(', ')', '*', '+', ',', '-', '.', '/':
				fallthrough
			case ':', ';', '<', '=', '>', '?', '@', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z', '[', '\\', ']', '^', '_', '`', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', '{', '|', '}', '~', 0x7F:
				if limit <= cursor {
					goto yy6
				}
				yyt2 = cursor
				goto yy16
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				yyt3 = cursor
				goto yy25
			case 0xC2, 0xC3, 0xC4, 0xC5, 0xC6, 0xC7, 0xC8, 0xC9, 0xCA, 0xCB, 0xCC, 0xCD, 0xCE, 0xCF, 0xD0, 0xD1, 0xD2, 0xD3, 0xD4, 0xD5, 0xD6, 0xD7, 0xD8, 0xD9, 0xDA, 0xDB, 0xDC, 0xDD, 0xDE, 0xDF:
				yyt2 = cursor
				goto yy18
			case 0xE0:
				yyt2 = cursor
				goto yy19
			case 0xE1, 0xE2, 0xE3, 0xE4, 0xE5, 0xE6, 0xE7, 0xE8, 0xE9, 0xEA, 0xEB, 0xEC, 0xED, 0xEE, 0xEF:
				yyt2 = cursor
				goto yy20
			case 0xF0:
				yyt2 = cursor
				goto yy21
			case 0xF1, 0xF2, 0xF3:
				yyt2 = cursor
				goto yy22
			case 0xF4:
				yyt2 = cursor
				goto yy23
			default:
				goto yy6
			}
		yy18:
			cursor++
			yych = input[cursor]
			switch yych {
			case 0x80, 0x81, 0x82, 0x83, 0x84, 0x85, 0x86, 0x87, 0x88, 0x89, 0x8A, 0x8B, 0x8C, 0x8D, 0x8E, 0x8F, 0x90, 0x91, 0x92, 0x93, 0x94, 0x95, 0x96, 0x97, 0x98, 0x99, 0x9A, 0x9B, 0x9C, 0x9D, 0x9E, 0x9F, 0xA0, 0xA1, 0xA2, 0xA3, 0xA4, 0xA5, 0xA6, 0xA7, 0xA8, 0xA9, 0xAA, 0xAB, 0xAC, 0xAD, 0xAE, 0xAF, 0xB0, 0xB1, 0xB2, 0xB3, 0xB4, 0xB5, 0xB6, 0xB7, 0xB8, 0xB9, 0xBA, 0xBB, 0xBC, 0xBD, 0xBE, 0xBF:
				goto yy16
			default:
				goto yy6
			}
		yy19:
			cursor++
			yych = input[cursor]
			switch yych {
			case 0xA0, 0xA1, 0xA2, 0xA3, 0xA4, 0xA5, 0xA6, 0xA7, 0xA8, 0xA9, 0xAA, 0xAB, 0xAC, 0xAD, 0xAE, 0xAF, 0xB0, 0xB1, 0xB2, 0xB3, 0xB4, 0xB5, 0xB6, 0xB7, 0xB8, 0xB9, 0xBA, 0xBB, 0xBC, 0xBD, 0xBE, 0xBF:
				goto yy18
			default:
				goto yy6
			}
		yy20:
			cursor++
			yych = input[cursor]
			switch yych {
			case 0x80, 0x81, 0x82, 0x83, 0x84, 0x85, 0x86, 0x87, 0x88, 0x89, 0x8A, 0x8B, 0x8C, 0x8D, 0x8E, 0x8F, 0x90, 0x91, 0x92, 0x93, 0x94, 0x95, 0x96, 0x97, 0x98, 0x99, 0x9A, 0x9B, 0x9C, 0x9D, 0x9E, 0x9F, 0xA0, 0xA1, 0xA2, 0xA3, 0xA4, 0xA5, 0xA6, 0xA7, 0xA8, 0xA9, 0xAA, 0xAB, 0xAC, 0xAD, 0xAE, 0xAF, 0xB0, 0xB1, 0xB2, 0xB3, 0xB4, 0xB5, 0xB6, 0xB7, 0xB8, 0xB9, 0xBA, 0xBB, 0xBC, 0xBD, 0xBE, 0xBF:
				goto yy18
			default:
				goto yy6
			}
		yy21:
			cursor++
			yych = input[cursor]
			switch yych {
			case 0x90, 0x91, 0x92, 0x93, 0x94, 0x95, 0x96, 0x97, 0x98, 0x99, 0x9A, 0x9B, 0x9C, 0x9D, 0x9E, 0x9F, 0xA0, 0xA1, 0xA2, 0xA3, 0xA4, 0xA5, 0xA6, 0xA7, 0xA8, 0xA9, 0xAA, 0xAB, 0xAC, 0xAD, 0xAE, 0xAF, 0xB0, 0xB1, 0xB2, 0xB3, 0xB4, 0xB5, 0xB6, 0xB7, 0xB8, 0xB9, 0xBA, 0xBB, 0xBC, 0xBD, 0xBE, 0xBF:
				goto yy20
			default:
				goto yy6
			}
		yy22:
			cursor++
			yych = input[cursor]
			switch yych {
			case 0x80, 0x81, 0x82, 0x83, 0x84, 0x85, 0x86, 0x87, 0x88, 0x89, 0x8A, 0x8B, 0x8C, 0x8D, 0x8E, 0x8F, 0x90, 0x91, 0x92, 0x93, 0x94, 0x95, 0x96, 0x97, 0x98, 0x99, 0x9A, 0x9B, 0x9C, 0x9D, 0x9E, 0x9F, 0xA0, 0xA1, 0xA2, 0xA3, 0xA4, 0xA5, 0xA6, 0xA7, 0xA8, 0xA9, 0xAA, 0xAB, 0xAC, 0xAD, 0xAE, 0xAF, 0xB0, 0xB1, 0xB2, 0xB3, 0xB4, 0xB5, 0xB6, 0xB7, 0xB8, 0xB9, 0xBA, 0xBB, 0xBC, 0xBD, 0xBE, 0xBF:
				goto yy20
			default:
				goto yy6
			}
		yy23:
			cursor++
			yych = input[cursor]
			switch yych {
			case 0x80, 0x81, 0x82, 0x83, 0x84, 0x85, 0x86, 0x87, 0x88, 0x89, 0x8A, 0x8B, 0x8C, 0x8D, 0x8E, 0x8F:
				goto yy20
			default:
				goto yy6
			}
		yy24:
			cursor++
			yych = input[cursor]
			switch yych {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				goto yy26
			default:
				goto yy6
			}
		yy25:
			cursor++
			yych = input[cursor]
			switch yych {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				goto yy27
			default:
				goto yy6
			}
		yy26:
			cursor++
			yych = input[cursor]
			switch yych {
			case ':':
				goto yy28
			default:
				goto yy6
			}
		yy27:
			cursor++
			yych = input[cursor]
			switch yych {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				yyt2 = yyt3
				goto yy26
			case ':':
				goto yy28
			default:
				goto yy6
			}
		yy28:
			cursor++
			yych = input[cursor]
			switch yych {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				goto yy29
			default:
				goto yy6
			}
		yy29:
			cursor++
			yych = input[cursor]
			switch yych {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				goto yy30
			default:
				goto yy6
			}
		yy30:
			cursor++
			yych = input[cursor]
			switch yych {
			case ':':
				goto yy31
			default:
				goto yy6
			}
		yy31:
			cursor++
			yych = input[cursor]
			switch yych {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				goto yy32
			default:
				goto yy6
			}
		yy32:
			cursor++
			yych = input[cursor]
			switch yych {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				goto yy33
			default:
				goto yy6
			}
		yy33:
			cursor++
			yynmatch = 2
			yypmatch[2] = yyt1
			yypmatch[3] = yyt2
			yypmatch[0] = yyt1
			yypmatch[1] = cursor
			{
				return getAllSubmatch(input, YYMAXNMATCH, yypmatch)
			}
		yy34:
			{
				return nil, -1
			}
		}

	}
}