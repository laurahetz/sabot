
import pandas as pd
import numpy as np
import sys

def get_bw_tex(df, sysname):
    out = ""
    unique_rates = df['rate'].unique()
    for rate in unique_rates:
        # Filter rows for the current rate
        filtered_df = df[df['rate'] == rate]

        # Concatenate 'BW_Total_Norm' values for all 'db_size' in ascending order
        bw_norm_values = [
            f"\\qty{{{row['BW_Total_Norm']:.2f}}}{{\\kibi\\byte}}"
            for _, row in filtered_df.sort_values('db_size').iterrows()
        ]
        concatenated_string = " & ".join(bw_norm_values)

        # Write the string to the file
        out += f"\\{sysname}{{}} ({rate}) & {concatenated_string} \\\\\n"
    return out

def process_bw_csvs(input_file, output_file):

    header =  "\\scalebox{0.8}{%\n"
    header += "\\begin{tabularx}{0.65\\textwidth}{lrrrrr}\\\\\\toprule\n"
    header += "& \\multicolumn{5}{c}{Number of clients}\\\\\n"
    header += "& $2^{10}$ & $2^{12}$ & $2^{14}$ & $2^{16}$ & $2^{18}$\\\\\\midrule\n"

    relatedwork = "Alpenhorn       & \\qty{7.08}{\\mebi\\byte} & \\qty{7.17}{\\mebi\\byte}& \\qty{7.53}{\\mebi\\byte}& \\qty{8.97}{\\mebi\\byte}& \\qty{14.10}{\\mebi\\byte}\\\\\\addlinespace\n"
    relatedwork += "Pudding w/o Nym & \\qty{125.57}{\\kibi\\byte} & \\qty{125.57}{\\kibi\\byte}         & \\qty{125.57}{\\kibi\\byte}         & \\qty{125.57}{\\kibi\\byte}         &\\qty{125.57}{\\kibi\\byte}\\\\\n"
    relatedwork += "Pudding w/ Nym  & \\qty{6.99}{\\mebi\\byte} & \\qty{6.99}{\\mebi\\byte} & \\qty{6.99}{\\mebi\\byte} & \\qty{6.99}{\\mebi\\byte}      & \\qty{6.99}{\\mebi\\byte}\\\\\\bottomrule\n"
    relatedwork += "\\end{tabularx}\n}"

    # Read the CSV file with headers
    df_bw_noauth = pd.read_csv(f'{input_file}_BW_NoAuth.csv')
    df_bw_auth = pd.read_csv(f'{input_file}_BW_Auth.csv')

    out_noauth = get_bw_tex(df_bw_noauth, "hprot")
    out_auth = get_bw_tex(df_bw_auth, "mprot")

    # Write the BW result to a text file (tab-separated for clarity)
    # Open the output file and manually format the lines
    with open(output_file, 'w') as f:
            f.write(f"{header}")
            # Write the string to the file
            f.write(f"{out_noauth}\\addlinespace\n{out_auth}\\addlinespace\n")
            f.write(f"{relatedwork}")


def process_rt_csvs(input_file, output_file):
    # Read the CSV file with headers
    df_noauth = pd.read_csv(f'{input_file}_RT_Multi_NoAuth.csv')
    df_auth = pd.read_csv(f'{input_file}_RT_Multi_Auth.csv')

    # Save the DataFrames in an array
    sysnames = ['hprot', 'mprot']

    unique_db_sizes = sorted(df_auth['db_size'].unique())
    unique_rates = sorted(df_auth['rate'].unique())

    with open(output_file, 'w') as f:
            # Write the string to the file
            header = "\\scalebox{0.8}{%\n"
            header += "\\begin{tabularx}{0.86\\textwidth}{rlcrrrrrrr}\\toprule"
            header += "& & Rate & S-Retrieval & S-Notify & R-GetNotify & R-Retrieval & R-Notify & S-GetNotify & \\textbf{Total}\\\\\n\\midrule\n"
            f.write(f"{header}")

            for db_size in unique_db_sizes:
                numrows = len(unique_rates)*2
                out = f"\\multirow{{{numrows}}}{{*}}{{$2^{{{int(np.log2(db_size))}}}$}}"

                rate_filtered_noauth = df_noauth[(df_noauth['db_size'] == db_size)]
                rate_filtered_auth = df_auth[(df_auth['db_size'] == db_size)]

                send_notify = (rate_filtered_noauth['RT_SendNotify'].sum() + rate_filtered_auth['RT_SendNotify'].sum())/6
                recv_getnotify = (rate_filtered_noauth['RT_RecvGetNotified'].sum() + rate_filtered_auth['RT_RecvGetNotified'].sum())/6
                recv_notify = (rate_filtered_noauth['RT_RecvNotify'].sum() + rate_filtered_auth['RT_RecvNotify'].sum())/6
                send_getnotify = (rate_filtered_noauth['RT_SendGetNotified'].sum() + rate_filtered_auth['RT_SendGetNotified'].sum())/6

                send_not_string = [f"\\multirow{{{numrows}}}{{*}}{{\\qty{{{send_notify:.2f}}}{{\\second}}}} & \\multirow{{{numrows}}}{{*}}{{\\qty{{{recv_getnotify:.2f}}}{{\\second}}}}",
                                    " & "," & "," & "," & "," & ",]
                recv_not_string = [f"\\multirow{{{numrows}}}{{*}}{{\\qty{{{recv_notify:.2f}}}{{\\second}}}} & \\multirow{{{numrows}}}{{*}}{{\\qty{{{send_getnotify:.2f}}}{{\\second}}}}",
                                   " & "," & "," & "," & "," & ",]
                ctr = 0
                for rate in unique_rates:
                    # Filter the DataFrames for the current db_size and rate
                    filtered_noauth = df_noauth[(df_noauth['db_size'] == db_size) & (df_noauth['rate'] == rate)]
                    filtered_auth = df_auth[(df_auth['db_size'] == db_size) & (df_auth['rate'] == rate)]
  
                    dataframes = [filtered_noauth, filtered_auth]
                    
                    for sysname, dataframe in zip(sysnames, dataframes):
                        rt_total = dataframe['RT_SendPIR'].iloc[0] + dataframe['RT_RecvPIR'].iloc[0] + send_notify + recv_notify
                        out += (
                            f" & \\{sysname}{{}} & {rate} & "
                            f"\\qty{{{dataframe['RT_SendPIR'].iloc[0]:.2f}}}{{\\second}} & {send_not_string[ctr]} & "
                            f"\\qty{{{dataframe['RT_RecvPIR'].iloc[0]:.2f}}}{{\\second}} & {recv_not_string[ctr]} & "
                            f"\\textbf{{ \\qty{{{rt_total:.2f}}}{{\\second}} }}\\\\\n"
                        )
                        ctr += 1
                    # if rate != unique_rates[-1]:
                    #     out += "\\addlinespace\n"
                f.write(f"{out}\\midrule\n")
            f.write("\\end{tabularx}\n}")
    
            
# Example usage: python sum_columns_to_txt.py input output
if __name__ == "__main__":
   
    input_file = sys.argv[1]
    output_file = sys.argv[2]
    process_bw_csvs(input_file, f"{output_file}_bw.tex")
    process_rt_csvs(input_file, f"{output_file}_rt.tex")