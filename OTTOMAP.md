# OttoMap Command Line Interface

OttoMap is a tool that translates TribeNet turn report files into JSON data and generates maps. This document provides instructions for running OttoMap from the command line.

The basic usage is:

1. Copy your turn report into a text file and save it to your `data/input` folder.
3. Run the `render` command to generate the WXX files.

## Important Assumptions

These instructions assume that you're running `ottomap` (on macOS and Linux) or `ottomap.exe` (on Windows) from the root directory of the project. We'll use the default directories (`data/input` and `data/output`) to keep the examples short.

> **NOTE**:
> - On Windows, use `ottomap.exe` to run commands.
> - On macOS and Linux, use `./ottomap` to run commands.

## Storing Turn Report Files

OttoMap expects turn report files to be stored in the `data/input` directory within the project. This directory includes `.gitignore` files to prevent accidentally uploading sensitive data to version control systems.

The report files must be text files (we can't parse the `.docx` files yet).
They must be named according to the following format:

      YYYY-MM.CLAN.report.txt

Where `YYYY` is the turn year (eg, 901), `MM` is the month (eg, 02), `CLAN` is the clan code (eg, 0991).
The extension `report.txt` tells OttoMap that it is a report file.

You can create the report files by opening up your turn report, selecting the entire file (Control-A or Command-A) and then pasting it into the text file.

> **NOTE**
> You should include all of your turn report files, not just the ones you want to generate a map for.

## Creating Maps

The `render` command reads the configuration and generates maps for each turn report.
Only the files in the `data/input` folder that match the YYYY-MM.CLAN.report.txt pattern are processed.

```bash
$ ottomap render --clan-id 0991 --show-grid-coords
```

Output example:
```
14:25:55 render.go:156: data:   /Users/wraith/JetBrains/tribenet/tn3/0138/data
14:25:55 render.go:157: input:  /Users/wraith/JetBrains/tribenet/tn3/0138/data/input
14:25:55 render.go:158: output: /Users/wraith/JetBrains/tribenet/tn3/0138/data/output
14:25:55 inputs.go:21: collect: input path: /Users/wraith/JetBrains/tribenet/tn3/0138/data/input
14:25:55 render.go:164: inputs: found 4 turn reports
14:25:55 render.go:203: "0901-04.0138": parsed      2 units in 878.75µs
14:25:55 render.go:203: "0901-05.0138": parsed      2 units in 648.209µs
14:25:55 render.go:203: "0901-06.0138": parsed      2 units in 999.167µs
14:25:55 render.go:203: "0901-07.0138": parsed      2 units in 993.334µs
14:25:55 render.go:205: parsed 4 inputs in to 4 turns and 8 units 4.835208ms
14:25:55 render.go:249: 0901-04:        2 units
14:25:55 render.go:249: 0901-05:        2 units
14:25:55 render.go:249: 0901-06:        2 units
14:25:55 render.go:249: 0901-07:        2 units
14:25:55 render.go:300: links: 6 good, 0 bad
14:25:55 render.go:350: updated        0 obscured 'Previous Hex' locations
14:25:55 render.go:351: updated        0 obscured 'Current Hex'  locations
14:25:55 walk.go:16: walk: input:        4 turns
14:25:55 walk.go:119: walk:        4 nodes: elapsed 60.375µs
14:25:55 map_world.go:31: map: collected       59 tiles
14:25:55 map_world.go:47: map: upper left  grid SA 1904
14:25:55 map_world.go:48: map: lower right grid SA 2912
14:25:55 map_world.go:49: map: todo: move grid creation from merge to here
14:25:55 map_world.go:158: map: collected       59 new     hexes
14:25:55 render.go:422: map:       59 nodes: elapsed 4.991917ms
14:25:55 writer.go:36: wxx: create: 59 tiles
14:25:55 writer.go:76: map: tile columns   11 rows    9
14:25:55 writer.go:85: map: tile columns   14 rows   12
14:25:55 render.go:430: created  /Users/wraith/JetBrains/tribenet/tn3/0138/data/output/0138.wxx
14:25:55 render.go:432: elapsed: 7.839ms

```

> **TODO**
> I need to document parsing errors.
> The report files sometimes contain typos that need to be updated.
> If you look at the docs/ERRORS.md file, you'll see examples of most.
> That really should be integrated with this document.

## Available Commands

OttoMap provides the following commands:

### `version`

The `version` command displays the current version of OttoMap.

```bash
$ ottomap version
```

Output example:
```
0.13.0
```

### `render`

The `render` command generates a map from the turn report files.

```bash
$ ottomap render --clan-id 0991 --show-grid-coords
```

You can specify additional options for the `map` command:

- `--turn`: Specify the last turn to generate a map for.

## Running OttoMap

To run OttoMap, follow these steps:

1. **Open a terminal or command prompt**:
    - On Windows, press `Win + R`, type `cmd`, and press `Enter`.
    - On macOS, press `Cmd + Space`, type `Terminal`, and press `Enter`.
    - On Linux, press `Ctrl + Alt + T`.

2. **Navigate to the OttoMap project directory**:
    - Use the `cd` command followed by the path to your project directory. For example:
      ```bash
      cd path/to/ottomap
      ```

3. **Place your turn report files in the `data/input` directory**.

4. **Run the desired command with any necessary options**:
    - To generate a map using the default settings, run:
      ```bash
      $ ottomap render --clan-id 0991
      ```

This will read all the turn report files in the `data/input` folder and create the Worldographer file `data/output/0991.wxx`.

Note: Detailed information about the configuration options and map generation settings can be found in the project's documentation.
