# FizzBuzz
print('\n'.join(str(i)
                if (line := 'Fizz' if i % 3 == 0 else '' + 'Buzz' if i % 5 == 0 else '') == ''
                else line
                for i in range(1, 101)))
