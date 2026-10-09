# Contract: Postman dynamic variables (118)

Source of truth: Postman's docs page "Dynamic variables" and the generator table in
`postmanlabs/postman-collection` → `lib/superstring/dynamic-variables.js`. The implementation
vendors that file's name list as `test/integration/resources/postman-dynamic-variables.txt`, and a
parity test fails if any name is missing or extra.

Rules:
- Resolved **after** all user scopes; a user variable with the same name wins.
- Every occurrence gets a fresh value (two `{{$guid}}` in one request differ).
- Values are strings with Postman's shape; word content may differ (different word lists).
- Time-based values use the app clock (`timeutils.Now`), so BDD can freeze time.
- Generator: `gofakeit/v7` where an equivalent exists; small in-repo lists/templates otherwise
  (marked *custom*).

| Group | Names → format |
|---|---|
| Common | `$guid`, `$randomUUID` → UUID v4 · `$timestamp` → Unix seconds · `$isoTimestamp` → `YYYY-MM-DDTHH:mm:ss.SSSZ` (UTC) |
| Text, numbers, colours | `$randomAlphaNumeric` → 1 char [a-z0-9] · `$randomBoolean` → `true`/`false` · `$randomInt` → integer 0–1000 · `$randomColor` → colour name · `$randomHexColor` → `#rrggbb` · `$randomAbbreviation` → e.g. `SQL` |
| Internet | `$randomIP` → IPv4 · `$randomIPV6` → IPv6 · `$randomMACAddress` → `xx:xx:xx:xx:xx:xx` · `$randomPassword` → 15 alphanumerics · `$randomLocale` → ISO 639-1 code (*custom* list) · `$randomUserAgent` · `$randomProtocol` → `http`/`https` (*custom*) · `$randomSemver` → `x.y.z` |
| Names | `$randomFirstName` · `$randomLastName` · `$randomFullName` · `$randomNamePrefix` → e.g. `Dr.` · `$randomNameSuffix` → e.g. `MD` |
| Profession | `$randomJobArea` (*custom*) · `$randomJobDescriptor` · `$randomJobTitle` · `$randomJobType` |
| Phone, address, location | `$randomPhoneNumber` → `###-###-####` · `$randomPhoneNumberExt` → `##-###-###-####` · `$randomCity` · `$randomStreetName` · `$randomStreetAddress` → `#### Name Suffix` · `$randomCountry` · `$randomCountryCode` → ISO 3166 alpha-2 · `$randomLatitude` · `$randomLongitude` |
| Images (all *custom* URL templates) | `$randomAvatarImage` → avatar URL · `$randomImageUrl` → `http://lorempixel.com/640/480` · `$randomAbstractImage`, `$randomAnimalsImage`, `$randomBusinessImage`, `$randomCatsImage`, `$randomCityImage`, `$randomFoodImage`, `$randomNightlifeImage`, `$randomFashionImage`, `$randomPeopleImage`, `$randomNatureImage`, `$randomSportsImage`, `$randomTransportImage` → `http://lorempixel.com/640/480/<category>` · `$randomImageDataUri` → `data:image/svg+xml;charset=UTF-8,…` |
| Finance | `$randomBankAccount` → 8 digits · `$randomBankAccountName` (*custom*) · `$randomCreditCardMask` → 4 digits · `$randomBankAccountBic` (*custom*) · `$randomBankAccountIban` (*custom*, checksum-valid) · `$randomTransactionType` (*custom*: deposit, withdrawal, payment, invoice) · `$randomCurrencyCode` → ISO 4217 · `$randomCurrencyName` · `$randomCurrencySymbol` (*custom*) · `$randomBitcoin` → address |
| Business | `$randomCompanyName` · `$randomCompanySuffix` · `$randomBs` · `$randomBsAdjective` (*custom*) · `$randomBsBuzz` · `$randomBsNoun` (*custom*) |
| Catchphrases | `$randomCatchPhrase` · `$randomCatchPhraseAdjective`, `$randomCatchPhraseDescriptor`, `$randomCatchPhraseNoun` (*custom*) |
| Databases (*custom*) | `$randomDatabaseColumn` · `$randomDatabaseType` · `$randomDatabaseCollation` · `$randomDatabaseEngine` |
| Dates | `$randomDateFuture` · `$randomDatePast` · `$randomDateRecent` → JS `Date.toString()` format (`Tue Jun 09 2026 21:10:36 GMT+0000 (Coordinated Universal Time)`) · `$randomWeekday` · `$randomMonth` |
| Domains, emails, usernames | `$randomDomainName` · `$randomDomainSuffix` · `$randomDomainWord` (*custom*) · `$randomEmail` · `$randomExampleEmail` → `…@example.com\|org\|net` (*custom*) · `$randomUserName` · `$randomUrl` |
| Files and directories | `$randomFileName` · `$randomFileType` · `$randomFileExt` · `$randomCommonFileName` · `$randomCommonFileType` · `$randomCommonFileExt` · `$randomFilePath` · `$randomDirectoryPath` · `$randomMimeType` (names/types/paths *custom*) |
| Stores | `$randomPrice` → `0.00`–`1000.00` · `$randomProduct` · `$randomProductAdjective` (*custom*) · `$randomProductMaterial` · `$randomProductName` · `$randomDepartment` |
| Grammar | `$randomNoun` · `$randomVerb` · `$randomIngverb` · `$randomAdjective` · `$randomWord` · `$randomWords` (2–5 words) · `$randomPhrase` |
| Lorem ipsum | `$randomLoremWord` · `$randomLoremWords` (3) · `$randomLoremSentence` · `$randomLoremSentences` (2–6) · `$randomLoremParagraph` · `$randomLoremParagraphs` (3, `\n`-separated) · `$randomLoremText` · `$randomLoremSlug` (3 words, `-`) · `$randomLoremLines` (1–5, `\n`) |

Format checks used by SC-005 live in the BDD feature as one regex per name.
