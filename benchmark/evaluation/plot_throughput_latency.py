import sys
import csv
import matplotlib.pyplot as plt
import numpy as np

file_path = str(sys.argv[1])

with open(file_path, 'r') as file:

    csv_reader = csv.reader(file, delimiter=',')

    x, y = [], []

    for row in csv_reader:
        x.append(row[0])
        y.append(row[1])

    
    plt.plot(x, y, '.')
    plt.show()