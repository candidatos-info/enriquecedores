import sys 

if __name__ == "__main__":
    if len(sys.argv) != 3:
        sys.exit("invalid number of arguments: {}".format(len(sys.argv)))
    user_name = sys.argv[1]
    password = sys.argv[2]
    # TODO read file app.yaml and apply replace in content using regexp
    