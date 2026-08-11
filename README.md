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

# items

A command line tool for searching and modifying the WoW item persistence gob file.

# Development notes

The auction house downloadable data is updated once an hour. The precise time might depend upon when the service was last started up after a maintenance. Sampling multiple times during a one-hour window will result in identical downloads. There are other people playing this same arbitrage game, so you have to be *very* quick (right at 10 after) to get in on the bargains before they are gone.

### WoW web APIs

https://develop.battle.net/documentation
https://develop.battle.net/documentation/world-of-warcraft/game-data-apis

### WoW web API Credentials

The WoW web API credentials are stored in the macOS keychain. Update them using the 'items' app.

When the credentials are created (or updated) you should run the 'keychain Access' app and set the two credentials to be accessible by any application. Otherwise, there will be an extra pop-up requiring authorization to access the keys.

### OAuth 2.0

[Reference implementation](https://github.com/douglasmakey/oauth2-example).

### Callback URI

This callback has been registered with Blizzard for this client ID:

```text
redirect_uri = 'http://localhost:8888/auth/blizzard/profile'
```

### Areas for improvement

Add tests
