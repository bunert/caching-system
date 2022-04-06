# importing csv module
import csv
  
# csv file name
filename = "eclog_1day.csv"
  
# initializing the titles and rows list
fields = []
rows = []
start_timestamp = 0

def process(rows, r):
    # remove UserId
    r.pop(1)
    # IpId, TimeStamp, HttpMethod, Uri, HttpVersion, ResponseCode, Bytes, Referrer, UserAgent
    # compute diff from start_timestamp
    # number of 100-nanosecond intervals that have elapsed
    # 100 ns = 0.0001 ms
    if r[2] == "GET":
        r[1] = (int(r[1])-start_timestamp)/10000
        rows.append(r)
  
# reading csv file
with open(filename, 'r') as csvfile:
    # creating a csv reader object
    csvreader = csv.reader(csvfile)
      
    # extracting field names through first row
    fields = next(csvreader)

    # extract start_timestamp from first data row    
    first_row = next(csvreader)
    start_timestamp = int(first_row[2])
    process(rows, first_row)

    # extracting each data row one by one
    for row in csvreader:
        process(rows, row)
  
    # get total number of rows
    print("Total no. of rows: %d"%(csvreader.line_num))
  

# get totabl number of GET requests
print("Total no. of GET requests: %d"% len(rows))

# print proportion of GET requests compared to total
print("Proportion of GET requests: {:.2%}".format(len(rows)/(csvreader.line_num)))

# printing the field names
print('Field names are:' + ', '.join(field for field in fields))

#  printing first 5 rows
print('\nFirst 5 rows are:\n')
for row in rows[:5]:
    # parsing each column of a row
    for col in row:
        print("%10s"%col)
    print('\n')