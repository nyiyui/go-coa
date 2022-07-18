#!/usr/bin/env python3
import colorsys
import sys

from PIL import Image

IMG_PATH = sys.argv[1]
IMG_WIDTH = int(sys.argv[2])
IMG_HEIGHT = int(sys.argv[3])


def get_pixel_iter(x: int, y: int):
    c1 = complex(x, y)
    c2 = complex(0, 0)
    for n in range(1, 900):
        if abs(c2) > 2:
            return n
        c2 = c2 ** 2 + c1
    return 0


def get_pixel(x: int, y: int):
    return tuple(map(lambda v: int(v * 255), colorsys.hsv_to_rgb(get_pixel_iter(x, y) / 254, 0.5, 0.5)))


if __name__ == '__main__':
    img = Image.new('RGB', (IMG_WIDTH, IMG_HEIGHT))
    for x in range(img.width):
        for y in range(img.height):
            img.putpixel((x, y), get_pixel(
                (x - 0.75 * img.width) / (0.25 * img.width),
                (y - 0.25 * img.height - 0.25 * img.width) / (0.25 * img.width),
            ))
    img.save(IMG_PATH)
