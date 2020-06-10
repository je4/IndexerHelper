# Histogram
erzeugt mit Hilfe von ImageMagick (convert) ein Histogramm 
basierend auf einer Colormap, welche im Konfigurationsfile definiert wird.

## Installation

    go get gitlab.switch.ch/memoriav/memobase-2020/services/histogram
    go build gitlab.switch.ch/memoriav/memobase-2020/services/histogram

## Start:

    histogram -cfg histogram.toml -img bildchen.jpg