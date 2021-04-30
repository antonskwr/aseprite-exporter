# Aseprite Exporter

CLI tool for exporting multiple Aseprite projects keeping the initial folder tree structure.

```
Usage:
    export -execpath EXECUTABLE_PATH -source SOURCE_DIR -target TARGET_DIR -db MODIFIED_TIME_DB

Usage of export:
    -db string
        DB path for keeping project's last modified time
    -execpath string
        Path to aseprite executable
    -mute
        Mute target directory overwrite warning
    -source string
        Path to root directory with aseprite projects
    -target string
        Path to directory for project tree to be exported into
```
