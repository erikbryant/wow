# wow

A command-line tool that analyzes the World of Warcraft Auction House to
* find arbitrage opportunities
* find items needed by your characters that are selling for cheap
* find battle pets needed
* find battle pets for resale
* generate files to configure the wowMerchant AddOn

### Find arbitrage opportunities

Scan the Auction House for arbitrage opportunities. Players auction items at a purchase price lower than vendors will pay. Find these arbitrage opportunities and display what profit is to be made.

### Find items needed
Scan the Auction House for items that my characters need that are selling at low prices.

### Find battle pets

Find battle pets that your characters do not own. If they are selling at a good price, suggest them.

Find battle pets that will resell well. Suggest buying those.

### Generate files for wowMerchant

The wowMerchant AddOn depends on certain data from the wow application. Scrape this data from the WoW web APIs and generate files consumable by wowMerchant.

This includes a cache of all current vendor prices. It also includes arbitrage items (selling at a discount to vendor prices).

# wowctl

A command line tool for searching and modifying the WoW item persistence gob file.

# secret

A command line tool for managing the WoW web API credentials.

# Maintenance Items

From time to time there will be maintenance tasks to complete. These generally result from Blizzard adding new items. The code is written with that in mind. Fail early, fail loudly. WoW's underlying data is not static. When Blizzard changes an API response or introduces new data that the application does not understand, we prefer an obvious failure or maintenance diagnostic over silently producing an incorrect result.

### iLevels

Profession tools have iLevels. Looking up a given profession tool (by itemID) in the auction house is not sufficient. You also have to include the iLevel you are looking for. We maintain a list of known profession tools and the known iLevels they can express. If an unknown profession tool is encountered that is not in this table a message will be emitted with the entry to add and which source file to add it. This will just be an empty entry. You will need to scan several auction houses to find items listed for sale that enumerate possible iLevels.

### Blizzard Items API endpoint 404

Blizzard provides a web API to retrieve item data. Not all item IDs are available through this API. Some valid item IDs will return a 404. If this happens the app will emit a message that the item ID was not found. Add a synthetic item using the wowctl tool. You can figure out what values to enter for the synthetic item by Googling for 'wow item id nnnnn'.

If you create new synthetic items (or change existing ones) be sure to run '/merch validate' in the WoW client. This will ensure that the price you entered for the item is the same as the price the client knows.

### Stale item data

When you use the '/merch scan' command in the wowMerchant addon (or the '/merch validate' command) the addon will validate that the price cache reflects values seen in the live system. Sometimes the item persistence is stale. In those cases, use wowctl to refresh those item IDs.

# Development notes

The auction house downloadable data is updated once an hour. The precise time might depend upon when the service was last started up after a maintenance. Sampling multiple times during a one-hour window will result in identical downloads. There are other people playing this same arbitrage game, so you have to be *very* quick to get in on the bargains before they are gone.

### WoW web APIs

https://develop.battle.net/documentation
https://develop.battle.net/documentation/world-of-warcraft/game-data-apis

### WoW web API Credentials

The WoW web API credentials are stored in the macOS keychain. Update them using the 'secret' app. The 'wow' app expects the credentials to be named `clientID` and `clientSecret`.

The credentials are needed to authenticate against the WoW web API. The authentication protocol is Oauth2. See this [Reference implementation](https://github.com/douglasmakey/oauth2-example).

This callback has been registered with Blizzard for this client ID:

```text
redirect_uri = 'http://localhost:8888/auth/blizzard/profile'
```

### Attributions

me: I wrote the original code.

douglasmakey: Provided an Oauth2 reference implementation.

Keybase: Wrote the original keystore implementation. I forked it and removed any code I did not use. Their original implementation is completely fine. I only forked it out of paranoia over supply-chain attacks.

ChatGPT: Cleaned up the code, made it idiomatic, wrote tests, suggested additional features, and contributed in countless other ways.

### Areas for improvement

The existing tests are old. The code has changed a lot since they were written. Evaluate them for usefulness, correctness, and coverage.

Test coverage is low. Add more tests

Measure test coverage?