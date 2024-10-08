# wow

Tools that use the World of Warcraft developer API. https://develop.battle.net/documentation/world-of-warcraft/game-data-apis

Scan the Auction House for arbitrage opportunities. Sometimes players put items up for auction at a purchase price lower than what the stores will pay to buy the item from you. Find these arbitrage opportunities and display what profit is to be made.

Also, scan the Auction House for good deals on items that my characters need. Sometimes players dump items for lower prices than my characters would have to pay at the stores. Or, sometimes hard to find items appear at good prices. Find these and display them.

The auction house downloadable data is updated once an hour. The precise time might depend upon when the service was last started up after a maintenance. Sampling multiple times during a one-hour window will result in identical downloads. There are other people playing this same arbitrage game, so you have to be *very* quick (right at 10 after) to get in on the bargains before they are gone.

## Development

https://develop.battle.net/documentation

### OAUTH 2.0

[Reference implementation](https://github.com/douglasmakey/oauth2-example).

### Callback URI

This callback has been registered with Blizzard for this client ID:

```text
redirect_uri = 'http://localhost:8000/auth/wow/callback'
```

## Oddities

Why does 'Strong Sniffin' Soup for Niffen' (204790) have a sell price, but is not sellable?
