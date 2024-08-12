# OttoMap

OttoMap is a tool that translates TribeNet turn report files into Worldographer maps.

## Overview
I'm planning on translating a small subset of the turn report.
See the files in the `domain` directory to get an idea of what we're looking at.

I think that will be enough data to feed the map generator.
Let me know if you think that there's something missing.

See the `OTTOMAP.md` file for an overview of running from command line and `BUILDING.md` for instructions on building the project.

> ISSUES: Please report any issues on the TribeNet Discord server.

### Status
OttoMap is in early development.
The turn report parser is nearly complete for land based movement.
Fleet movement has not been implemented (it's waiting on actual examples from turn reports)

The command line interface seems to be working but the documentation is incomplete.

The web interface is not yet implemented.
The single user server has been started but requires changes to the turn report parser.
I don't want to break the CLI, so this is proceeding slowly at best.

## Input Data
OttoMap expects all turn reports to be in text files in a single directory.

OttoMap loads all files that match the pattern "YEAR-MONTH.CLAN_ID.report.txt."
YEAR and MONTH are the three-digit year and two-digit month from the "Current Turn" line of the report.
CLAN_ID is the four-digit identifier for the clan (it must include the leading zero).

```bash
$ ls -1 input/*.txt

input/899-12.0138.report.txt
input/900-01.0138.report.txt
input/900-02.0138.report.txt
input/900-03.0138.report.txt
input/900-04.0138.report.txt
input/900-05.0138.report.txt
```

The files are created by opening the turn report (the `.DOCX` file),
selecting all the text, and pasting it into a plain text file.

```bash
$ ls -1 input/*.docx
input/899-12.0138.Turn-Report.docx
input/900-01.0138.Turn-Report.docx
input/900-02.0138.Turn-Report.docx
input/900-03.0138.Turn-Report.docx
input/900-04.0138.Turn-Report.docx
input/900-05.0138.Turn-Report.docx

$ file input/*

input/899-12.0138.Turn-Report.docx: Microsoft Word 2007+
input/899-12.0138.report.txt:        Unicode text, UTF-8 text
input/900-01.0138.Turn-Report.docx: Microsoft Word 2007+
input/900-01.0138.report.txt:        Unicode text, UTF-8 text
input/900-02.0138.Turn-Report.docx: Microsoft Word 2007+
input/900-02.0138.report.txt:        Unicode text, UTF-8 text
input/900-03.0138.Turn-Report.docx: Microsoft Word 2007+
input/900-03.0138.report.txt:        Unicode text, UTF-8 text
input/900-04.0138.Turn-Report.docx: Microsoft Word 2007+
input/900-04.0138.report.txt:        Unicode text, UTF-8 text
input/900-05.0138.Turn-Report.docx: Microsoft Word 2007+
input/900-05.0138.report.txt:        Unicode text, UTF-8 text
```

Spaces, line breaks, page breaks, and section breaks are important to the parser.
Please try to avoid altering them.

> WARNING: Please don't make changes your original turn report (the `.DOCX` file).
> You may not be able to revert back to the original file.

## License
OttoMap is licensed under version 3 of the GNU Affero General Public License.

OttoMap is built using packages that have different licenses.
All of these packages must be used in accordance with their original licenses;
including them in this application does not change their license terms to the AGPLv3.

Please see the individual packages for their license terms.

## Grids and the Big Map
The big map is divided into 676 grids arranged in 26 columns and 26 rows.
The grids use letters, not digits, for their coordinates on the big map.
The grid at the top left is (A, A), top right is (A, Z), bottom left is (Z, A), and bottom right is (Z, Z).

The ID for a grid is row then column.
The ID for the grid at top left is AA, top right is AZ, bottom left is ZA, and bottom right is ZZ.

Each grid has 30 columns and 21 rows.
The hex at the top left is (1, 1), top right is (1, 30), bottom left is (21, 1), and bottom right is (21, 30).

The ID for a hex is two-digit column then two-digit row.
The ID for the hex at the top left is 0101, top right is 3001, bottom left is 0121, and bottom right is 3021.

Hexes are "flat" on the top and even rows are shifted down.
For example, hex (13, 10) has

1. (13, 9) to the north
2. (14, 9) to the north-east
3. (14, 10) to the south-east
4. (13, 11) to the south
5. (12, 10) to the south-west
6. (12, 9) to the north-west

In turn reports, a hex in the grid is usually displayed as "AA 1310."

You can convert grid coordinates to big map coordinates.
A coordinate like "VN 0810" is "(N8, V10)."
That's row N, column V on the big map, then column 8, row 10 on the grid.

You can convert also convert grid coordinates to absolute coordinates by scaling the column and row values.
For our coordinate of "VN 0810," "N" is the 14th grid from the left and "V" is the 22nd grid from the top.
This gives us a column of (14 - 1) * 30 + (8 - 1) = 397 and row of (22 - 1) * 21 + (10 - 1) = 450, or (397, 450)
(We subtract one before multiplying because absolute coordinates start at zero, not one.)

## Parsing Errors
The report processor has been updated to fail on unexpected input.
I know this is annoying, but it prevents bad data from going into the map.

There are two causes for this: typos and new scenarios.

### Fixing typos in the input
Typos don't happen often, but when they do, you need to fix them and restart.
If you don't understand what needs to be fixed, please ask for help on the TribeNet Discord's `#mapping` channel.

### New scenarios
This is more common than typos since TribeNet supports so many actions.

You'll usually find a new scenario in the Scouting results.
I need to update the code and the test suites, so please ask for help on the TribeNet Discords `#mapping-tool` channel.

Adding code can take a while.
In the meantime, the only work-around is to delete the new scenario from the input and restart.
Results for that unit are going to be "off" until the code is fixed, but you'll be able to map out the rest of the turn.

## Big Map notes

![big_map_crossing](docs/big_map_crossing.png)

## Hex Movement
The map has flat hexagons with odd columns shoved down.
(I know, 0201 is SE of 0101, but 0101 is really (0, 0).)

```go
// hexDirectionVectors defines the vectors used to determine the coordinates
// of the neighboring column based on the direction and the odd/even column
// property of the starting hex.
//
// NB: grids start at 0101 and hexes at (0,0), so "odd" and "even" are based
//     on the hex coordinates, not the grid.
var hexDirectionVectors map[string]map[string][2]int
```

![grid1206](docs/grid1206.png)

```go
hexDirectionVectors["odd-column"]= {
    "N" : {+0, -1}, // ## 1206 -> (11, 05) -> (11, 04) -> ## 1205
    "NE": {+1, +0}, // ## 1206 -> (11, 05) -> (12, 05) -> ## 1306
    "SE": {+1, +1}, // ## 1206 -> (11, 05) -> (12, 06) -> ## 1307
    "S" : {+0, +1}, // ## 1206 -> (11, 05) -> (11, 06) -> ## 1207
    "SW": {-1, +1}, // ## 1206 -> (11, 05) -> (10, 06) -> ## 1107
    "NW": {-1, +0}, // ## 1206 -> (11, 05) -> (10, 05) -> ## 1106
}
```

![grid1306](docs/grid1306.png)

```go
hexDirectionVectors["even-column"] = {
    "N" : {+0, -1}, // ## 1306 -> (12, 05) -> (12, 04) -> ## 1305
    "NE": {+1, -1}, // ## 1306 -> (12, 05) -> (13, 04) -> ## 1405
    "SE": {+1, +0}, // ## 1306 -> (12, 05) -> (13, 05) -> ## 1406
    "S" : {+0, +1}, // ## 1306 -> (12, 05) -> (12, 06) -> ## 1307
    "SW": {-1, +0}, // ## 1306 -> (12, 05) -> (11, 05) -> ## 1206
    "NW": {-1, -1}, // ## 1306 -> (12, 05) -> (11, 04) -> ## 1205
}
```

## Roadmap

1. Replace the CLI with a web front end.
2. A future version of the tool will convert the turn report files into JSON data that you can use to create your own maps.
3. Usable documentation.