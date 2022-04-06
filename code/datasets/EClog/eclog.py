import pandas as pd
import matplotlib.pyplot as plt

df = pd.read_csv("eclog_1day.csv")

# print type of each column
print(df.dtypes)
print("\n")

# drop UserAgent (unused for now)
df.drop('UserAgent', axis=1, inplace=True)

# get start_timestamp of first row
start_timestamp = df['TimeStamp'].iloc[0]



# get total number of rows
total_number = len(df)
print("Total no. of rows: {}".format(total_number))

# drop if not a GET request
indexNames = df[df['HttpMethod'] != 'GET'].index
df.drop(indexNames , inplace=True)

total_get = len(df)
# get totabl number of GET requests
print("Total no. of GET requests: {}".format(total_get))

# print proportion of GET requests compared to total
print("Proportion of GET requests: {:.2%}".format(total_get/total_number))

# change timestamp to difference from start_timestamp
df['TimeStamp'] = df['TimeStamp'].apply(lambda x: (x-start_timestamp)/10000)

# number of request per object
objects_counts = df.groupby('Uri')['IpId'].count()
print(objects_counts.size)
objects_counts.plot(kind='bar')
plt.show()

# peek into data
print("\n Data sample:")
pd.set_option('display.max_rows', None)
pd.set_option('display.max_columns', None)
pd.set_option('display.width', 2000)
pd.set_option('display.float_format', '{:20,.2f}'.format)
pd.set_option('display.max_colwidth', None)
print(df.head(10))