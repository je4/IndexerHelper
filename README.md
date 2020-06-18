# Histogram
erzeugt mit Hilfe von ImageMagick (convert) ein Histogramm 
basierend auf einer Colormap, welche im Konfigurationsfile definiert wird.

## Installation

    go get gitlab.switch.ch/memoriav/memobase-2020/services/histogram
    go build gitlab.switch.ch/memoriav/memobase-2020/services/histogram/cmd/histogram
    go build gitlab.switch.ch/memoriav/memobase-2020/services/histogram/cmd/webservice

## Start:

### Commandline Histogram
    histogram -cfg histogram.toml -img bildchen.jpg
    
### Webservice        
    webservice -cfg histogram.toml
    
    
    
## Beispiel:
    PS C:\temp> /daten/go/bin\histogram.exe -cfg /daten/go/src\gitlab.switch.ch\memoriav\memobase-2020\services\histogram\histogram.toml -img act_binde.png
    {
     "BlueViolet": 27,
     "DimGray": 170,
     "DustyRose": 60,
     "Gray": 90,
     "GreenCopper": 8906,
     "Plum": 417,
     "Thistle": 226,
     "Violet": 104
    }    

## Webservice request

### Histogram
    #curl http://localhost:83/histogram/mnt/c/temp/Icon_pudelrudel_video.png
    {"Brown":600,"HotPink":401,"NeonPink":971,"Pink":62,"Plum":5225,"Scarlet":2034,"VioletRed":707}
    
### ValidateVideo
    #curl http://localhost:83/validateav/mnt/c/temp/pudel.mp4
    {"status":"error","message":""}
    
    #curl http://localhost:83/validateav/mnt/c/Users/juergen.enge/Downloads/Tony%20Conrad%27s%20Art%20Show%20at%20Greene%20Naftali%20Gallery%20-%20Art%20Show.webm
    {"status":"ok","message":"[matroska,webm @ 0x7fffca78dc40] Element at 0x951cfd ending at 0x951d0e exceeds containing master element ending at 0x951cf1"}
    
### ValidateImage
    #curl http://localhost:83/validateimage/mnt/c/daten/go/src/gitlab.switch.ch/memoriav/memobase-2020/services/histogram/parliamentdefect.jpg
    {"status":"error","message":"identify: Premature end of JPEG file `/mnt/c/daten/go/src/gitlab.switch.ch/memoriav/memobase-2020/services/histogram/parliamentdefect.jpg' @ warning/jpeg.c/JPEGWarningHandler/352."}
        