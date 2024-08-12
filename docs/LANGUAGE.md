# Language

| Phrase     | Meaning                                                                                     |
|------------|---------------------------------------------------------------------------------------------|
| Clan       | The unit with a 4-character ID staring with 0. The Clan is the parent of other Tribes.      |
| Element    | Any unit that has a 6-character ID. Any unit that has a Tribe for a parent.                 |
| Hex Report | A section of the movement results line. These sections are delimited by backslashes.        |
| MRL        | Movement Results Line. A single line from the turn report detailing the results of a move.  |
| Tribe      | Any unit with a 4-character ID. A Clan is a Tribe.                                          | 
| Unit       | Anything that can be issued a move order: Tribe, Courier, Element, Fleet, Garrison.         |

## Movement Results Lines

The Movement Results Lines are the lines that contain the results of a move.
They started with a prefix and are composed of Hex Reports.

## Hex Reports

Hex Reports are sections of the MRL.
They are delimited by backslashes and contain detailed information on what is in the hex.

1. Terrain type
2. Neighboring Hexes (oceans, lakes, mountains, etc.)
3. Edge features (rivers, passes, fords, etc.)
4. Hex names (settlements, villages, etc.)
5. Resources (coal, iron ore, etc.)
6. Random Encounters (horses, saddlebags, etc.)
7. Units in the hex

## Nits

There are five lines in the turn report that are used as Movement Results Lines.
These are:

1. Tribe Follows
2. Tribe Movement
3. Fleet Movement
4. Scout
5. Status

Status is not really a movement, but it turns out that it contains Hex Reports that can be parsed like tribe movements.

## Sources

`0141 Fred` mentioned "Hex Report" on the TribeNet Discord.
I thought that was a great name for the section.