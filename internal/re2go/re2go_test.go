package re2go

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_IdiosyncracyPatternSubmatch(t *testing.T) {
	var str string

	assertSuccess := func(str string, expectedStart int, expectedParts ...string) {
		parts, start := IdiosyncracyPatternSubmatch(str)
		assert.Len(t, parts, len(expectedParts))
		for i := range expectedParts {
			assert.Equal(t, expectedParts[i], parts[i])
		}
		assert.Equal(t, expectedStart, start)
	}

	assertFail := func(str string) {
		parts, _ := IdiosyncracyPatternSubmatch(str)
		assert.Len(t, parts, 0)
	}

	// Should match English pattern
	str = "This article was updated on 2023/09/15, with the latest information included."
	assertSuccess(str, 19, "dated on 2023/09/15", "2023", "09", "15")

	str = "Published: 2022/08/03, this study changed the field forever."
	assertSuccess(str, 0, "Published: 2022/08/03", "2022", "08", "03")

	str = "The release date: 2021.07.12 saw massive improvements in the technology."
	assertSuccess(str, 12, "date: 2021.07.12", "2021", "07", "12")

	str = "Released date 20.04.2019 for the new product line."
	assertSuccess(str, 9, "date 20.04.2019", "20", "04", "2019")

	str = "Published on 15.03.2020, the document provided significant insights."
	assertSuccess(str, 10, "on 15.03.2020", "15", "03", "2020")

	str = "Final update on 11/08/2021 confirmed all prior assumptions."
	assertSuccess(str, 8, "date on 11/08/2021", "11", "08", "2021")

	str = "The file was last modified on 2020/05/10, according to the logs."
	assertSuccess(str, 27, "on 2020/05/10", "2020", "05", "10")

	str = "The paper was originally published 03/12/2018 and later revised."
	assertSuccess(str, 25, "published 03/12/2018", "03", "12", "2018")

	str = "On 2021.09.30, the latest version was uploaded."
	assertSuccess(str, 0, "On 2021.09.30", "2021", "09", "30")

	str = "An update on the date 2022/06/17 was recorded in the system."
	assertSuccess(str, 5, "date on the date 2022/06/17", "2022", "06", "17")

	str = "Revised date: 01.11.2020 after careful analysis."
	assertSuccess(str, 8, "date: 01.11.2020", "01", "11", "2020")

	str = "Published 13.03.2019, this report provided key findings."
	assertSuccess(str, 0, "Published 13.03.2019", "13", "03", "2019")

	str = "On 2019/07/22, they finalized the specifications."
	assertSuccess(str, 0, "On 2019/07/22", "2019", "07", "22")

	str = "The event took place on date: 15/02/2021, a memorable occasion."
	assertSuccess(str, 24, "date: 15/02/2021", "15", "02", "2021")

	str = "The new policy was implemented on 2020.04.05, ensuring compliance."
	assertSuccess(str, 31, "on 2020.04.05", "2020", "04", "05")

	str = "Updated on 2021/01/20, this version included bug fixes."
	assertSuccess(str, 2, "dated on 2021/01/20", "2021", "01", "20")

	str = "Published on date 2022/12/01, this marked a turning point."
	assertSuccess(str, 13, "date 2022/12/01", "2022", "12", "01")

	str = "Final update was made 07.07.2021, as indicated by the records."
	assertSuccess(str, 8, "date was made 07.07.2021", "07", "07", "2021")

	str = "They shared the news on 03/11/2019 after the meeting."
	assertSuccess(str, 21, "on 03/11/2019", "03", "11", "2019")

	str = "We received the package on 2021/03/22, confirming the delivery date.	"
	assertSuccess(str, 24, "on 2021/03/22", "2021", "03", "22")

	// Shouldn't match English pattern
	assertFail("The date: 1/123/2023 is incorrect, as the day is invalid.")
	assertFail("Published on 2023-13-01, a date that does not exist.")
	assertFail("Updated on 2/11/1, with a year too short.")
	assertFail("The book was written on 2023-05/15, the format is inconsistent.")
	assertFail("In 31.13.2021, there was no such month.")
	assertFail("Published 12/2023.05, but the format is wrong.")
	assertFail("Updated: 16.2023.12, a completely invalid date.")
	assertFail("Recorded date: 10-25.1234, but the day and month are swapped.")
	assertFail("He mentioned it's released in 2023/10.15, which should be consistent.")
	assertFail("The event took place on 2021-20-02, but the day is too high.")

	// Should match German pattern
	str = "Der Bericht wurde am Veröffentlicht am: 12.08.2023 abgeschlossen."
	assertSuccess(str, 21, "Veröffentlicht am: 12.08.2023", "12", "08", "2023")

	str = "Datum: 15.03.2022, als der Vertrag unterzeichnet wurde."
	assertSuccess(str, 0, "Datum: 15.03.2022", "15", "03", "2022")

	str = "Stand: 10.10.2021, die Zahlen wurden aktualisiert."
	assertSuccess(str, 0, "Stand: 10.10.2021", "10", "10", "2021")

	str = "Dieses Dokument wurde Veröffentlicht am 01.01.2020."
	assertSuccess(str, 22, "Veröffentlicht am 01.01.2020", "01", "01", "2020")

	str = "Das Meeting fand am Datum 05.07.2019 statt."
	assertSuccess(str, 20, "Datum 05.07.2019", "05", "07", "2019")

	str = "Stand: 30.11.2020, alle Informationen sind korrekt."
	assertSuccess(str, 0, "Stand: 30.11.2020", "30", "11", "2020")

	str = "Die Änderungen wurden am Veröffentlicht am 22.02.2021 vorgenommen."
	assertSuccess(str, 26, "Veröffentlicht am 22.02.2021", "22", "02", "2021")

	str = "Datum: 01.12.2022, das Projekt startete offiziell."
	assertSuccess(str, 0, "Datum: 01.12.2022", "01", "12", "2022")

	str = "Veröffentlicht am 05.09.2018, wurde das Update angekündigt."
	assertSuccess(str, 0, "Veröffentlicht am 05.09.2018", "05", "09", "2018")

	str = "Das System ist auf dem Stand: 25.05.2023."
	assertSuccess(str, 23, "Stand: 25.05.2023", "25", "05", "2023")

	str = "Datum: 12.10.2017, alle Daten wurden gesichert."
	assertSuccess(str, 0, "Datum: 12.10.2017", "12", "10", "2017")

	str = "Die offizielle Veröffentlichung war Veröffentlicht am 14.02.2020."
	assertSuccess(str, 37, "Veröffentlicht am 14.02.2020", "14", "02", "2020")

	str = "Stand: 29.08.2016, das Update war stabil."
	assertSuccess(str, 0, "Stand: 29.08.2016", "29", "08", "2016")

	str = "Datum: 20.06.2021, alle Details sind korrekt."
	assertSuccess(str, 0, "Datum: 20.06.2021", "20", "06", "2021")

	str = "Veröffentlicht am 17.11.2023, begann das Event."
	assertSuccess(str, 0, "Veröffentlicht am 17.11.2023", "17", "11", "2023")

	str = "Datum: 08.08.2019, der Bericht wurde abgeschlossen."
	assertSuccess(str, 0, "Datum: 08.08.2019", "08", "08", "2019")

	str = "Das neue Gesetz wurde Veröffentlicht am 01.04.2022."
	assertSuccess(str, 22, "Veröffentlicht am 01.04.2022", "01", "04", "2022")

	str = "Stand: 30.09.2023, die Arbeiten sind abgeschlossen."
	assertSuccess(str, 0, "Stand: 30.09.2023", "30", "09", "2023")

	str = "Datum: 11.12.2015, als das Projekt abgeschlossen wurde."
	assertSuccess(str, 0, "Datum: 11.12.2015", "11", "12", "2015")

	str = "Veröffentlicht am 06.06.2020, wurde der Artikel geschrieben."
	assertSuccess(str, 0, "Veröffentlicht am 06.06.2020", "06", "06", "2020")

	// Shouldn't match German pattern
	assertFail("Stand: 15-03-2022, alle Informationen wurden bestätigt.")
	assertFail("Veröffentlicht am: 12/08/2023, war der Bericht zugänglich.")
	assertFail("Datum am: 15.03.22, als das Projekt veröffentlicht wurde.")
	assertFail("Stand: 2021.10.10, die Daten wurden aktualisiert.")
	assertFail("Dieses Dokument wurde Veröffentlicht: 01.01.20.")
	assertFail("Stand am: 30.11.20, alle Informationen sind hier.")
	assertFail("Veröffentlicht 5.09.2018, das Update wurde geteilt.")
	assertFail("Datum: 12/10/17, der Vertrag wurde unterzeichnet.")
	assertFail("Stand: 29-8-16, der Bericht wurde hochgeladen.")
	assertFail("Datum: 20-06-21, der Status ist aktuell.")

	// Should match Turkiye language
	str = "Son güncellenme tarihi: 12/08/2023 itibariyle sistem çalışıyor."
	assertSuccess(str, 4, "güncellenme tarihi: 12/08/2023", "12", "08", "2023")

	str = "Yayınlanma: 15.03.2022 tarihinde makale yayımlandı."
	assertSuccess(str, 0, "Yayınlanma: 15.03.2022", "15", "03", "2022")

	str = "Bu belge, güncellenme tarihi: 01.01.2020 olarak işaretlenmiş."
	assertSuccess(str, 10, "güncellenme tarihi: 01.01.2020", "01", "01", "2020")

	str = "Yayımlanma: 05.07.2019 tarihinde duyuruldu."
	assertSuccess(str, 0, "Yayımlanma: 05.07.2019", "05", "07", "2019")

	str = "Dosya güncellenme  tarihi: 10.10.2021, tüm bilgiler doğru."
	assertSuccess(str, 6, "güncellenme  tarihi: 10.10.2021", "10", "10", "2021")

	str = "Güncellenme tarihi: 30/11/2020, sistem kontrol edildi."
	assertSuccess(str, 0, "Güncellenme tarihi: 30/11/2020", "30", "11", "2020")

	str = "Yayınlanma tarihi: 22.02.2021, içerik yayınlandı."
	assertSuccess(str, 0, "Yayınlanma tarihi: 22.02.2021", "22", "02", "2021")

	str = "Dosyanın yayımlanma tarihi 01.12.2022 olarak belirlendi."
	assertSuccess(str, 10, "yayımlanma tarihi 01.12.2022", "01", "12", "2022")

	str = "Güncellenme 15.09.2018, yazılım yenilendi."
	assertSuccess(str, 0, "Güncellenme 15.09.2018", "15", "09", "2018")

	str = "Bu belgede en son güncelleme tarihi: 25/05/2023."
	assertSuccess(str, 18, "güncelleme tarihi: 25/05/2023", "25", "05", "2023")

	str = "Bu belge 12/08/2023'te güncellendi."
	assertSuccess(str, 9, "12/08/2023'te güncellendi", "12", "08", "2023")

	str = "Makale 15.03.2022’de yayımlandı ve kullanıma açıldı."
	assertSuccess(str, 7, "15.03.2022’de yayımlandı", "15", "03", "2022")

	str = "Rapor 01/01/2020'da güncellendi, yeni veriler eklendi."
	assertSuccess(str, 6, "01/01/2020'da güncellendi", "01", "01", "2020")

	str = "Duyuru 05.07.2019 tarihinde yayımlandı."
	assertSuccess(str, 7, "05.07.2019 tarihinde yayımlandı", "05", "07", "2019")

	str = "Belge 10.10.2021'de güncellendi, içerik yenilendi."
	assertSuccess(str, 6, "10.10.2021'de güncellendi", "10", "10", "2021")

	str = "Dosya 30/11/2020'te güncellendi."
	assertSuccess(str, 6, "30/11/2020'te güncellendi", "30", "11", "2020")

	str = "İçerik 22.02.2021’de yayımlandı."
	assertSuccess(str, 9, "22.02.2021’de yayımlandı", "22", "02", "2021")

	str = "Rapor 01/12/2022'da güncellendi ve onaylandı."
	assertSuccess(str, 6, "01/12/2022'da güncellendi", "01", "12", "2022")

	str = "Proje 25/05/2023 tarihinde yayımlandı."
	assertSuccess(str, 6, "25/05/2023 tarihinde yayımlandı", "25", "05", "2023")

	// Shouldn't match Turkiye language
	assertFail("Yayınlanma tarihi 15-03-22, bu dosyada yer alıyor.")
	assertFail("Guncelleme tarihi: 12.8.2023, yanlış formatta.")
	assertFail("Son güncellenme: 2021.10.10 olarak gösteriliyor.")
	assertFail("Dosya guncelleme  tarihi 01/01/20, eksik bilgi.")
	assertFail("Güncellenme tarihi: 5-9-2018, format hatalı.")
	assertFail("Belge 12-08-2023 tarihinde güncellendi, tarih formatı yanlış.")
	assertFail("İçerik 15/03/22'de yayımiandı, eksik yıl bilgisi var.")
	assertFail("Dosya 2020.01.01'te güncelendi, format hatalı.")
	assertFail("Duyuru 05/7/2019'da yaımlandı, gün ve ay tam değil.")
	assertFail("Rapor 10-10-21’de yayımlandı, eksik yıl formatı.")
}

func Test_SelectYmdPattern(t *testing.T) {
	assertSuccess := func(s string, expected string) {
		indexes := SelectYmdPattern(s)
		assert.Len(t, indexes, 1)

		match := s[indexes[0][0]:indexes[0][1]]
		assert.Equal(t, expected, match)
	}

	assertFail := func(s string) {
		indexes := SelectYmdPattern(s)
		assert.Empty(t, indexes)
	}

	assertSuccess("The event occurred on 12/08/2023 and was a success.", " 12/08/2023 ")
	assertSuccess("Document finalized on 15.03.2022, as noted.", " 15.03.2022,")
	assertSuccess("The meeting was scheduled for 01-01-2020.", " 01-01-2020.")
	assertSuccess("The system update happened on 05/07/2019.", " 05/07/2019.")
	assertSuccess("Please note the update on 10.10.2021.", " 10.10.2021.")
	assertSuccess("The last revision was on 30/11/2020.", " 30/11/2020.")
	assertSuccess("The article was published on 22-02-2021.", " 22-02-2021.")
	assertSuccess("The records show 01/12/2022 as the latest entry.", " 01/12/2022 ")
	assertSuccess("The software was released on 15.09.2018.", " 15.09.2018.")
	assertSuccess("The new policy is effective from 25/05/2023.", " 25/05/2023.")

	assertFail("The report was due on 2023/12/08, incorrect format.")
	assertFail("The project started on 15.03.22, missing full year.")
	assertFail("Event scheduled for 05/07/19, missing full year.")
	assertFail("The update happened 10/10/21, missing separators.")
}

func Test_SlashesPattern(t *testing.T) {
	assertSuccess := func(s string, expected string) {
		indexes := SlashesPattern(s)
		assert.Len(t, indexes, 1)

		match := s[indexes[0][0]:indexes[0][1]]
		assert.Equal(t, expected, match)
	}

	assertFail := func(s string) {
		indexes := SlashesPattern(s)
		assert.Empty(t, indexes)
	}

	assertSuccess("http://example.com/test/12/05/21/sample", "/12/05/21/")
	assertSuccess("https://mysite.org/article/9/4/99/about", "/9/4/99/")
	assertSuccess("https://website.com/post/03/11/20/page", "/03/11/20/")
	assertSuccess("http://news.net/view/30.10.99/info", "/30.10.99/")
	assertSuccess("https://example.org/entry/31.01.20/about", "/31.01.20/")
	assertSuccess("http://domain.com/path/3/12/29/article", "/3/12/29/")
	assertSuccess("https://randomsite.net/page/01.11.21/details", "/01.11.21/")
	assertSuccess("http://othersite.com/report/15.08.19/section", "/15.08.19/")
	assertSuccess("https://demo.com/blog/29.09.99/post", "/29.09.99/")
	assertSuccess("http://sample.net/path/30.05.29/test", "/30.05.29/")

	assertFail("http://example.com/test/42/13/21/sample")
	assertFail("https://mysite.org/article/9/4/210/about")
	assertFail("https://website.com/post/40.12.20/page")
	assertFail("http://news.net/view/30-10-99/info")
	assertFail("https://example.org/entry/31.23.20/about")
}

func Test_MmYyyyPattern(t *testing.T) {
	assertSuccess := func(s string, expected string) {
		indexes := MmYyyyPattern(s)
		assert.Len(t, indexes, 1)

		match := s[indexes[0][0]:indexes[0][1]]
		assert.Equal(t, expected, match)
	}

	assertFail := func(s string) {
		indexes := MmYyyyPattern(s)
		assert.Empty(t, indexes)
	}

	assertSuccess("I have an appointment on 09/2023 at the clinic.", " 09/2023 ")
	assertSuccess("The meeting was rescheduled to 3.2022.", " 3.2022.")
	assertSuccess("We will start the project by 07-2021.", " 07-2021.")
	assertSuccess("His birthday is 12.2025, and we're planning a surprise.", " 12.2025,")
	assertSuccess("She moved to the new city in 04/2020.", " 04/2020.")
	assertSuccess("The contract starts from 01-2024.", " 01-2024.")
	assertSuccess("The event took place on 8/2019.", " 8/2019.")
	assertSuccess("We completed the survey in 06.2021.", " 06.2021.")
	assertSuccess("The package was delivered on 10-2022.", " 10-2022.")
	assertSuccess("Their vacation is scheduled for 11/2026.", " 11/2026.")

	assertFail("He left on 22/2021, which was unexpected.")
	assertFail("Her visit is scheduled for 02-219.")
	assertFail("I submitted the paper on 07.22.")
	assertFail("The exam will be held on 8/20211.")
	assertFail("We expect the results by 01/3020.")
}

func Test_FindLongTextPattern(t *testing.T) {
	success := func(s, year, month, day string) {
		y, m, d, ok := FindLongTextPattern(s)
		assert.True(t, ok)
		assert.Equal(t, year, y)
		assert.Equal(t, month, m)
		assert.Equal(t, day, d)
	}

	fail := func(s string) {
		y, m, d, ok := FindLongTextPattern(s)
		assert.False(t, ok)
		assert.Empty(t, y)
		assert.Empty(t, m)
		assert.Empty(t, d)
	}

	success("I was born on January 1st, 2000.", "2000", "January", "1")
	success("He arrived on February 29th, 2020.", "2020", "February", "29")
	success("The conference is scheduled for 5th of March, 2021.", "2021", "March", "5")
	success("They left on July 4, 1999.", "1999", "July", "4")
	success("The event took place on 31st October 2005.", "2005", "October", "31")
	success("Her birthday is on the 21st of August, 1998.", "1998", "August", "21")
	success("We met on 12th July, 2010.", "2010", "July", "12")
	success("His graduation was on 3rd of April 2002.", "2002", "April", "3")
	success("The anniversary party is on 9th September 2022.", "2022", "September", "9")
	success("She started work on November 11, 2018.", "2018", "November", "11")
	success("Er wurde am 1. Januar 2000 geboren.", "2000", "Januar", "1")
	success("Die Konferenz fand am 15. März 2022 statt.", "2022", "März", "15")
	success("Sie reiste am 31. Oktober 1999 ab.", "1999", "Oktober", "31")
	success("Wir feiern am 20. Juli 2021.", "2021", "Juli", "20")
	success("Der Vertrag begann am 4. April 2005.", "2005", "April", "4")
	success("Tatildeydik 12 Temmuz 2019.", "2019", "Temmuz", "12")
	success("Onlar 5 Haziran 2020'de taşındı.", "2020", "Haziran", "5")
	success("Mezuniyet töreni 18. Kasım 2003'te gerçekleşti.", "2003", "Kasım", "18")
	success("Toplantı 29 Şubat 2020'de yapılacak.", "2020", "Şubat", "29")
	success("Festival 10 Ağustos 2022'de başlayacak.", "2022", "Ağustos", "10")

	fail("The meeting is set for 11 Januarie, 2023.")
	fail("He was born on 29th of February, 2107.")
	fail("We went on March 41, 2001.")
	fail("She celebrated her birthday on 16sd May, 2020.")
	fail("They arrived on 31st of the April, 2021.")
	fail("Er kam am 32. Januir 2001 an.")
	fail("Das Treffen war am 29, Februar 1995.")
	fail("Toplantı 31ts Haziran 2021'de yapıldı.")
	fail("Onlar 14st Ekam 2023'te geldi.")
	fail("Das Fest findet am 33. Aprıl 2008 statt.")
}

func Test_TimestampPatternSubmatch(t *testing.T) {
	success := func(s, expected string) {
		parts, _ := TimestampPatternSubmatch(s)
		assert.Len(t, parts, 2)
		assert.Equal(t, expected, parts[1])
	}

	fail := func(s string) {
		parts, _ := TimestampPatternSubmatch(s)
		assert.Empty(t, parts)
	}

	success("The event occurred on 2021-07-15 14:35:20.", "2021-07-15")
	success("The meeting was logged at 1999-12-31 23:59:59.", "1999-12-31")
	success("His flight landed on 2000-01-01T00:00:00.", "2000-01-01")
	success("The system rebooted on 2005-06-10.12:45:30.", "2005-06-10")
	success("She sent the email on 2022-08-03 09:12:45.", "2022-08-03")
	success("The server crashed on 2010-09-15/17:30:05.", "2010-09-15")

	fail("The event occurred on 2024-07-15 at 14:35:20.")
	fail("The meeting was logged at 1999-13-31, 23:59:59.")
	fail("The system rebooted on 2005-06-32 at 12:45:30.")
}

func Test_CopyrightPattern(t *testing.T) {
	success := func(s, expectedMatch, expectedYear string) {
		indexes := CopyrightPattern(s)
		assert.Len(t, indexes, 1)

		match := s[indexes[0][0]:indexes[0][1]]
		assert.Equal(t, expectedMatch, match)

		year := s[indexes[0][2]:indexes[0][3]]
		assert.Equal(t, expectedYear, year)
	}

	fail := func(s string) {
		indexes := CopyrightPattern(s)
		assert.Empty(t, indexes)
	}

	success("© 1998-2022 All rights reserved.", "© 1998-2022 ", "2022")
	success("Copyright 2000-2019 by Example Corp.", "Copyright 2000-2019 ", "2019")
	success("This content is © 2021.", "© 2021.", "2021")
	success("&copy; 1995-2020 by the author.", "&copy; 1995-2020 ", "2020")
	success("Published under (c) 2010-2021 terms.", "(c) 2010-2021 ", "2021")
	success("© 2023 Example Company.", "© 2023 ", "2023")
	success("Copyright 1997-2018. All rights reserved.", "Copyright 1997-2018.", "2018")
	success("&copy; 1999-2020 Corporation.", "&copy; 1999-2020 ", "2020")
	success("(c) 1996-2021, All rights reserved.", "(c) 1996-2021,", "2021")
	success("© 1995-2010 by John Doe.", "© 1995-2010 ", "2010")

	fail("Copyright 1925-3030.")
	fail("&copy; 1989-1995 by the publisher.")
	fail("Published in 2024 under (c).")
}

func Test_ThreePattern(t *testing.T) {
	success := func(s, expectedMatch string) {
		indexes := ThreePattern(s)
		assert.Len(t, indexes, 1)

		match := s[indexes[0][2]:indexes[0][3]]
		assert.Equal(t, expectedMatch, match)
	}

	fail := func(s string) {
		indexes := ThreePattern(s)
		assert.Empty(t, indexes)
	}

	success("https://example.com/2023/08/15/0", "2023/08/15")
	success("http://mysite.org/blog/2022/07/01/1", "2022/07/01")
	success("https://website.com/articles/1999/12/31/0", "1999/12/31")
	success("http://news.net/view/2020/11/05/1", "2020/11/05")
	success("https://demo.org/path/2021/06/20/0", "2021/06/20")
	success("http://content.com/reports/2023/02/28/1", "2023/02/28")
	success("https://example.net/events/2019/09/17/0", "2019/09/17")
	success("http://archive.org/2020/10/10/1", "2020/10/10")
	success("https://site.com/posts/2022/03/05/0", "2022/03/05")
	success("http://data.net/log/2021/04/30/1", "2021/04/30")

	fail("http://example.com/2023/8/15/0")
	fail("https://mysite.org/blog/2022/07/323/1")
	fail("http://website.com/articles/1999-12-31/0")
	fail("https://news.net/view/2020/11/05")
	fail("http://demo.org/path/202/06/20/0")
}

func Test_ThreeLoosePattern(t *testing.T) {
	success := func(s, expectedMatch string) {
		indexes := ThreeLoosePattern(s)
		assert.Len(t, indexes, 1)

		match := s[indexes[0][2]:indexes[0][3]]
		assert.Equal(t, expectedMatch, match)
	}

	fail := func(s string) {
		indexes := ThreeLoosePattern(s)
		assert.Empty(t, indexes)
	}

	success("https://example.com/2023/08/15/details", "2023/08/15")
	success("http://mysite.org/archive/1999-12-31/info", "1999-12-31")
	success("https://website.com/posts/2021.07.01/summary", "2021.07.01")
	success("http://news.net/view/2020/11/05/updates", "2020/11/05")
	success("https://demo.org/files/2019-09-17/report", "2019-09-17")
	success("http://content.com/docs/2022/03/05/document", "2022/03/05")
	success("https://example.net/data/2023.02.28/record", "2023.02.28")
	success("http://archive.org/2020/10/10/article", "2020/10/10")
	success("https://site.com/media/2018-06-21/photo", "2018-06-21")
	success("http://data.net/files/2021.04.30/details", "2021.04.30")

	fail("https://example.com/2023/8/15/details")
	fail("http://mysite.org/archive/19999-12-32/info")
	fail("https://website.com/posts/2021/07//01/summary")
	fail("http://news.net/view/202-11-05/updates")
	fail("https://demo.org/files/202/06/20/report")
}

func Test_DateStringsPattern(t *testing.T) {
	success := func(s, expectedMatch string) {
		indexes := DateStringsPattern(s)
		assert.Len(t, indexes, 1)

		match := s[indexes[0][2]:indexes[0][3]]
		assert.Equal(t, expectedMatch, match)
	}

	fail := func(s string) {
		indexes := DateStringsPattern(s)
		assert.Empty(t, indexes)
	}

	success("The file was created on X19501230X.", "X19501230X")
	success("I found document from Y19891101Y in the archives.", "Y19891101Y")
	success("The report was due by Z20050321Z, but it arrived late.", "Z20050321Z")
	success("Our event took place on A20190214A and was a great success.", "A20190214A")
	success("He said the date is B19740615B, so we should proceed accordingly.", "B19740615B")
	success("They moved into the house on M19651201M after signing the papers.", "M19651201M")
	success("The deadline for the submission is C20071130C, don't miss it.", "C20071130C")
	success("The last update was on W20230815W, according to the logs.", "W20230815W")
	success("Her birthday is on V19820602V, you should send her a card.", "V19820602V")
	success("The system was reset on K19991225K, and everything worked fine.", "K19991225K")

	fail("The meeting happened on 195101330, but we missed the announcement.")
	fail("He arrived on B119741131B, which seems like an invalid date.")
}

func Test_YyyyMmPattern(t *testing.T) {
	success := func(s, expectedMatch string) {
		indexes := YyyyMmPattern(s)
		assert.Len(t, indexes, 1)

		match := s[indexes[0][2]:indexes[0][3]]
		assert.Equal(t, expectedMatch, match)
	}

	fail := func(s string) {
		indexes := YyyyMmPattern(s)
		assert.Empty(t, indexes)
	}

	success("The invoice was sent on X2023/08X, please confirm.", "2023/08")
	success("Our meeting is scheduled for Y2019-04Y at noon.", "2019-04")
	success("The event happened on Z2020.12Z, and it was amazing.", "2020.12")
	success("We received the package on X2005-03X in good condition.", "2005-03")
	success("The product was released in Y2021.09Y after much anticipation.", "2021.09")
	success("Please review the document dated Z2010-11Z as soon as possible.", "2010-11")
	success("The last update was on A2022/07A, and it worked perfectly.", "2022/07")
	success("We expect the shipment in W2008/01W by the end of the week.", "2008/01")
	success("Her passport was issued on T2018-10T before the trip.", "2018-10")
	success("The article was published in N2023.06N last week.", "2023.06")

	fail("The order was placed on X2023/13X, but that date seems incorrect.")
	fail("The document is from Y2019-00Y, which doesn't look right.")
	fail("I checked the record dated Z2022.14Z, but found no information.")
	fail("His visa expired in W2009/15W, which cannot be valid.")
	fail("The delivery is scheduled for T2018-00T, but that date is impossible.")
}

func Test_SimplePattern(t *testing.T) {
	success := func(s, expectedMatch string) {
		indexes := SimplePattern(s)
		assert.Len(t, indexes, 1)

		match := s[indexes[0][2]:indexes[0][3]]
		assert.Equal(t, expectedMatch, match)
	}

	fail := func(s string) {
		indexes := SimplePattern(s)
		assert.Empty(t, indexes)
	}

	success("The company was founded in 1998 and grew rapidly.", "1998")
	success("He graduated from university in 2003 with honors.", "2003")
	success("Our project began in 2020 and is still ongoing.", "2020")
	success("The concert took place in 1999 and was unforgettable.", "1999")
	success("She was born in 2015, just a few years ago.", "2015")
	success("The law was passed in 2005 and is now in effect.", "2005")

	fail("The company was established in 1989, before the tech boom.")
	fail("She was born in 2050, in a future world.")
	fail("His graduation was in 2090, which sounds impossible.")
	fail("The document mentioned a date from 1985, long before the project began.")
}
