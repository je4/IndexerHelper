# Histogram
erzeugt mit Hilfe von ImageMagick (convert) ein Histogramm 
basierend auf einer Colormap, welche im Konfigurationsfile definiert wird.

## Installation

    go get gitlab.switch.ch/memoriav/memobase-2020/services/histogram
    go build gitlab.switch.ch/memoriav/memobase-2020/services/histogram

## Start:

    histogram -cfg histogram.toml -img bildchen.jpg
    
    
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
