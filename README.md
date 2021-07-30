# Image-Filters

RGBA Convolution for Blur (B), Greyscale (G), Sharpen (S) and Edge Detection (E)

# Usage

- Clone this repository (git clone <clone link>)
- Place all .png images that you would like to apply filters on into data/in/images
- Change effects.txt to point to the images and the corresponding effects that you would like to apply on them.
  e.g. {"inPath": "LondonEye.png", "effects": ["G","E","S","B"]} (The filters are applied in order)
- cd into editor/
- COMMANDS supported:
  - go run editor.go images (this will run sequentially)
  - go run editor.go images bsp <num threads> (this will run BSP with the number of threads specified)
  - go run editor.go images pipeline <num threads> (this will run pipeline with the number of threads specified)
- The output images will be in data/out/

Please contact me for a report analysing the performance of the three paradigms.
