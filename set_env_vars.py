import sys 
import re

"""This script get USER_NAME and PASSWORD environtment
variables used on app engine from Github Secrets and
replace on app.yaml."""

app_engine_file = "app.yaml"

if __name__ == "__main__":
    if len(sys.argv) != 3:
        sys.exit("invalid number of arguments: {}".format(len(sys.argv)))
    user_name = sys.argv[1]
    password = sys.argv[2]
    file_content = ""
    with open (app_engine_file, "r") as file:
        app_engine_file_content = file.read()
        line = re.sub(r"##PASSWORD", password, app_engine_file_content) # replacing ##PASSWORD by the given password
        line = re.sub(r"##USER_NAME", user_name, line) # replacing ##USER_NAME by the given user name
        file_content = line
    with open (app_engine_file, "w") as file:
        file.write(file_content)
        