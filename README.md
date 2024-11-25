[На русском](README_ru.md)

# ImgToASCII
Program for converting images to ASCII.

This project is inspired by the video [I Tried Turning Games Into Text](https://www.youtube.com/watch?v=gg40RWiaHRY).

## Goal
To study algorithms for graphics processing.

## How it works
A series of filters are applied to the image, and then the result is output to an HTML file with the option to save colors. The filters are needed to preserve the boundaries of objects and output them as `_ / \ |` characters in ASCII.

Since the goal is to study algorithms, the focus is not on performance, so the image is processed on the CPU.

## Filters
### Original
![Original](./examples/raimei.png)
### Image in ASCII
![Image in ASCII](./examples/ascii.png)
### Image in ASCII with color
![Image in ASCII](./examples/ascii-colour.png)
### Gaussian blur
![Gaussian blur](./examples/gauss.jpg)
### Difference of Gaussians
![Difference of Gaussians](./examples/DoG.jpg)
### Sobel operator
Boundary angle is shown in color
![Sobel operator](./examples/sobel-colour.jpg)
### Sobel operator + Difference of Gaussians
Boundary angle is shown in color
![Sobel operator + Difference of Gaussians](./examples/sobel-dog-colour.jpg)

## How to compile
1. Install [Go](https://go.dev)
2. Copy the source code
3. Compile with the command `go build . -o <output_file>`

## How to use
`./<compiled_file> <image_file> true/false`

true/false - get ASCII in color or not