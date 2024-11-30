# Data

NOTE:
Almost all tables will have Clan as part of the key. This allows us to store reports for multiple players while preventing them from seeing each other's data.

## Key Types

1. Turn - a turn is identified by a Year and a Month. It is always formatted as YYYY-MM.
2. Year - a year in the game is an integer in the range 899 through 1234.
3. Month - a month in the game is an integer in the range 1 through 12.
4. Clan - players are identified by Clan. It is an integer in the range 1 through 999. It is always formatted as a four-digit number.
5. Unit - a unit is an element that can move. It is a string consisting of the Clan followed by a two-character suffix.
6. Hex - a hex is a tile on the game map. Hexes on the map are flat-topped. They are identified by their row and rolumn, which are both positive integers.
7. Direction - a code indicating the direction of movement. Examples are "N" (north), "SE" (southeast), and "S" (south).
8. Terrain - a code describing the type of terrain in a hex. The Terrain is used to generate the map. Examples are "PR" (prairie) or "SW" (swamp).
9. Border - a code describing the border separating two hexes. Examples are "R" (river) or "C" (canal).
10. Passage - a code describing a passage across or through the border separating two hexes. Examples are "SR" (stone road) or "MP" (mountain pass). 
11. Resources - a code describing a resource in a hex. Examples are "Coal Mine" or "Iron Ore."
12. Settlement - a settlement is identified by its location and a name. Settlements are not permanent, so we must track the turns that it existed.
13. Transients - units are free to move between hexes. Transient records capture where they ended a turn.

## Tile
A tile is a location on the game board which is used when generating the map.
Tiles are uniquely identified by their location.

Tiles are assigned attributes such as terrain, borders, passages, and resources.
This data is added to the Tile as units explore the game board.

## Transients
Transients are identified by Turn, Location, and Unit.
Multiple units may occupy the same location during a turn.

## Settlements
Settlements are identified by a generated integer.
Settlements are built by players, maintained, and may be abandoned, so we must track their duration.
We do this with two columns, Eff_Turn and End_Turn.
Eff_Turn is the first Turn that a Settlement was seen.
Eff_Turn is initially set to the last turn (9999-12) and is only updated when the settlement is abandoned.
When that happend, Eff_Turn is updated.
For mapping purposes, a Settlement is placed on the map starting on Eff_Turn and removed starting on End_Turn.
Settlement names are not required to be unique.

## Movement

Units may travel between hexes during the course of the game.
Move captures the starting location, the direction of travel, and all results of the move.
A Move is uniquely identified by the Unit, Turn, and a sequence number indicating the order of the move within the turn. 

A Step captures the results of a Move.  

1. Starting Location - the Hex the Unit was moving from.
2. Direction - the Direction the Unit wanted to move.
3. Succeeded - a boolean that is true only if the Move succeeded.
4. Reason For Failure - text describing why the Move failed. Optional and set only if the Move failed.
5. Ending Location - the Hex the Unit ended up in. If the Move failed, this will be the same value as the Starting Location.
6. Terrain - the Terrain of the Ending Location's hex.
7. Borders - each edge of the hex may have a border.
8. Passages - each edge of the hex may have a passage.
9. Settlement - a step may find a settlement.
10. Units - a step may find multiple units.

## Schema

```sqlite
-- Define the tables for the schema

-- --------------------------------------------------------------------------
-- Turns
--
-- The application assumes that we have a start turn of 899-12 and
-- end turn of 9999-12 pre-populated.
CREATE TABLE turns
(
    id    INTEGER PRIMARY KEY, -- calculated as (year-899) * 12 + month
    year  INTEGER CHECK (year BETWEEN 899 AND 9999),
    month INTEGER CHECK (month BETWEEN 1 AND 12),
    UNIQUE (year, month)
);

-- --------------------------------------------------------------------------
-- Clans
--
-- We currently store only one clan per database for security,
-- so this field is for reference only until that changes
-- (which will be never since it requires buy-in from all players).
CREATE TABLE clans
(
    id INTEGER PRIMARY KEY -- Range 1-999
);

-- --------------------------------------------------------------------------
-- Units
--
-- Note that we never want to add scout units to the Transients table.
-- The format of the id is xxxx for the clan and tribes,
-- xxxx([cefg][1-9]) for couriers, elements, fleets, and garrisons,
-- and xxxx([cefg][1-9])?(s[1-8]) for scouts.
CREATE TABLE units
(
    id       TEXT PRIMARY KEY,
    clan_id  INTEGER NOT NULL REFERENCES clans (id),               -- not really needed for single-clan database
    is_scout INTEGER NOT NULL DEFAULT 0 CHECK (is_scout in (0, 1)) -- true only if unit is a scout
);

-- --------------------------------------------------------------------------
-- Border Codes
-- 
-- This table stores the codes that describe a tile border.
CREATE TABLE border_codes
(
    code        TEXT NOT NULL PRIMARY KEY, -- R, CANAL, etc.
    descr       TEXT NOT NULL,             -- River, Canal, etc.
    wxx_feature TEXT NOT NULL,             -- how to draw the border in worldographer
    UNIQUE (descr)
);

-- --------------------------------------------------------------------------
-- Item Codes
-- 
-- This table stores the codes that describe an item that can be found in a tile.
--
-- I don't think we need to track these; the map render is not even aware of them.
CREATE TABLE item_codes
(
    code  TEXT NOT NULL PRIMARY KEY, -- JEWELS, PONIES, RICH PERSON, etc.
    descr TEXT NOT NULL,             -- Jewels, Ponies, Rich Person, etc.
    UNIQUE (descr)
);

-- --------------------------------------------------------------------------
-- Passage Codes
-- 
-- This table stores the codes that describe a tile passage.
CREATE TABLE passage_codes
(
    code        TEXT NOT NULL PRIMARY KEY, -- FORD, PASS, STONY ROAD, etc.
    descr       TEXT NOT NULL,             -- Ford, Mountain Pass, Stony Road, etc.
    wxx_feature TEXT NOT NULL,             -- how to draw the passage in worldographer
    UNIQUE (descr)
);

-- --------------------------------------------------------------------------
-- Resource Codes
-- 
-- This table stores the codes that describe a tile resource.
CREATE TABLE resource_codes
(
    code        TEXT NOT NULL PRIMARY KEY, -- COAL, IRON ORE, etc.
    descr       TEXT NOT NULL,             -- Coal, Iron Orer, etc.
    wxx_feature TEXT NOT NULL,             -- how to draw the resource in worldographer
    UNIQUE (descr)
);

-- --------------------------------------------------------------------------
-- Terrain Codes
-- 
-- This table stores the codes that describe a tile terrain.
CREATE TABLE terrain_codes
(
    code        TEXT NOT NULL PRIMARY KEY, -- PR, LJM, etc
    long_code   TEXT NOT NULL,             -- PRAIRIE, LOW JUNGLE MOUNTAINS, etc
    descr       TEXT NOT NULL,
    wxx_terrain TEXT NOT NULL,
    UNIQUE (long_code),
    UNIQUE (descr)
);

-- --------------------------------------------------------------------------
-- the tile tables are used to render the map. the map generator understands
-- the effective date logic on the tables and uses it to create maps that
-- show the results "as of" a particular turn. future generators might even
-- use that information to trace movement paths for units.
-- --------------------------------------------------------------------------

-- --------------------------------------------------------------------------
-- Tiles
--
-- The direction columns (north, south, etc.) link tiles to their neighbors.
-- I am not sure that they are needed, but they make navigation queries simpler.
--
-- The last visited/last scouted values can be derived from either the Moves or
-- Transients tables. They may be removed if they make updates too expensive.
--
-- Tile attributes are stored in child tables because the values can change from
-- turn to turn or even move to move. For example, Fleet Movement could report a
-- tile as Unknown Water in one move, and then as Ocean in another.
--
-- It would be great if we could build a unique key on grid, row, and col, but
-- we can't since the early turn report obscured the grid. This setup, though,
-- allows us to easily update the grid, row, and col when we are able to compute
-- their values.
--
-- Anyway, we have to treat them as mutable since players are required to provide
-- missing values for early turn reports. This values will likely be updated once
-- the player gets reports that have the actual grid values.
CREATE TABLE tiles
(
    id              INTEGER PRIMARY KEY,
    grid            TEXT    NOT NULL,              -- usually ## or AA through ZZ, sometimes N/A
    row             INTEGER NOT NULL,              -- 0 only when grid is N/A
    col             INTEGER NOT NULL,              -- 0 only when grid is N/A
    north           INTEGER REFERENCES tiles (id),
    north_east      INTEGER REFERENCES tiles (id),
    north_west      INTEGER REFERENCES tiles (id),
    south           INTEGER REFERENCES tiles (id),
    south_east      INTEGER REFERENCES tiles (id),
    south_west      INTEGER REFERENCES tiles (id),
    last_visited_on INTEGER REFERENCES turns (id), -- last turn the tile was visited by a unit
    last_scouted_on INTEGER REFERENCES turns (id)  -- last turn the tile was scouted by a unit
);

-- --------------------------------------------------------------------------
-- Tile Border Details
--
-- These are derived after parsing all the movement results for a turn.
-- In other words, these are the details for the tile at the end of the turn.
--
-- We have to treat tile borders as mutable data because there are bugs
-- in the report generation process.
--
-- Assumption: each border of a tile can contain only one border feature.
-- This is likely invalid because of bugs.
-- 
-- The application is responsible for ensuring that the effective dated logic remains
-- consistent for all rows.
CREATE TABLE tile_border_details
(
    tile_id   INTEGER NOT NULL REFERENCES tiles (id),
    effdt     INTEGER NOT NULL REFERENCES turns (id), -- turn the entry becomes active
    enddt     INTEGER NOT NULL REFERENCES turns (id), -- turn the entry becomes inactive
    border_cd TEXT    NOT NULL REFERENCES border_codes (code),
    direction TEXT    NOT NULL CHECK (direction in ('N', 'NE', 'SE', 'S', 'SW', 'NW')),
    PRIMARY KEY (tile_id, border_cd, direction, effdt)
);

-- --------------------------------------------------------------------------
-- Tile Passage Details
--
-- These are derived after parsing all the movement results for a turn.
-- In other words, these are the details for the tile at the end of the turn.
--
-- We have to treat tile passages as mutable data because there are bugs
-- in the report generator and parser.
--
-- Assumption: each border of a tile can contain only one border feature.
-- This is likely invalid because of bugs.
-- 
-- The application is responsible for ensuring that the effective dated logic remains
-- consistent for all rows.
CREATE TABLE tile_passage_details
(
    tile_id    INTEGER NOT NULL REFERENCES tiles (id),
    effdt      INTEGER NOT NULL REFERENCES turns (id), -- turn the entry becomes active
    enddt      INTEGER NOT NULL REFERENCES turns (id), -- turn the entry becomes inactive
    passage_cd TEXT    NOT NULL REFERENCES passage_codes (code),
    direction  TEXT    NOT NULL CHECK (direction in ('N', 'NE', 'SE', 'S', 'SW', 'NW')),
    PRIMARY KEY (tile_id, effdt, passage_cd, direction)
);

-- --------------------------------------------------------------------------
-- Tile Resource Details
--
-- These are derived after parsing all the movement results for a turn.
-- In other words, these are the details for the tile at the end of the turn.
--
-- We have to treat tile resources as mutable data because there are bugs
-- in the report generator and parser.
--
-- Assumption: each tile can contain only one resource.
-- This is something that should be verified (but might be invalid because
-- of bugs, anyway).
--
-- The application is responsible for ensuring that the effective dated logic remains
-- consistent for all rows.
CREATE TABLE tile_resource_details
(
    tile_id     INTEGER NOT NULL REFERENCES tiles (id),
    effdt       INTEGER NOT NULL REFERENCES turns (id), -- turn the entry becomes active
    enddt       INTEGER NOT NULL REFERENCES turns (id), -- turn the entry becomes inactive,
    resource_cd TEXT    NOT NULL REFERENCES resource_codes (code),
    PRIMARY KEY (tile_id, effdt, resource_cd)
);

-- --------------------------------------------------------------------------
-- Tile Settlement Details
--
-- These are derived after parsing all the movement results for a turn.
-- In other words, these are the details for the tile at the end of the turn.
--
-- We have to treat settlements as mutable data because they can be destroyed or
-- abandoned. Also, there are bugs in the report generator and parser.
--
-- Assumption: tiles shouldn't have multiple settlements but there
-- are bugs in the report generation process and the parser, so we
-- have to allow them. We will silently merge duplicate names into
-- a single row, though.
--
-- Known issue: players won't know that a settlement has been
-- abandoned or destroyed until they send a unit to its location.
-- 
-- The application is responsible for ensuring that the effective dated logic remains
-- consistent for all rows.
CREATE TABLE tile_settlement_details
(
    tile_id INTEGER NOT NULL REFERENCES tiles (id),
    effdt   INTEGER NOT NULL REFERENCES turns (id), -- turn the entry becomes active
    enddt   INTEGER NOT NULL REFERENCES turns (id), -- turn the entry becomes inactive,
    name    TEXT    NOT NULL,
    PRIMARY KEY (tile_id, effdt, name)
);

-- --------------------------------------------------------------------------
-- Tile Terrain Details
--
-- These are derived after parsing all the movement results for a turn.
-- In other words, these are the details for the tile at the end of the turn.
--
-- We have to treat tile terrain as mutable data because of Fleet Movement reports.
-- Also, there are bugs in the report generator and parser.
--
-- Assumption: each tile can contain multiple terrain codes because of Fleet Movement
-- reports and bugs.
--
-- The application is responsible for ensuring that the effective dated logic remains
-- consistent for all rows.
CREATE TABLE tile_terrain_details
(
    tile_id    INTEGER NOT NULL REFERENCES tiles (id),
    effdt      INTEGER NOT NULL REFERENCES turns (id), -- turn the entry becomes active
    enddt      INTEGER NOT NULL REFERENCES turns (id), -- turn the entry becomes inactive,
    terrain_cd TEXT    NOT NULL REFERENCES units (id),
    PRIMARY KEY (tile_id, effdt, terrain_cd)
);

-- --------------------------------------------------------------------------
-- Tile Transient Details
--
-- These are derived after parsing all the movement results for a turn.
-- In other words, these are the details for the tile at the end of the turn.
--
-- We have to treat tile transients as mutable data because units are mobile and
-- there are bugs in the report generator and parser.
--
-- Unintended benefit of this table is it tracks where every unit ends the turn
-- as well as the last known location for any unit. It might be useful to add an
-- attribute to track the turn the unit was last seen.
--
-- Note: we must not add scout units to this table. If I knew how to enforce that
-- with a check constraint, I would.
-- 
-- The application is responsible for ensuring that the effective dated logic remains
-- consistent for all rows.
CREATE TABLE tile_transient_details
(
    tile_id INTEGER NOT NULL REFERENCES tiles (id),
    effdt   INTEGER NOT NULL REFERENCES turns (id), -- turn the entry becomes active
    enddt   INTEGER NOT NULL REFERENCES turns (id), -- turn the entry becomes inactive,
    unit_id TEXT    NOT NULL REFERENCES units (id),
    PRIMARY KEY (tile_id, effdt, unit_id)
);

-- --------------------------------------------------------------------------
-- the moves and move detail tables capture the results of each move.
-- they are not needed after the tile detail tables are updated. 
-- it may be cheaper just to keep them in memory and not load them at all.
-- --------------------------------------------------------------------------

-- --------------------------------------------------------------------------
-- Moves
--
-- This table stores information on all of the moves paresed from the turn reports.
-- We're assuming that there's no need to track the entry back to the source.
--
-- If a move fails, starting_tile and ending_tile must be set to the same value.
--
-- Warning: The Follow and Goes To moves don't have directions.
--
-- We could use a synthetic key (turn + unit + step) but that would make querying
-- the child tables irksome.
--
-- TODO: Fleet Moves have to be integrated into this somehow.
CREATE TABLE moves
(
    id             INTEGER PRIMARY KEY, -- unique identifier for the movement
    turn_id        INTEGER NOT NULL REFERENCES turns (id),
    unit_id        TEXT    NOT NULL REFERENCES units (id),
    step_no        INTEGER NOT NULL,    -- order of the step within the Move
    starting_tile  INTEGER NOT NULL REFERENCES tiles (id),
    action         TEXT    NOT NULL,    -- kind of movement (Still, Follow, Scout) or direction
    ending_tile    INTEGER NOT NULL REFERENCES tiles (id),
    terrain_cd     TEXT    NOT NULL REFERENCES terrain_codes (code),
    failure_reason TEXT,                -- set only if the move failed
    CONSTRAINT CHECK (action in ('STILL', 'SCOUT', 'N', 'NE', 'SE', 'S', 'SW', 'NW')),
    UNIQUE (turn_id, unit_id, step_no)
);

-- --------------------------------------------------------------------------
-- Move Border Details
-- 
-- This table stores details about the tile borders that were found during
-- a move. The details are the border feature and the edge.
--
-- The details are always for the ending tile of the move.
CREATE TABLE move_border_details
(
    move_id   INTEGER NOT NULL REFERENCES moves (id),
    border_cd TEXT    NOT NULL REFERENCES border_codes (code),
    edge      TEXT    NOT NULL CHECK (edge in ('N', 'NE', 'SE', 'S', 'SW', 'NW')),
    PRIMARY KEY (move_id, border_cd, edge)
);

-- --------------------------------------------------------------------------
-- Move Passage Details
-- 
-- This table stores details about the border passages that were found during
-- a move. The details are the type of passage and the edge.
--
-- The details are always for the ending tile of the move.
CREATE TABLE move_passage_details
(
    move_id    INTEGER NOT NULL REFERENCES moves (id),
    passage_cd TEXT    NOT NULL REFERENCES passage_codes (code),
    edge       TEXT    NOT NULL CHECK (edge in ('N', 'NE', 'SE', 'S', 'SW', 'NW')),
    PRIMARY KEY (move_id, passage_cd, edge)
);

-- --------------------------------------------------------------------------
-- Move Resource Details
-- 
-- This table stores details about the tile resources that were found during
-- a move. The details are the type of resource and the edge.
--
-- The details are always for the ending tile of the move.
CREATE TABLE move_resource_details
(
    move_id     INTEGER NOT NULL REFERENCES moves (id),
    resource_cd TEXT    NOT NULL REFERENCES resource_codes (code),
    PRIMARY KEY (move_id, resource_cd)
);

-- --------------------------------------------------------------------------
-- Move Settlement Details
-- 
-- This table stores the names of settlements that were found during a move.
--
-- The details are always for the ending tile of the move.
CREATE TABLE move_settlement_details
(
    move_id INTEGER NOT NULL REFERENCES moves (id),
    name    TEXT    NOT NULL,
    PRIMARY KEY (move_id, name)
);

-- --------------------------------------------------------------------------
-- Move Transient Details
-- 
-- This table stores the units that were found during a move.
--
-- The details are always for the ending tile of the move.
CREATE TABLE move_transient_details
(
    move_id INTEGER NOT NULL REFERENCES moves (id),
    unit_id TEXT    NOT NULL REFERENCES units (id),
    PRIMARY KEY (move_id, unit_id)
);

```