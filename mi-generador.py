import sys

filename = sys.argv[1]
n = int(sys.argv[2])

with open(filename,"w+") as f:
    for i in range(1, n+1):
        f.write("client" + str(i) + ":")
