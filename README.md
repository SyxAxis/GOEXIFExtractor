# GOEXIFExtractor

Uses the GO library : https://github.com/dsoprea/go-exif.git

Working on a very large project to handle all the metadata from over 1,000 JPGs to be prepared for a photo book to be published. 
Used Lightroom to dump to JPG with ALL METADATA, then needed something to extract the EXIF so it could be used next to every single image as a quote.

So I wrote this code to scan down a directory tree looking for JPG files, extarcting the key 10 EXIF fields I needed and dumping into a CSV.

If you find it useful, help yourself.

Usage:

Normal use ( pulls around 20 most common EXIF fields by default ):

.\EXIFExtract -filepath "C:\ExportedPictures" -csvfile results.csv

Specify the EXIF fields you want :

.\EXIFExtract -filepath "C:\ExportedPictures" -csvfile results.csv -exif-fields "ImageDescription,ExposureTime,FNumber,ISOSpeedRatings"

