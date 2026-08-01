# /profile/user/wow/collections/transmogs response format

The map contains two top-level keys: "appearance_sets" and "slots". "appearance_sets" has appearance set IDs. "slots" has appearance IDs. Be sure to treat the two IDs separately!

```json
 {
   "appearance_sets": [
    {
      "key": {
        "href": "https:us.api.blizzard.com/data/wow/item-appearance/set/23?namespace=static-12.0.7_67808-us"
      },
      "name": "Wild Combatant's Leather Armor",
      "id": 23
    },
    ...
   ],
   "slots": [
     {
       "slot": {
            "type": "HEAD",
            "name": "Head"
       },
       "appearances": [
       {
         "key": {
            "href": "https:us.api.blizzard.com/data/wow/item-appearance/358?namespace=static-11.1.5_60179-us"
         },
         "id": 358
       },
       ...
     },
     {
       "slot": {
         "type": "PROFESSION_TOOL",
         "name": "Profession Tool"
       },
       "appearances": [...]
     },
     {
       "slot": {
         "type": "PROFESSION_GEAR",
         "name": "Profession Equipment"
       },
       "appearances": [...]
     }
   ]
 }
```
