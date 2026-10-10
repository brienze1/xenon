package webservice

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/brianvoe/gofakeit/v7"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/webservice/dynamicdata"
	"github.com/brienze1/xenon/pkg/logger"
	"github.com/brienze1/xenon/pkg/timeutils"
)

const (
	alphaNumericLower = "abcdefghijklmnopqrstuvwxyz0123456789"
	alphaNumeric      = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	base58Alphabet    = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	upperLetters      = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	javaScriptDate    = "Mon Jan 02 2006 15:04:05 GMT-0700"
	javaScriptZone    = " (Coordinated Universal Time)"
	isoMilliseconds   = "2006-01-02T15:04:05.000Z"
	imageBaseUrl      = "http://lorempixel.com/640/480"
	avatarUrl         = "https://cloudflare-ipfs.com/ipfs/Qmd3W5DuhgHirLHGVixi6V76LhCkZUz6pnFt5AJBiyvHye/avatar/%d.jpg"
	imageDataSvg      = `<svg xmlns="http://www.w3.org/2000/svg" version="1.1" baseProfile="full" width="640" height="480"><rect width="100%%" height="100%%" fill="%s"/><text x="0" y="20" font-size="20" text-anchor="start" fill="white">640x480</text></svg>`
	ibanCheckModulus  = 97
)

type dynamicGenerator struct {
	generate func(faker *gofakeit.Faker) string
	name     string
}

type dynamicValuesWebservice struct {
	faker      *gofakeit.Faker
	generators map[string]func(faker *gofakeit.Faker) string
	names      []string
}

func NewDynamicValuesWebservice(source adapter.RandomSource) adapter.DynamicValuesWebservice {
	generators := dynamicGenerators()
	webservice := &dynamicValuesWebservice{
		faker:      gofakeit.NewFaker(source, true),
		generators: make(map[string]func(faker *gofakeit.Faker) string, len(generators)),
		names:      make([]string, 0, len(generators)),
	}
	for _, generator := range generators {
		webservice.generators[generator.name] = generator.generate
		webservice.names = append(webservice.names, generator.name)
	}
	return webservice
}

func (d *dynamicValuesWebservice) Generate(ctx context.Context, name string) (string, bool) {
	logger.Debug(ctx, "Started", map[string]any{"name": name})

	generate, found := d.generators[name]
	if !found {
		logger.Debug(ctx, "Finished", map[string]any{"name": name, "found": false})
		return "", false
	}
	value := generate(d.faker)

	logger.Debug(ctx, "Finished", map[string]any{"name": name, "found": true})
	return value, true
}

func (d *dynamicValuesWebservice) Names(ctx context.Context) []string {
	logger.Debug(ctx, "Started")

	names := append([]string{}, d.names...)

	logger.Debug(ctx, "Finished", map[string]any{"names": len(names)})
	return names
}

func dynamicGenerators() []dynamicGenerator {
	return []dynamicGenerator{
		{name: "$guid", generate: uuid},
		{name: "$timestamp", generate: func(*gofakeit.Faker) string { return strconv.FormatInt(timeutils.Now().Unix(), 10) }},
		{name: "$isoTimestamp", generate: func(*gofakeit.Faker) string { return timeutils.Now().UTC().Format(isoMilliseconds) }},
		{name: "$randomInt", generate: func(f *gofakeit.Faker) string { return strconv.Itoa(f.Number(0, 1000)) }},
		{name: "$randomPhoneNumber", generate: func(f *gofakeit.Faker) string {
			return fmt.Sprintf("%03d-%03d-%04d", f.Number(0, 999), f.Number(0, 999), f.Number(0, 9999))
		}},
		{name: "$randomPhoneNumberExt", generate: func(f *gofakeit.Faker) string {
			return fmt.Sprintf("%02d-%03d-%03d-%04d", f.Number(0, 99), f.Number(0, 999), f.Number(0, 999), f.Number(0, 9999))
		}},
		{name: "$randomLocale", generate: pick(dynamicdata.Locales)},
		{name: "$randomWords", generate: func(f *gofakeit.Faker) string { return words(f, f.Number(2, 5), " ") }},
		{name: "$randomFilePath", generate: func(f *gofakeit.Faker) string { return directoryPath(f) + "/" + fileName(f) }},
		{name: "$randomDirectoryPath", generate: directoryPath},
		{name: "$randomCity", generate: func(f *gofakeit.Faker) string { return f.City() }},
		{name: "$randomStreetName", generate: streetName},
		{name: "$randomStreetAddress", generate: func(f *gofakeit.Faker) string { return strconv.Itoa(f.Number(100, 99999)) + " " + streetName(f) }},
		{name: "$randomCountry", generate: func(f *gofakeit.Faker) string { return f.Country() }},
		{name: "$randomCountryCode", generate: func(f *gofakeit.Faker) string { return f.CountryAbr() }},
		{name: "$randomLatitude", generate: func(f *gofakeit.Faker) string { return fmt.Sprintf("%.4f", f.Float64Range(-90, 90)) }},
		{name: "$randomLongitude", generate: func(f *gofakeit.Faker) string { return fmt.Sprintf("%.4f", f.Float64Range(-180, 180)) }},
		{name: "$randomColor", generate: func(f *gofakeit.Faker) string { return f.SafeColor() }},
		{name: "$randomDepartment", generate: pick(dynamicdata.Departments)},
		{name: "$randomProductName", generate: func(f *gofakeit.Faker) string {
			return join(f, " ", dynamicdata.ProductAdjectives, dynamicdata.ProductMaterials, dynamicdata.Products)
		}},
		{name: "$randomProductAdjective", generate: pick(dynamicdata.ProductAdjectives)},
		{name: "$randomProductMaterial", generate: pick(dynamicdata.ProductMaterials)},
		{name: "$randomProduct", generate: pick(dynamicdata.Products)},
		{name: "$randomCompanyName", generate: func(f *gofakeit.Faker) string { return f.LastName() + " " + f.CompanySuffix() }},
		{name: "$randomCompanySuffix", generate: func(f *gofakeit.Faker) string { return f.CompanySuffix() }},
		{name: "$randomCatchPhrase", generate: func(f *gofakeit.Faker) string {
			return join(f, " ", dynamicdata.CatchPhraseAdjectives, dynamicdata.CatchPhraseDescriptors, dynamicdata.CatchPhraseNouns)
		}},
		{name: "$randomBs", generate: func(f *gofakeit.Faker) string {
			return join(f, " ", dynamicdata.BsBuzzwords, dynamicdata.BsAdjectives, dynamicdata.BsNouns)
		}},
		{name: "$randomCatchPhraseAdjective", generate: pick(dynamicdata.CatchPhraseAdjectives)},
		{name: "$randomCatchPhraseDescriptor", generate: pick(dynamicdata.CatchPhraseDescriptors)},
		{name: "$randomCatchPhraseNoun", generate: pick(dynamicdata.CatchPhraseNouns)},
		{name: "$randomBsAdjective", generate: pick(dynamicdata.BsAdjectives)},
		{name: "$randomBsBuzz", generate: pick(dynamicdata.BsBuzzwords)},
		{name: "$randomBsNoun", generate: pick(dynamicdata.BsNouns)},
		{name: "$randomDatabaseColumn", generate: pick(dynamicdata.DatabaseColumns)},
		{name: "$randomDatabaseType", generate: pick(dynamicdata.DatabaseTypes)},
		{name: "$randomDatabaseCollation", generate: pick(dynamicdata.DatabaseCollations)},
		{name: "$randomDatabaseEngine", generate: pick(dynamicdata.DatabaseEngines)},
		{name: "$randomDatePast", generate: func(f *gofakeit.Faker) string {
			now := timeutils.Now()
			return javaScriptDateString(f.DateRange(now.AddDate(-1, 0, 0), now))
		}},
		{name: "$randomDateFuture", generate: func(f *gofakeit.Faker) string {
			now := timeutils.Now()
			return javaScriptDateString(f.DateRange(now, now.AddDate(1, 0, 0)))
		}},
		{name: "$randomDateRecent", generate: func(f *gofakeit.Faker) string {
			now := timeutils.Now()
			return javaScriptDateString(f.DateRange(now.AddDate(0, 0, -1), now))
		}},
		{name: "$randomMonth", generate: func(f *gofakeit.Faker) string { return time.Month(f.Number(1, 12)).String() }},
		{name: "$randomWeekday", generate: func(f *gofakeit.Faker) string { return time.Weekday(f.Number(0, 6)).String() }},
		{name: "$randomBankAccount", generate: func(f *gofakeit.Faker) string { return fmt.Sprintf("%08d", f.Number(0, 99999999)) }},
		{name: "$randomBankAccountName", generate: pick(dynamicdata.BankAccountNames)},
		{name: "$randomCreditCardMask", generate: func(f *gofakeit.Faker) string { return fmt.Sprintf("%04d", f.Number(0, 9999)) }},
		{name: "$randomPrice", generate: func(f *gofakeit.Faker) string { return fmt.Sprintf("%.2f", f.Float64Range(0, 1000)) }},
		{name: "$randomTransactionType", generate: pick(dynamicdata.TransactionTypes)},
		{name: "$randomCurrencyCode", generate: func(f *gofakeit.Faker) string { return f.CurrencyShort() }},
		{name: "$randomCurrencyName", generate: func(f *gofakeit.Faker) string { return f.CurrencyLong() }},
		{name: "$randomCurrencySymbol", generate: pick(dynamicdata.CurrencySymbols)},
		{name: "$randomBitcoin", generate: func(f *gofakeit.Faker) string {
			return f.RandomString([]string{"1", "3"}) + characters(f, base58Alphabet, f.Number(25, 33))
		}},
		{name: "$randomBankAccountIban", generate: iban},
		{name: "$randomBankAccountBic", generate: func(f *gofakeit.Faker) string {
			return characters(f, upperLetters, 4) + f.CountryAbr() + characters(f, upperLetters, 2)
		}},
		{name: "$randomAbbreviation", generate: pick(dynamicdata.Abbreviations)},
		{name: "$randomAdjective", generate: func(f *gofakeit.Faker) string { return f.AdjectiveDescriptive() }},
		{name: "$randomNoun", generate: func(f *gofakeit.Faker) string { return f.NounCommon() }},
		{name: "$randomVerb", generate: func(f *gofakeit.Faker) string { return f.VerbAction() }},
		{name: "$randomIngverb", generate: pick(dynamicdata.Ingverbs)},
		{name: "$randomPhrase", generate: func(f *gofakeit.Faker) string { return f.HackerPhrase() }},
		{name: "$randomAvatarImage", generate: func(f *gofakeit.Faker) string { return fmt.Sprintf(avatarUrl, f.Number(0, 1249)) }},
		{name: "$randomImageUrl", generate: func(*gofakeit.Faker) string { return imageBaseUrl }},
		{name: "$randomAbstractImage", generate: image("abstract")},
		{name: "$randomAnimalsImage", generate: image("animals")},
		{name: "$randomBusinessImage", generate: image("business")},
		{name: "$randomCatsImage", generate: image("cats")},
		{name: "$randomCityImage", generate: image("city")},
		{name: "$randomFoodImage", generate: image("food")},
		{name: "$randomNightlifeImage", generate: image("nightlife")},
		{name: "$randomFashionImage", generate: image("fashion")},
		{name: "$randomPeopleImage", generate: image("people")},
		{name: "$randomNatureImage", generate: image("nature")},
		{name: "$randomSportsImage", generate: image("sports")},
		{name: "$randomTransportImage", generate: image("transport")},
		{name: "$randomImageDataUri", generate: func(f *gofakeit.Faker) string {
			return "data:image/svg+xml;charset=UTF-8," + url.PathEscape(fmt.Sprintf(imageDataSvg, f.SafeColor()))
		}},
		{name: "$randomEmail", generate: func(f *gofakeit.Faker) string { return f.Email() }},
		{name: "$randomExampleEmail", generate: func(f *gofakeit.Faker) string {
			return userName(f) + "@example." + f.RandomString([]string{"com", "org", "net"})
		}},
		{name: "$randomUserName", generate: userName},
		{name: "$randomProtocol", generate: pick(dynamicdata.Protocols)},
		{name: "$randomUrl", generate: func(f *gofakeit.Faker) string { return f.RandomString(dynamicdata.Protocols) + "://" + f.DomainName() }},
		{name: "$randomDomainName", generate: func(f *gofakeit.Faker) string { return f.DomainName() }},
		{name: "$randomDomainSuffix", generate: func(f *gofakeit.Faker) string { return f.DomainSuffix() }},
		{name: "$randomDomainWord", generate: func(f *gofakeit.Faker) string { return f.AdjectiveDescriptive() + "-" + f.NounCommon() }},
		{name: "$randomIP", generate: func(f *gofakeit.Faker) string {
			return fmt.Sprintf("%d.%d.%d.%d", f.Number(0, 255), f.Number(0, 255), f.Number(0, 255), f.Number(0, 255))
		}},
		{name: "$randomIPV6", generate: func(f *gofakeit.Faker) string { return hexGroups(f, "%04x", 0xffff, 8) }},
		{name: "$randomUserAgent", generate: func(f *gofakeit.Faker) string { return f.UserAgent() }},
		{name: "$randomHexColor", generate: func(f *gofakeit.Faker) string { return fmt.Sprintf("#%06x", f.Number(0, 0xffffff)) }},
		{name: "$randomMACAddress", generate: func(f *gofakeit.Faker) string { return hexGroups(f, "%02x", 0xff, 6) }},
		{name: "$randomPassword", generate: func(f *gofakeit.Faker) string { return characters(f, alphaNumeric, 15) }},
		{name: "$randomLoremWord", generate: func(f *gofakeit.Faker) string { return f.LoremIpsumWord() }},
		{name: "$randomLoremWords", generate: func(f *gofakeit.Faker) string { return loremWords(f, 3, " ") }},
		{name: "$randomLoremSentence", generate: loremSentence},
		{name: "$randomLoremSlug", generate: func(f *gofakeit.Faker) string { return loremWords(f, 3, "-") }},
		{name: "$randomLoremSentences", generate: func(f *gofakeit.Faker) string { return loremSentences(f, f.Number(2, 6), " ") }},
		{name: "$randomLoremParagraph", generate: loremParagraph},
		{name: "$randomLoremParagraphs", generate: func(f *gofakeit.Faker) string {
			return strings.Join([]string{loremParagraph(f), loremParagraph(f), loremParagraph(f)}, "\n")
		}},
		{name: "$randomLoremText", generate: loremParagraph},
		{name: "$randomLoremLines", generate: func(f *gofakeit.Faker) string { return loremSentences(f, f.Number(1, 5), "\n") }},
		{name: "$randomFirstName", generate: func(f *gofakeit.Faker) string { return f.FirstName() }},
		{name: "$randomLastName", generate: func(f *gofakeit.Faker) string { return f.LastName() }},
		{name: "$randomFullName", generate: func(f *gofakeit.Faker) string { return f.FirstName() + " " + f.LastName() }},
		{name: "$randomJobTitle", generate: func(f *gofakeit.Faker) string {
			return join(f, " ", dynamicdata.JobDescriptors, dynamicdata.JobAreas, dynamicdata.JobTypes)
		}},
		{name: "$randomNamePrefix", generate: pick(dynamicdata.NamePrefixes)},
		{name: "$randomNameSuffix", generate: pick(dynamicdata.NameSuffixes)},
		{name: "$randomJobDescriptor", generate: pick(dynamicdata.JobDescriptors)},
		{name: "$randomJobArea", generate: pick(dynamicdata.JobAreas)},
		{name: "$randomJobType", generate: pick(dynamicdata.JobTypes)},
		{name: "$randomUUID", generate: uuid},
		{name: "$randomBoolean", generate: func(f *gofakeit.Faker) string { return strconv.FormatBool(f.Bool()) }},
		{name: "$randomWord", generate: func(f *gofakeit.Faker) string { return f.NounCommon() }},
		{name: "$randomAlphaNumeric", generate: func(f *gofakeit.Faker) string { return characters(f, alphaNumericLower, 1) }},
		{name: "$randomFileName", generate: fileName},
		{name: "$randomCommonFileName", generate: func(f *gofakeit.Faker) string {
			return words(f, f.Number(1, 3), "_") + "." + f.RandomString(dynamicdata.CommonFileExtensions)
		}},
		{name: "$randomMimeType", generate: pick(dynamicdata.MimeTypes)},
		{name: "$randomCommonFileType", generate: pick(dynamicdata.CommonFileTypes)},
		{name: "$randomCommonFileExt", generate: pick(dynamicdata.CommonFileExtensions)},
		{name: "$randomFileType", generate: pick(dynamicdata.FileTypes)},
		{name: "$randomFileExt", generate: func(f *gofakeit.Faker) string { return f.FileExtension() }},
		{name: "$randomSemver", generate: func(f *gofakeit.Faker) string {
			return fmt.Sprintf("%d.%d.%d", f.Number(0, 9), f.Number(0, 9), f.Number(0, 9))
		}},
	}
}

func uuid(f *gofakeit.Faker) string {
	return f.UUID()
}

func pick(values []string) func(f *gofakeit.Faker) string {
	return func(f *gofakeit.Faker) string {
		return f.RandomString(values)
	}
}

func image(category string) func(f *gofakeit.Faker) string {
	return func(*gofakeit.Faker) string {
		return imageBaseUrl + "/" + category
	}
}

func join(f *gofakeit.Faker, separator string, lists ...[]string) string {
	parts := make([]string, 0, len(lists))
	for _, list := range lists {
		parts = append(parts, f.RandomString(list))
	}
	return strings.Join(parts, separator)
}

func words(f *gofakeit.Faker, count int, separator string) string {
	parts := make([]string, 0, count)
	for range count {
		parts = append(parts, f.NounCommon())
	}
	return strings.Join(parts, separator)
}

func loremWords(f *gofakeit.Faker, count int, separator string) string {
	parts := make([]string, 0, count)
	for range count {
		parts = append(parts, f.LoremIpsumWord())
	}
	return strings.Join(parts, separator)
}

func loremSentence(f *gofakeit.Faker) string {
	sentence := loremWords(f, f.Number(3, 10), " ")
	return strings.ToUpper(sentence[:1]) + sentence[1:] + "."
}

func loremSentences(f *gofakeit.Faker, count int, separator string) string {
	sentences := make([]string, 0, count)
	for range count {
		sentences = append(sentences, loremSentence(f))
	}
	return strings.Join(sentences, separator)
}

func loremParagraph(f *gofakeit.Faker) string {
	return loremSentences(f, f.Number(3, 6), " ")
}

func characters(f *gofakeit.Faker, alphabet string, count int) string {
	var builder strings.Builder
	for range count {
		builder.WriteByte(alphabet[f.IntN(len(alphabet))])
	}
	return builder.String()
}

func hexGroups(f *gofakeit.Faker, format string, max, count int) string {
	groups := make([]string, 0, count)
	for range count {
		groups = append(groups, fmt.Sprintf(format, f.Number(0, max)))
	}
	return strings.Join(groups, ":")
}

func directoryPath(f *gofakeit.Faker) string {
	return "/" + strings.Join([]string{f.RandomString(dynamicdata.Directories), f.NounCommon()}, "/")
}

func fileName(f *gofakeit.Faker) string {
	return words(f, f.Number(1, 3), "_") + "." + f.FileExtension()
}

func userName(f *gofakeit.Faker) string {
	return f.AdjectiveDescriptive() + f.RandomString([]string{".", "_"}) + f.NounCommon() + strconv.Itoa(f.Number(0, 99))
}

func streetName(f *gofakeit.Faker) string {
	return f.LastName() + " " + f.RandomString(dynamicdata.StreetSuffixes)
}

func javaScriptDateString(date time.Time) string {
	return date.UTC().Format(javaScriptDate) + javaScriptZone
}

func iban(f *gofakeit.Faker) string {
	format := dynamicdata.IbanFormats[f.IntN(len(dynamicdata.IbanFormats))]
	bban := characters(f, "0123456789", format.BbanDigits)
	return format.Country + fmt.Sprintf("%02d", ibanCheckDigits(format.Country, bban)) + bban
}

// ibanCheckDigits follows ISO 13616: move the country and "00" behind the BBAN, letters as 10-35, then 98 - mod 97.
func ibanCheckDigits(country, bban string) int {
	remainder := 0
	for _, character := range bban + country + "00" {
		digits := string(character)
		if unicode.IsLetter(character) {
			digits = strconv.Itoa(int(character-'A') + 10)
		}
		for _, digit := range digits {
			remainder = (remainder*10 + int(digit-'0')) % ibanCheckModulus
		}
	}
	return ibanCheckModulus + 1 - remainder
}
