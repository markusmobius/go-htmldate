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
