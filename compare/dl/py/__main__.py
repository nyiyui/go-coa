import sys
import urllib.request


def download(url: str, to: str):
    with open(to, 'wb') as dst_file:
        dst_file.write(urllib.request.urlopen(url).read())


if __name__ == '__main__':
    with open(sys.argv[1]) as file:
        for i, line in enumerate(file):
            if line == '' or line[0] == '#':
                continue
            to, url = line.split(' ')
            url = url[:-1]
            try:
                download(url, to)
            except Exception as err:
                print(f'error: {str(err)}')
            else:
                print(f'downloaded {url} to {to}')
