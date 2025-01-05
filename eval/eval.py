import pandas as pd
import sys

INFILE = str(sys.argv[1])

# import data
frame = pd.read_csv(INFILE)

# add column with experiment id (same for all repetitions of one parameter combination)
frame['EXPID'] = (
    frame['db_size'].astype(str) + 
    frame['key_length'].astype(str) + 
    frame['value_length'].astype(str) + 
    frame['malicious'].astype(str) + 
    frame['rate'].astype(str) + 
    frame['multi_client'].astype(str) + 
    frame['num_threads'].astype(str)
)

# add column of total bandwidth values
frame['BW_Total'] = (
    frame['BW_SendGetNotifiedDown'] +
    frame['BW_SendPIRUp'] +
    frame['BW_SendPIRDown'] +
    frame['BW_SendNotifyUp'] +
    frame['BW_SendNotifyDown'] +
    frame['BW_RecvGetNotifiedUp'] +
    frame['BW_RecvGetNotifiedDown'] +
    frame['BW_RecvPIRUp'] +
    frame['BW_RecvPIRDown'] +
    frame['BW_RecvNotifyUp'] +
    frame['BW_RecvNotifyDown'] +
    frame['BW_SendGetNotifiedUp'] +
    frame['BW_SendGetNotifiedDown']
).map(lambda x : x / 1024)

# add column of total latency values
frame['RT_Total'] = (
    frame['RT_SendPIR'] +
    frame['RT_SendNotify'] +
    frame['RT_RecvGetNotified'] +
    frame['RT_RecvPIR'] +
    frame['RT_RecvNotify'] +
    frame['RT_SendGetNotified']
)

# # scale latency columns
# frame['RT_Total'] = frame['RT_Total'].apply(lambda x : x / 1000)
# frame['RT_SendPIR'] = frame['RT_SendPIR'].apply(lambda x : x / 1000)
# frame['RT_SendNotify'] = frame['RT_SendNotify'].apply(lambda x : x / 1000)
# frame['RT_RecvGetNotified'] = frame['RT_RecvGetNotified'].apply(lambda x : x / 1000)
# frame['RT_RecvPIR'] = frame['RT_RecvPIR'].apply(lambda x : x / 1000)
# frame['RT_RecvNotify'] = frame['RT_RecvNotify'].apply(lambda x : x / 1000)
# frame['RT_SendGetNotified'] = frame['RT_SendGetNotified'].apply(lambda x : x / 1000)

# # scale latency columns
frame['RT_Total'] = frame['RT_Total'].apply(lambda x : x / 1000 /1000) # seconds
frame['RT_SendPIR'] = frame['RT_SendPIR'].apply(lambda x : x / 1000 /1000) # seconds
frame['RT_SendNotify'] = frame['RT_SendNotify'].apply(lambda x : x / 1000   /1000) # seconds
frame['RT_RecvGetNotified'] = frame['RT_RecvGetNotified'].apply(lambda x : x / 1000 /1000) # seconds
frame['RT_RecvPIR'] = frame['RT_RecvPIR'].apply(lambda x : x / 1000/1000)
frame['RT_RecvNotify'] = frame['RT_RecvNotify'].apply(lambda x : x / 1000 /1000) # seconds
frame['RT_SendGetNotified'] = frame['RT_SendGetNotified'].apply(lambda x : x / 1000 /1000) # seconds




print("EXPERIMENT 1: Bandwidth (KB) / No Auth")
tmp_frame = frame.loc[(frame['malicious'] == False) & (frame['multi_client'] == False)]
r_frame = tmp_frame[['EXPID', 'db_size', 'rate', 'BW_Total']].groupby(['EXPID']).mean().sort_values(['db_size', 'rate'])
r_frame['BW_Total_Norm'] = r_frame['BW_Total'] / r_frame['rate']
print(r_frame[['db_size', 'rate', 'BW_Total_Norm']])
print("\n")

print("EXPERIMENT 2: Bandwidth (KB) / Yes Auth")
tmp_frame = frame.loc[(frame['malicious'] == True) & (frame['multi_client'] == False)]
r_frame = tmp_frame[['EXPID', 'db_size', 'rate', 'BW_Total']].groupby(['EXPID']).mean().sort_values(['db_size', 'rate'])
r_frame['BW_Total_Norm'] = r_frame['BW_Total'] / r_frame['rate']
print(r_frame[['db_size', 'rate', 'BW_Total_Norm']])
print("\n")

print("EXPERIMENT 3: Computation / No Auth / All Values S")
tmp_frame = frame.loc[(frame['malicious'] == False) & (frame['multi_client'] == True)]
r_frame = tmp_frame[['EXPID', 'db_size', 'RT_SendPIR', 'RT_SendNotify', 'RT_RecvGetNotified', 'RT_RecvPIR', 'RT_RecvNotify', 'RT_SendGetNotified', 'RT_Total']].groupby(['EXPID']).mean().sort_values('db_size')
print(r_frame)
print("\n")

print("EXPERIMENT 3: Computation / Yes Auth / All Values S")
tmp_frame = frame.loc[(frame['malicious'] == True) & (frame['multi_client'] == True)]
r_frame = tmp_frame[['EXPID', 'db_size', 'RT_SendPIR', 'RT_SendNotify', 'RT_RecvGetNotified', 'RT_RecvPIR', 'RT_RecvNotify', 'RT_SendGetNotified', 'RT_Total']].groupby(['EXPID']).mean().sort_values('db_size')
print(r_frame)
print("\n")


print("EXPERIMENT 4: Computation / No Auth / All Values S")
tmp_frame = frame.loc[(frame['malicious'] == False) & (frame['multi_client'] == False) &(frame['rate']==1)]
r_frame = tmp_frame[['EXPID', 'db_size', 'RT_SendPIR', 'RT_SendNotify', 'RT_RecvGetNotified', 'RT_RecvPIR', 'RT_RecvNotify', 'RT_SendGetNotified', 'RT_Total']].groupby(['EXPID']).mean().sort_values('db_size')
print(r_frame)
print("\n")

print("EXPERIMENT 4: Computation / Yes Auth / All Values S")
tmp_frame = frame.loc[(frame['malicious'] == True) & (frame['multi_client'] == False) & (frame['rate']==1)]
r_frame = tmp_frame[['EXPID', 'db_size', 'RT_SendPIR', 'RT_SendNotify', 'RT_RecvGetNotified', 'RT_RecvPIR', 'RT_RecvNotify', 'RT_SendGetNotified', 'RT_Total']].groupby(['EXPID']).mean().sort_values('db_size')
print(r_frame)
print("\n")