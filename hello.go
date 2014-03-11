package main

import (
    "fmt"
    "image"
)

func main() {
    m := image.NewRGBA(image.Rect(56, 56, 100, 100))
    fmt.Println(m.Bounds())
    fmt.Println(m.At(66, 66).RGBA())
}
