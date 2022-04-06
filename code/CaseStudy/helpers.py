
#!/usr/bin/python
"""
Helper functions

"""
 

from datetime import date, datetime
import logging
import json
import sys
import os

# check cmdline argument for dates
# expected format: %Y%m%d
def valid_date(s):
    try:
        return date.strftime(datetime.strptime(s, "%Y%m%d"), "%Y%m%d")
    except ValueError:
        msg = "Not a valid date: hehehe '{0}'.".format(s)
        raise TypeError(msg)

# setup logger
def setup_logger(verbose_flag):
    logger = logging.getLogger()
    logger.setLevel(logging.DEBUG)
    ch = logging.StreamHandler(sys.stdout)
    formatter = logging.Formatter('%(asctime)s - %(levelname)s - %(message)s')
    ch.setFormatter(formatter)
    logger.addHandler(ch)
    if verbose_flag == False:
        ch.setLevel(logging.ERROR)
    else:
        ch.setLevel(logging.DEBUG)

# open json file
def open_json(file):
    try:
        with open(file, 'r') as fp:
            return json.load(fp)
    except IOError:
        logging.error('file %s not found, exit', os.path.basename(file))
        exit()


# write json file
def write_json(data, file):
    if (os.path.isfile(file)):
        if (not yes_or_no("Overwrite file"+  os.path.basename(file))):
            logging.error("file not overwritten, exit.")
            exit()
    
    logging.info("write file %s", os.path.basename(file))
    with open(file, 'w') as fp:
        json.dump(data, fp)

# prompt user to safe file
def yes_or_no(question):
    while "the answer is invalid":
        reply = str(input(question+' (y/n): ')).lower().strip()
        if reply[:1] == 'y':
            return True
        if reply[:1] == 'n':
            return False