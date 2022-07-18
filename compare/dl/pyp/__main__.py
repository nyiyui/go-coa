import multiprocessing.pool
import sys
import urllib.request


def download(url: str, to: str):
    with open(to, 'wb') as dst_file:
        dst_file.write(urllib.request.urlopen(url).read())


if __name__ == '__main__':
    with open(sys.argv[1]) as file:
        lines = list(filter(lambda line: line != "" and line[0] != "#", file))


    def download_line(line: str) -> None:
        to = line.split(' ')[0].lower()
        url = line[len(to) + 1:-1]
        try:
            download(url, to)
        except Exception as err:
            print(f'error: {str(err)}')
        else:
            print(f'downloaded {url} to {to}')


    with multiprocessing.pool.Pool() as pool:
        pool.map(download_line, lines)
