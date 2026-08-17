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