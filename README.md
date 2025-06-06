# shapr-backup

Simple application to export all synced drawings from Shapr3d

## Usage:
```
Usage: shapr-backup.exe [flags] [filenames/wildcards ...]

Flags:
      -h, --help            Show context-sensitive help.
      --project-root=<dir>  Default "$LOCALAPPDATA\Packages\Shapr3D.Shapr3D_dvv5p1vgwv6mp"
      -d, --target="."      Export directory ($EXPORT_DIR).
      -r, --add-revision    Add revision ID to filename.
      -s, --add-dirs        Make folders for export.
```

Note that directory names are ignored.  Filenames and wildcards are matched against the names of drawing, regardless of directory.

If multiple drawings match *flue* in different directories, where is no way to select only one.

## Suggested usage

What seems to work well is:
-	Right click on a drawing
-	Click on “select”  (goes into selection mode)
-	Press “CTRL-A” to select all drawings
-	Right click on a drawing, and pick “Download Now”
-	It should download all drawings..

Then just create an empty folder, CD into it, and run “shapr-backup”.

You should end up with a directory full of “.shapr” files containing working backups of all sync’d drawings.

Use “-r” to add the drawing revision to the file name. (useful if you have multiple backups)
Use “-s” to recreate directories instead of saving all files to current directory. (useful if you want to make it look exactly like the remote structure)

If the same file appears twice, you get a number after the filename..

