import os
import json

def calc(data):
    # Initialize the total number of commits and count of weeks
    total_commits = 0
    num_weeks = 0

    # Iterate over each week's data in the provided JSON data
    for week_data in data:
        # Ensure the data includes 'total'
        if 'total' in week_data:
            total_commits += week_data['total']
            num_weeks += 1

    # Calculate the average if num_weeks is greater than zero
    if num_weeks > 0:
        average_commits_per_week = total_commits / num_weeks
        print(f"Average commits per week: {average_commits_per_week}")
    else:
        print("No data available to calculate average commits.")


directory = 'json'

# Iterate over all files in the directory
for root, dirs, files in os.walk(directory):
    for file in files:
        # Check if the file is a JSON file
        if file.endswith('.json'):
            # Construct the full file path
            file_path = os.path.join(root, file)
            # Open and parse the JSON file
            with open(file_path, 'r', encoding='utf-8') as json_file:
                try:
                    data = json.load(json_file)
                    print(file_path)
                    # Perform operations with the data
                    calc(data)
                except json.JSONDecodeError:
                    print(f"Error decoding JSON from file {file_path}")