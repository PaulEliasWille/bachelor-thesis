import json
import pandas as pd

def read_file(path: str):
    with open(path, "r") as f:
        return f.read()

def read_json_file(path: str):
    content = read_file(path)
    return json.loads(content)

def get_json_array_diff(old_path: str, new_path: str):
    old_array = read_json_file(old_path)
    new_array = read_json_file(new_path)

    added_elements = []
    for new_element in new_array:
        if new_element not in old_array:
            added_elements.append(new_element)

    return added_elements

def sample_csv(in_path, out_path, num_samples):
    df = pd.read_csv(in_path)
    df = df.sample(n=num_samples)
    df.to_csv(out_path, index=False)

def draw_charts():
    pass

def main():
    # sample_csv("data/highlyRelevantRepositories.csv", "data/highlyRelevantRepositoriesSample.csv", 50)
    # added_elements = get_json_array_diff("data/relevantRepositoryIds.json", "data/highlyRelevantRepositoryIds.json")
    # print(json.dumps(added_elements))
    pass

if __name__ == "__main__":
    main()