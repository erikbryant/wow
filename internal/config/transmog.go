package config

// FlakyAppearanceIDs WoW says I own them, but this app thinks I don't
var FlakyAppearanceIDs = map[int64]struct{}{
	// These are not real appearances; they generate false positives
	573:   {}, // Various equippable profession items
	577:   {}, // Various equippable profession items
	870:   {}, // Ammo
	1884:  {}, // Various fish held in offhand
	2016:  {}, // Various fish held in offhand
	2019:  {}, // Various fish held in offhand
	70361: {}, // Elegant Artisan's Cooking Hat
	78217: {}, // Elegant Artisan's Fishing Hat

	// NOT part of an appearance set (so less interesting)
	1172:  {}, // Ghostly Bracers
	22334: {}, // Shrediron's Shredder
	22335: {}, // Shrediron's Shredder
	22392: {}, // Shadowtome
	22547: {}, // Shadowtome
	22750: {}, // {}steel Waistguard
	22902: {}, // Hexweave Cowl
	22905: {}, // Hexweave Mantle
	22911: {}, // Hexweave Cowl
	22914: {}, // Hexweave Mantle
	22939: {}, // Steelforged Saber
	22940: {}, // Steelforged Saber
	23254: {}, // {}steel Waistguard
	56701: {}, // Choral Hood
	56702: {}, // Choral Amice
	56703: {}, // Choral Vestments
	56704: {}, // Choral Sash
	56705: {}, // Choral Leggings
	56706: {}, // Choral Slippers
	56707: {}, // Choral Wraps
	56708: {}, // Choral Handwraps
	56859: {}, // Staccato Helm
	56860: {}, // Staccato Mantle
	56861: {}, // Staccato Vest
	56862: {}, // Staccato Belt
	56863: {}, // Staccato Leggings
	56864: {}, // Staccato Boots
	56865: {}, // Staccato Cuffs
	56866: {}, // Staccato Grips
	57167: {}, // Harmonium Helm
	57168: {}, // Harmonium Spaulders
	57169: {}, // Harmonium Breastplate
	57170: {}, // Harmonium Girdle
	57171: {}, // Harmonium Legplates
	57172: {}, // Harmonium Percussive Stompers
	57173: {}, // Harmonium Vambrace
	57174: {}, // Harmonium Gauntlets
	57175: {}, // Antecedent Drape
	57224: {}, // Anthemic Coif
	57225: {}, // Anthemic Shoulders
	57226: {}, // Anthemic Cuirass
	57227: {}, // Anthemic Links
	57228: {}, // Anthemic Legguards
	57230: {}, // Anthemic Bracers
	57231: {}, // Anthemic Gauntlets
	78230: {}, // Scepter of Spectacle: Order

	// NOT part of an appearance set  (so less interesting) [Horde only]
	37116: {}, // Enchanter's Sorcerous Scepter

	// Part of an appearance set (so of great interest), but rarely available
	//18561: {}, // Fists of Lightning
	//18575: {}, // Nightfire Robe
	//18715: {}, // Greyshadow Gloves
	//22757: {}, // {}steel Armguards
	//22906: {}, // Hexweave Bracers
	//22915: {}, // Hexweave Bracers
	//23247: {}, // {}steel Armguards
	//24178: {}, // {Brilliant, Nimble, Powerful} Burnished Cloak
	//24180: {}, // {Brilliant, Nimble, Powerful} Burnished Cloak
	32066: {}, // Fashionable Autumn Cloak
	32237: {}, // Aristocrat's Winter Drape
	//33357: {}, // Sash of the Unredeemed
	//33365: {}, // Sash of the Unredeemed
	33423: {}, // Treads of Panicked Escape
	33439: {}, // Treads of Panicked Escape
	33496: {}, // Cord of Pilfered Rosaries
	33497: {}, // Treads of Violent Intrusion
	//33716: {}, // Moon-Wrought Clasp
	//34314: {}, // Pristine Moon-Wrought Clasp
	//34558: {}, // Cuffs of the Viridian Flameweavers
	//34870: {}, // Gloves of Abhorrent Strategies
	//34886: {}, // Gloves of Abhorrent Strategies
	35092: {}, // Wristguards of Ominous Forging
	35101: {}, // Wristguards of Ominous Forging
	//38275: {}, // Reinforced Test Subject Shackles
	//38291: {}, // Reinforced Test Subject Shackles
	38359: {}, // Bloody Experimenter's Wraps
	38409: {}, // Crushproof Vambraces
	38830: {}, // Cord of Zandalari Resolve
	39969: {}, // Gauntlets of Crashing Tides
	39976: {}, // Gauntlets of Crashing Tides
	39980: {}, // Brutish Myrmidon's Vambraces
	39983: {}, // Brutish Myrmidon's Vambraces
	39987: {}, // Gauntlets of Crashing Tides
	40325: {}, // Cloak of Blessed Depths
	40811: {}, // Belt of Concealed Intent
	40813: {}, // Belt of Concealed Intent
	40883: {}, // Footpads of Terrible Delusions
	40886: {}, // Footpads of Terrible Delusions
	40967: {}, // Gauntlets of Nightmare Manifest
	40970: {}, // Gauntlets of Nightmare Manifest
	44135: {}, // Bindings of the Subjugated
	44178: {}, // Bindings of the Subjugated
	//57228: {}, // Anthemic Legguards
	//80187: {}, // Skyless Coif
	//80188: {}, // Skyless Epaulets

}

func FlakyAppearanceID(id int64) bool {
	_, ok := FlakyAppearanceIDs[id]
	return ok
}
