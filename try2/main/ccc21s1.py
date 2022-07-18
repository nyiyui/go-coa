import sys

heights, widths = map(lambda line: list(map(int, line.split())), sys.stdin.read().split('\n')[1:3])
area = 0.0
for i in range(len(widths)):
    area += (heights[i] + heights[i + 1]) * widths[i]

area /= 2
print(int(area) if area.is_integer() else area)
