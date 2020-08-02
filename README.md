# GOEXIFExtractor

Uses the GO library : https://github.com/dsoprea/go-exif.git

Working on my first book ( "Photographing London" ) for Fotovue.com, we had to get images ready for publishing. That meant extracting comments, ISO, aperature, shutter speed, camera and lens info for over 1,000 images so the metadata could be placed next to every single image in the book. 

First used Lightroom to ensure all the metadata was correct, especially the comments and titles. Next used Jeff Friedl's superb Lightroom plugins to collate image folder layouts and dump them to disk based layouts ( LR can't do this. ). Dump all selected images as JPG with ALL METADATA.

I wrote this utilit in GO to scan down a directory tree looking for JPG files, extracting the key EXIF fields I needed and dumping into a CSV.

If you find it useful, help yourself.

Usage:

Normal use ( pulls around 20 most common EXIF fields by default ):

.\EXIFExtract -filepath "C:\ExportedPictures" -csvfile results.csv

Specify the EXIF fields you want :

.\EXIFExtract -filepath "C:\ExportedPictures" -csvfile results.csv -exif-fields "ImageDescription,ExposureTime,FNumber,ISOSpeedRatings"

