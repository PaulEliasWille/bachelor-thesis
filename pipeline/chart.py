import json
import copy
import seaborn as sns; sns.set_theme(style="ticks")
import matplotlib.pyplot as plt
import matplotlib.ticker as ticker
import pandas as pd
import numpy as np
import itertools
import functools

def sort_respecting(items, order):
  order_idx_map = {order: idx for idx, order in enumerate(order)}

  def compare_fn(lhs, rhs):
    lhs_order_idx = order_idx_map.get(lhs, None)
    rhs_order_idx = order_idx_map.get(rhs, None)

    return lhs_order_idx - rhs_order_idx
  
  return sorted(items, key=functools.cmp_to_key(compare_fn))

def unique(items):
  unique_items = []
  for item in items:
    if item not in unique_items:
      unique_items.append(item)
  return unique_items

platform_mapping = {
  "aws": "AWS",
  "azure": "Azure",
  "gcp": "GCP",
  "firebase": "GCP",
  "cloudflare": "Cloudflare",
  "vercel": "Vercel",
  "netlify": "Netlify",
  "openwhisk": "OpenWhisk",
  "fastly":  "Fastly",
  "unknown": "Unknown",
}

framework_mapping = {
  "aws_cloudformation_and_sam": "AWS CloudFormation & SAM",
  "serverless": "Serverless",
  "azure_functions": "Azure Functions",
  "wrangler": "Wrangler",
  "firebase": "Firebase",
  "terraform": "Terraform",
  "aws_cdk_and_sst": "AWS CDK & SST",
  "vercel": "Vercel",
  "azure_durable_functions": "Azure Durable Functions",
  "azure_resource_manager": "Azure Resource Manager",
  "architect": "Architect",
  "netlify": "Netlify",
  "hono": "Hono",
  "openwhisk": "OpenWhisk",
  "alexa_skills_kit": "Alex Skills Kit",
  "fastly": "Fastly",
  "gcp_functions": "GCP Functions",
}

framework_category_mapping = {
  "aws_cloudformation_and_sam": "IaC",
  "serverless": "Third-party",
  "azure_functions": "General-purpose Cloud Provider",
  "wrangler": "Specialized Cloud Provider",
  "firebase": "Specialized Cloud Provider",
  "terraform": "IaC",
  "aws_cdk_and_sst": "IaC",
  "vercel": "Specialized Cloud Provider",
  "azure_durable_functions": "General-purpose Cloud Provider",
  "azure_resource_manager": "IaC",
  "architect": "Third-party",
  "netlify": "Specialized Cloud Provider",
  "hono": "Third-party",
  "openwhisk": "Self-hosted",
  "alexa_skills_kit": "Third-party",
  "fastly": "Specialized Cloud Provider",
  "gcp_functions": "General-purpose Cloud Provider",
}

trigger_type_mapping = {
  "http": "HTTP",
  "other": "Event",
  "queue": "Queue",
  "schedule": "Schedule",
  "topic": "Topic",
  "unknown": "Unknown",
}

execution_location_mapping = {
  "region": "Region",
  "edge": "Edge",
  "unknown": "Unknown",
}

#############
### Files ###
#############

def plot_num_files_per_application(repos, out_dir):
  num_repos_by_file_count = {
    "1-10": 0,
    "11-25": 0,
    "26-50": 0,
    "51-100": 0,
    "101-200": 0,
    "201-500": 0,
    "501-1000": 0,
    "1001-10000": 0,
  }
  num_repos = len(repos)

  for repo in repos:
    complexity = repo.get("Complexity", {})
    files = complexity.get("Files", [])
    num_files = len(files)

    if num_files <= 10:
      num_repos_by_file_count["1-10"] += 1
    elif num_files <= 25:
      num_repos_by_file_count["11-25"] += 1
    elif num_files <= 50:
      num_repos_by_file_count["26-50"] += 1
    elif num_files <= 100:
      num_repos_by_file_count["51-100"] += 1
    elif num_files <= 200:
      num_repos_by_file_count["101-200"] += 1
    elif num_files <= 500:
      num_repos_by_file_count["201-500"] += 1
    elif num_files <= 1000:
      num_repos_by_file_count["501-1000"] += 1
    else:
      num_repos_by_file_count["1001-10000"] += 1

  data = {
    "NumFiles": num_repos_by_file_count.keys(),
    "Percentage": [x / num_repos for x in num_repos_by_file_count.values()],
  }

  df = pd.DataFrame(data)

  fig, ax = plt.subplots(figsize=(10, 6))
  sns.barplot(x="Percentage", y="NumFiles", data=df, orient="h", ax=ax)

  ax.xaxis.set_major_locator(ticker.MultipleLocator(0.05))

  for index, value in enumerate(df["Percentage"]):
    plt.text(value + 0.0025, index, f'{value:.2%}', va="center")

  plt.xlim(0, max(df["Percentage"]) * 1.1)

  plt.xlabel("Percentage of Applications")
  plt.ylabel("Number of Files")

  ax.xaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: ""))
  ax.xaxis.set_minor_formatter(ticker.NullFormatter())

  plt.savefig(f"{out_dir}/num_files_per_application.png", transparent=True, bbox_inches='tight')
  plt.close()

def plot_number_of_files_repository_distribution_v1(repos, out_dir):
  total_num_repos = len(repos)
  total_num_files = 0

  num_files_list = []
  for repo in repos:
    complexity = repo.get("Complexity", {})
    files = complexity.get("Files", [])
    num_files_list.append(len(files))
    total_num_files += len(files)

  num_files_list = sorted(num_files_list)

  x = [0]
  y = [0]
  
  agg_num_repos = 0
  for num_files in num_files_list:
    agg_num_repos += 1

    x.append(num_files)
    y.append(agg_num_repos / total_num_repos)

  data = {
    "NumFiles": x,
    "Percentage": y,
  }

  df = pd.DataFrame(data)

  fig, ax = plt.subplots(figsize=(6, 3))
  sns.lineplot(x="NumFiles", y="Percentage", data=df, ax=ax)
  ax.set_xscale("log")

  plt.xlabel("Number of Files")
  plt.ylabel("Cml Fraction of Repositories [%]")

  ax.yaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: f"{x*100:.0f}"))
  ax.yaxis.set_minor_formatter(ticker.NullFormatter())

  # plt.xlim(0, max(df["Percentage"]) * 1.1)
  plt.xlim(min(num_files_list), max(num_files_list))
  plt.ylim(0.0, 1.05)

  # ax.xaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: ""))
  # ax.xaxis.set_minor_formatter(ticker.NullFormatter())

  plt.savefig(f"{out_dir}/number_of_files_repository_distribution_v1.png", transparent=True, bbox_inches='tight')
  plt.close()

def plot_number_of_files_repository_distribution_v2(repos, out_dir):
  file_category_percentages_by_category = {
    "Source Code": [],
    "Documentation": [],
    "Data": [],
    "Asset": [],
    "Other": []
  }

  num_files_list = []
  for repo in repos:
    complexity = repo.get("Complexity", {})
    files = complexity.get("Files", [])

    num_files = len(files)
    num_files_by_category = {
      "Source Code": 0,
      "Documentation": 0,
      "Data": 0,
      "Asset": 0,
      "Other": 0
    }

    for file in files:
      category = file["Category"]
      if category not in num_files_by_category:
        continue
      num_files_by_category[category] += 1

    for category, category_num_files in num_files_by_category.items():
      if category not in file_category_percentages_by_category:
        continue
      file_category_percentages_by_category[category].append(category_num_files / num_files)

    num_files_list.append(num_files)

  x = []
  y = []
  z = []

  for num_files, asset_percentage, data_percentage, documentation_percentage, other_percentage, source_code_percentage in zip(
    num_files_list, file_category_percentages_by_category["Asset"],
    file_category_percentages_by_category["Data"], file_category_percentages_by_category["Documentation"],
    file_category_percentages_by_category["Other"], file_category_percentages_by_category["Source Code"]):
    x.append(num_files)
    y.append("Asset")
    z.append(asset_percentage)

    x.append(num_files)
    y.append("Data")
    z.append(data_percentage)

    x.append(num_files)
    y.append("Documentation")
    z.append(documentation_percentage)

    x.append(num_files)
    y.append("Other")
    z.append(other_percentage)

    x.append(num_files)
    y.append("Source Code")
    z.append(source_code_percentage)

  data = {
    "NumFiles": x,
    "Category": y,
    "Percentage": z,
  }

  df = pd.DataFrame(data)

  fig, ax = plt.subplots(figsize=(6, 3))
  sns.swarmplot(data=df, x="NumFiles", y="Category", hue="Percentage")
  #sns.lineplot(x="NumFiles", y="Percentage", data=df, ax=ax)
  ax.set_xscale("log")

  plt.xlabel("Number of Files")
  #plt.ylabel("Cml Fraction of Repositories [%]")

  #ax.yaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: f"{x*100:.0f}"))
  #ax.yaxis.set_minor_formatter(ticker.NullFormatter())

  #plt.xlim(0, max(df["Percentage"]) * 1.1)
  #plt.xlim(min(x), max(x))
  #plt.ylim(0.0, 1.05)

  # ax.xaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: ""))
  # ax.xaxis.set_minor_formatter(ticker.NullFormatter())

  plt.savefig(f"{out_dir}/number_of_files_repository_distribution_v2.png", transparent=True, bbox_inches='tight')
  plt.close()

def plot_number_of_files_repository_distribution_v3(repos, out_dir):
  num_files_list = []
  for repo in repos:
    complexity = repo.get("Complexity", {})
    files = complexity.get("Files", [])
    num_files_list.append(len(files))

  data = {
    "NumFiles": num_files_list,
  }

  df = pd.DataFrame(data)

  fig, ax = plt.subplots(figsize=(6, 3))
  sns.violinplot(x="NumFiles", data=df, ax=ax, log_scale=True, inner=None, alpha=0.6)
  # for Matplotlib version >= 1.5
  plt.gca().set_prop_cycle(None)
  sns.stripplot(data=data, x="NumFiles", log_scale=True, alpha=0.8, edgecolor="w", linewidth=0.5)
  #ax.set_xscale("log")

  plt.xlabel("Number of Files")
  #plt.ylabel("Cml Fraction of Repositories [%]")

  #ax.yaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: f"{x*100:.0f}"))
  #ax.yaxis.set_minor_formatter(ticker.NullFormatter())

  # plt.xlim(0, max(df["Percentage"]) * 1.1)
  #plt.xlim(min(num_files_list), max(num_files_list))
  #plt.ylim(0.0, 1.05)

  # ax.xaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: ""))
  # ax.xaxis.set_minor_formatter(ticker.NullFormatter())

  plt.savefig(f"{out_dir}/number_of_files_repository_distribution_v3.png", transparent=True, bbox_inches='tight')
  plt.close()

def plot_number_of_files_category_distribution(repos, out_dir):
  category_order = ["Source Code", "Documentation", "Data", "Asset", "Other"][::-1]

  data_list = []
  for repo in repos:
    complexity = repo.get("Complexity", {})
    files = complexity.get("Files", [])
    
    num_files_by_category = {k: 0 for k in category_order}
    for file in files:
      category = file["Category"]
      num_files_by_category[category] += 1
    
    data_list.append({
      "num_files_by_category": num_files_by_category,
      "num_files": len(files)
    })

  data_list = sorted(data_list, key=lambda x: x["num_files"])

  x = []
  y = [[] for _ in category_order]

  for data in data_list:
    x.append(data["num_files"])

    for idx, category in enumerate(category_order):
      if data["num_files"] != 0:
        y[idx].append(data["num_files_by_category"][category] / data["num_files"])
      else:
        y[idx].append(0)
      
  data = {
    "x": x,
  }

  num_areas = 0
  for idx, category in enumerate(category_order):
    data[category] = y[idx]
    if sum(y[idx]) > 0:
      num_areas += 1

  df = pd.DataFrame(data)

  # Apply a moving average with a window size of your choice
  window_size = 20
  smoothed_df = df.rolling(window=window_size, center=True).mean()

  current_palette = sns.color_palette()

  fig, ax = plt.subplots(figsize=(6, 3))
  plt.stackplot(smoothed_df['x'], smoothed_df.drop(columns='x').T, labels=smoothed_df.columns[1:], colors=current_palette[:num_areas][::-1])
   
  ax.set_xscale("log")

  plt.xlabel("Number of Files")
  plt.ylabel("Distribution of File Categories [%]")

  ax.yaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: f"{x*100:.0f}"))
  ax.yaxis.set_minor_formatter(ticker.NullFormatter())

  # plt.xlim(0, max(df["Percentage"]) * 1.1)
  plt.xlim(smoothed_df["x"].min(), smoothed_df["x"].max())
  plt.ylim(0.0, 1.0)

  plt.legend()

  handles, labels = ax.get_legend_handles_labels()
  ax.legend(handles[::-1], labels[::-1], title='Line', loc='upper left')

  sns.move_legend(
    ax, "upper left",
    bbox_to_anchor=(1, 1), ncol=1,
    title=None, frameon=False,
  )

  # ax.xaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: ""))
  # ax.xaxis.set_minor_formatter(ticker.NullFormatter())

  plt.savefig(f"{out_dir}/number_of_files_category_distribution.png", transparent=True, bbox_inches='tight')
  plt.close()

def plot_distribution_of_files_by_category_by_application_size(repos, out_dir):
  file_count_by_repo_size = {
    "1-10": 0,
    "11-25": 0,
    "26-50": 0,
    "51-100": 0,
    "101-200": 0,
    "201-500": 0,
    "501-1000": 0,
    "1001-10000": 0,
  }

  file_count_by_category = {
    "Source Code": 0,
    "Asset": 0,
    "Other": 0,
    "Data": 0,    
    "Documentation": 0,
  }

  file_count_by_category_by_repo_size = {
    "1-10": copy.deepcopy(file_count_by_category),
    "11-25": copy.deepcopy(file_count_by_category),
    "26-50": copy.deepcopy(file_count_by_category),
    "51-100": copy.deepcopy(file_count_by_category),
    "101-200": copy.deepcopy(file_count_by_category),
    "201-500": copy.deepcopy(file_count_by_category),
    "501-1000": copy.deepcopy(file_count_by_category),
    "1001-10000": copy.deepcopy(file_count_by_category),
  }

  file_count_by_repo_size_by_category = {
    "Source Code": copy.deepcopy(file_count_by_repo_size),
    "Asset": copy.deepcopy(file_count_by_repo_size),
    "Other": copy.deepcopy(file_count_by_repo_size),
    "Data": copy.deepcopy(file_count_by_repo_size),    
    "Documentation": copy.deepcopy(file_count_by_repo_size),    
  }

  for repo in repos:
    complexity = repo.get("Complexity", {})
    files = complexity.get("Files", {})
    repo_file_count = len(files)
    for file in files:
      category = file["Category"]

      file_count_by_category[category] += 1

      if repo_file_count <= 10:
        file_count_by_repo_size["1-10"] += 1
        file_count_by_category_by_repo_size["1-10"][category] += 1
        file_count_by_repo_size_by_category[category]["1-10"] += 1
      elif repo_file_count <= 25:
        file_count_by_repo_size["11-25"] += 1
        file_count_by_category_by_repo_size["11-25"][category] += 1
        file_count_by_repo_size_by_category[category]["11-25"] += 1
      elif repo_file_count <= 50:
        file_count_by_repo_size["26-50"] += 1
        file_count_by_category_by_repo_size["26-50"][category] += 1
        file_count_by_repo_size_by_category[category]["26-50"] += 1
      elif repo_file_count <= 100:
        file_count_by_repo_size["51-100"] += 1
        file_count_by_category_by_repo_size["51-100"][category] += 1
        file_count_by_repo_size_by_category[category]["51-100"] += 1
      elif repo_file_count <= 200:
        file_count_by_repo_size["101-200"] += 1
        file_count_by_category_by_repo_size["101-200"][category] += 1
        file_count_by_repo_size_by_category[category]["101-200"] += 1
      elif repo_file_count <= 500:
        file_count_by_repo_size["201-500"] += 1
        file_count_by_category_by_repo_size["201-500"][category] += 1
        file_count_by_repo_size_by_category[category]["201-500"] += 1
      elif repo_file_count <= 1000:
        file_count_by_repo_size["501-1000"] += 1
        file_count_by_category_by_repo_size["501-1000"][category] += 1
        file_count_by_repo_size_by_category[category]["501-1000"] += 1
      else:
        file_count_by_repo_size["1001-10000"] += 1
        file_count_by_category_by_repo_size["1001-10000"][category] += 1
        file_count_by_repo_size_by_category[category]["1001-10000"] += 1

  fig, ax = plt.subplots(figsize=(10, 6))

  offset_per_repo_size = {x: 1.0 for x in file_count_by_repo_size.keys()}

  for category, category_file_count_by_repo_size in list(file_count_by_repo_size_by_category.items())[::-1]:
    df = pd.DataFrame({
      "RepoSize": file_count_by_repo_size.keys(),
      "Percentage": [offset_per_repo_size[repo_size] for repo_size in file_count_by_repo_size.keys()],
    })

    sns.barplot(x="Percentage", y="RepoSize", data=df, orient="h", ax=ax, label=category)

    for repo_size in file_count_by_repo_size.keys():
      if category_file_count_by_repo_size[repo_size] > 0:
        #print(f"{category} + {repo_size} -> {category_file_count_by_repo_size[repo_size] / file_count_by_repo_size[repo_size]}")
        offset_per_repo_size[repo_size] -= category_file_count_by_repo_size[repo_size] / file_count_by_repo_size[repo_size]
      #else:
        #print(f"{category} + {repo_size} -> 0")

  handles, labels = ax.get_legend_handles_labels()
  ax.legend(handles[::-1], labels[::-1], title='Line', loc='upper left')
  # sns.move_legend(ax, "upper left", bbox_to_anchor=(0, -0.3))

  sns.move_legend(
    ax, "lower left",
    bbox_to_anchor=(0, 1), ncol=5,
    title=None, frameon=False,
  )

  ax.xaxis.set_major_locator(ticker.MultipleLocator(0.2))

  plt.xlabel("Distribution [%] of Files by Category")
  plt.ylabel("Number of Files in Repository")

  plt.xlim(0, 1)

  ax.xaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: f"{x*100:.0f}"))
  ax.xaxis.set_minor_formatter(ticker.NullFormatter())

  # plt.show()
  plt.savefig(f"{out_dir}/distribution_of_files_by_category_by_application_size.png", transparent=True, bbox_inches="tight", pad_inches=1)
  plt.close()

#####################
### Lines of Code ###
#####################

def plot_loc_repository_distribution_v1(repos, out_dir):
  total_num_repos = len(repos)
  total_loc = 0

  loc_list = []
  for repo in repos:
    complexity = repo.get("Complexity", {})
    files = complexity.get("Files", [])
    repo_loc = 0
    for file in files:
      category = file["Category"]
      if category != "Source Code":
        continue

      repo_loc += file["LOC"]
      total_loc += file["LOC"]

    loc_list.append(repo_loc)

  loc_list = sorted(loc_list)

  x = [0]
  y = [0]
  
  agg_num_repos = 0
  for loc in loc_list:
    agg_num_repos += 1

    x.append(loc)
    y.append(agg_num_repos / total_num_repos)

  data = {
    "LoC": x,
    "Percentage": y,
  }

  df = pd.DataFrame(data)

  fig, ax = plt.subplots(figsize=(6, 3))
  sns.lineplot(x="LoC", y="Percentage", data=df, ax=ax)
  ax.set_xscale("log")

  plt.xlabel("LoC")
  plt.ylabel("Cml Fraction of Repositories [%]")

  ax.yaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: f"{x*100:.0f}"))
  ax.yaxis.set_minor_formatter(ticker.NullFormatter())

  # plt.xlim(0, max(df["Percentage"]) * 1.1)
  plt.xlim(min(loc_list), max(loc_list))
  plt.ylim(0.0, 1.05)

  # ax.xaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: ""))
  # ax.xaxis.set_minor_formatter(ticker.NullFormatter())

  plt.savefig(f"{out_dir}/loc_repository_distribution_v1.png", transparent=True, bbox_inches='tight')
  plt.close()

def plot_loc_repository_distribution_v2(repos, out_dir):
  total_loc = 0
  loc_list = []
  for repo in repos:
    complexity = repo.get("Complexity", {})
    files = complexity.get("Files", [])
    repo_loc = 0
    for file in files:
      category = file["Category"]
      if category != "Source Code":
        continue

      repo_loc += file["LOC"]
      total_loc += file["LOC"]

    loc_list.append(repo_loc)

  data = {
    "LoC": loc_list,
  }

  df = pd.DataFrame(data)

  fig, ax = plt.subplots(figsize=(6, 3))
  sns.violinplot(x="LoC", data=df, ax=ax, log_scale=True, inner=None, alpha=0.6)
  # for Matplotlib version >= 1.5
  plt.gca().set_prop_cycle(None)
  sns.stripplot(data=data, x="LoC", log_scale=True, alpha=0.8, edgecolor="w", linewidth=0.5)
  #ax.set_xscale("log")

  plt.xlabel("LoC")
  #plt.ylabel("Cml Fraction of Repositories [%]")

  #ax.yaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: f"{x*100:.0f}"))
  #ax.yaxis.set_minor_formatter(ticker.NullFormatter())

  # plt.xlim(0, max(df["Percentage"]) * 1.1)
  #plt.xlim(min(num_files_list), max(num_files_list))
  #plt.ylim(0.0, 1.05)

  # ax.xaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: ""))
  # ax.xaxis.set_minor_formatter(ticker.NullFormatter())

  plt.savefig(f"{out_dir}/loc_repository_distribution_v2.png", transparent=True, bbox_inches='tight')
  plt.close()

def plot_loc_language_distribution(repos, out_dir):
  language_mapping = {
    "JavaScript": "JavaScript",
    "JSX": "JSX",
    "TypeScript": "TypeScript",
    "HTML": "HTML, CSS & Sass",
    "CSS": "HTML, CSS & Sass",
    "Sass": "HTML, CSS & Sass"
  }
  language_order = ["JavaScript", "TypeScript", "JSX", "HTML, CSS & Sass", "Other"][::-1]

  data_list = []
  for repo in repos:
    complexity = repo.get("Complexity", {})
    files = complexity.get("Files", [])
    
    repo_loc = 0
    loc_by_language = {k: 0 for k in language_order}
    for file in files:
      category = file["Category"]
      if category != "Source Code":
        continue

      language = file["Language"]
      language = language_mapping.get(language, "Other")
      loc_by_language[language] += file["LOC"]
      repo_loc += file["LOC"]
    
    data_list.append({
      "loc_by_language": loc_by_language,
      "loc": repo_loc
    })

  data_list = sorted(data_list, key=lambda x: x["loc"])

  x = []
  y = [[] for _ in language_order]

  for data in data_list:
    x.append(data["loc"])

    for idx, language in enumerate(language_order):
      if data["loc"] != 0:
        y[idx].append(data["loc_by_language"][language] / data["loc"])
      else:
        y[idx].append(0)
      
  data = {
    "x": x,
  }

  num_areas = 0
  for idx, language in enumerate(language_order):
    data[language] = y[idx]
    if sum(y[idx]) > 0:
      num_areas += 1

  df = pd.DataFrame(data)

  # Apply a moving average with a window size of your choice
  window_size = 15
  smoothed_df = df.rolling(window=window_size, center=True).mean()

  current_palette = sns.color_palette()

  fig, ax = plt.subplots(figsize=(6, 3))
  plt.stackplot(smoothed_df['x'], smoothed_df.drop(columns='x').T, labels=smoothed_df.columns[1:], colors=current_palette[:num_areas][::-1])
  
  ax.set_xscale("log")

  plt.xlabel("LoC")
  plt.ylabel("Distribution of Languages [%]")

  ax.yaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: f"{x*100:.0f}"))
  ax.yaxis.set_minor_formatter(ticker.NullFormatter())

  # plt.xlim(0, max(df["Percentage"]) * 1.1)
  plt.xlim(smoothed_df["x"].min(), smoothed_df["x"].max())
  plt.ylim(0.0, 1.0)

  plt.legend()

  handles, labels = ax.get_legend_handles_labels()
  ax.legend(handles[::-1], labels[::-1], title='Line', loc='upper left')

  sns.move_legend(
    ax, "upper left",
    bbox_to_anchor=(1, 1), ncol=1,
    title=None, frameon=False,
  )

  # ax.xaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: ""))
  # ax.xaxis.set_minor_formatter(ticker.NullFormatter())

  plt.savefig(f"{out_dir}/loc_language_distribution.png", transparent=True, bbox_inches='tight')
  plt.close()

def plot_loc_per_application(repos, out_dir):
  loc_count = {
    "0-50": 0,
    "51-100": 0,
    "101-200": 0,
    "201-500": 0,
    "501-1000": 0,
    "1001-2000": 0,
    "2001-5000": 0,
    "5001-10000": 0,
    "10001-20000": 0,
    "20001-50000": 0,
    "50001+": 0,
  }
  num_repos = len(repos)

  for repo in repos:
    complexity = repo.get("Complexity", {})
    files = complexity.get("Files", [])
    
    total_loc = 0
    for file in files:
      if file["Category"] != "Source Code":
        continue
      if file["Language"] != "JavaScript":
        continue
      total_loc += file["LOC"]

    if total_loc <= 50:
      loc_count["0-50"] += 1
    elif total_loc <= 100:
      loc_count["51-100"] += 1
    elif total_loc <= 200:
      loc_count["101-200"] += 1
    elif total_loc <= 500:
      loc_count["201-500"] += 1
    elif total_loc <= 1000:
      loc_count["501-1000"] += 1
    elif total_loc <= 2000:
      loc_count["1001-2000"] += 1
    elif total_loc <= 5000:
      loc_count["2001-5000"] += 1
    elif total_loc <= 10000:
      loc_count["5001-10000"] += 1
    elif total_loc <= 20000:
      loc_count["10001-20000"] += 1
    elif total_loc <= 50000:
      loc_count["20001-50000"] += 1
    else:
      loc_count["50001+"] += 1

  data = {
    "LoC": loc_count.keys(),
    "Percentage": [x / num_repos for x in loc_count.values()],
  }

  df = pd.DataFrame(data)

  fig, ax = plt.subplots(figsize=(10, 6))
  sns.barplot(x="Percentage", y="LoC", data=df, orient="h", ax=ax)

  ax.xaxis.set_major_locator(ticker.MultipleLocator(0.05))

  for index, value in enumerate(df["Percentage"]):
    plt.text(value + 0.0025, index, f'{value:.2%}', va="center")

  plt.xlim(0, max(df["Percentage"]) * 1.1)

  plt.xlabel("Percentage of Applications")
  plt.ylabel("Lines of Code")

  ax.xaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: ""))
  ax.xaxis.set_minor_formatter(ticker.NullFormatter())

  plt.savefig(f"{out_dir}/lines_of_code_per_application.png", transparent=True, bbox_inches='tight')
  plt.close()

#################
### Languages ###
#################

def plot_languages_by_loc_files_and_applications(repos, out_dir):
  language_count_by_application = {}
  language_count_by_loc = {}
  language_count_by_files = {}

  num_repos = len(repos)
  num_files = 0
  num_loc = 0

  for repo in repos:
    repo_languages = {}

    complexity = repo.get("Complexity", {})
    files = complexity.get("Files", [])
    for file in files:  
      if file["Category"] != "Source Code":
        continue
      repo_languages[file["Language"]] = True

    for language in repo_languages.keys():
      if language not in language_count_by_application:
        language_count_by_application[language] = 0

      language_count_by_application[language] += 1

  for repo in repos:
    complexity = repo.get("Complexity", {})
    files = complexity.get("Files", [])
    for file in files:  
      language = file["Language"]
      category = file["Category"]

      if category != "Source Code":
        continue

      if language not in language_count_by_files:
        language_count_by_files[language] = 0

      language_count_by_files[language] += 1
      num_files += 1

  for repo in repos:
    complexity = repo.get("Complexity", {})
    files = complexity.get("Files", [])
    for file in files:  
      language = file["Language"]
      category = file["Category"]
      loc = file["LOC"]

      if category != "Source Code":
        continue

      if language not in language_count_by_loc:
        language_count_by_loc[language] = 0

      language_count_by_loc[language] += loc
      num_loc += loc

  languages = []
  percentage = []
  criteria = []

  languages.extend(language_count_by_loc.keys())
  percentage.extend([x / num_loc for x in language_count_by_loc.values()])
  criteria.extend(["LoC"] *  len(language_count_by_loc.keys()))

  languages.extend(language_count_by_files.keys())
  percentage.extend([x / num_files for x in language_count_by_files.values()])
  criteria.extend(["Files"] * len(language_count_by_files.keys()))

  languages.extend(language_count_by_application.keys())
  percentage.extend([x / num_repos for x in language_count_by_application.values()])
  criteria.extend(["Applications"] * len(language_count_by_application.keys()))

  data = {
    "Language": languages,
    "Percentage": percentage,
    "Criteria": criteria,
  }

  df = pd.DataFrame(data)
  df = df.sort_values("Percentage", ascending=False)

  unique_languages = set()
  for index, row in df.iterrows():
    if len(unique_languages) < 10:
      unique_languages.add(row["Language"])

  df = df[df['Language'].isin(list(unique_languages))]

  fig, ax = plt.subplots(figsize=(10, 6))
  sns.barplot(x="Percentage", y="Language", hue="Criteria", data=df, orient="h", ax=ax)

  ax.xaxis.set_major_locator(ticker.MultipleLocator(0.20))

  plt.xlim(0, 1)

  plt.xlabel("Percentage of")
  plt.ylabel("Languages")

  ax.xaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: f"{x*100:.0f}%"))
  ax.xaxis.set_minor_formatter(ticker.NullFormatter())

  plt.savefig(f"{out_dir}/languages_by_loc_files_and_applications.png", transparent=True, bbox_inches='tight')
  plt.close()

###########################
### Number of Functions ###
###########################

def plot_javascript_loc_and_num_functions(repos, out_dir):
  data = { "LoC": [], "NumFunctions": [], "Language": [] }
  unique_languages = set()

  for repo in repos:
    complexity = repo.get("Complexity", {})
    files = complexity.get("Files", [])
    num_functions = repo.get("NumFunctions", 0)
    
    repo_loc_by_language = {}
    for file in files:
      category = file["Category"]
      language = file["Language"]
      loc = file["LOC"]

      if category not in ["Source Code"]:
        continue

      if language not in ["JavaScript"]:
        continue

      if language not in repo_loc_by_language:
        repo_loc_by_language[language] = 0
        
      repo_loc_by_language[language] += loc

    for language, loc in repo_loc_by_language.items():
      unique_languages.add(language)

      data["LoC"].append(loc)
      data["NumFunctions"].append(num_functions)
      data["Language"].append(language)

  df = pd.DataFrame(data)

  language_mapping = {}
  
  for language in unique_languages:
    df_language = df[df["Language"] == language]
    df_language = df_language.drop(["Language"], axis=1)
    corr = df_language['LoC'].corr(df_language['NumFunctions'])

    language_mapping[language] = f"{language} (Datapoints: {len(df_language):.0f} | Corr: {corr:.2f})"

  for idx, language in enumerate(data["Language"]):
    data["Language"][idx] = language_mapping.get(language, language)

  df = pd.DataFrame(data)

  fig, ax = plt.subplots(figsize=(6, 3))
  sns.scatterplot(
    x="LoC",
    y="NumFunctions",
    data=df,
    hue="Language",
    legend=False,
    alpha=0.8,
  )
  ax.set_xscale("log")

  plt.xlim(30, 130_000)
  plt.yticks([0, 5, 10, 20, 40, 60])

  plt.xlabel("JavaScript LoC")
  plt.ylabel("Number of Functions")

  plt.savefig(f"{out_dir}/javascript_loc_and_num_functions.png", transparent=True, bbox_inches='tight')
  plt.close()

def plot_num_functions_per_framework(repos, out_dir):
  num_functions_list = []
  frameworks_list = []
  shared_list = []
  for repo in repos:
    num_functions_by_framework = {}
    
    functions = repo.get("Functions", [])
    for function in functions:
      framework = function["Framework"]
      framework = framework_mapping.get(framework, framework)
      if framework not in num_functions_by_framework:
        num_functions_by_framework[framework] = 0
      num_functions_by_framework[framework] += 1

    for k, v in num_functions_by_framework.items():
      frameworks_list.append(k)
      num_functions_list.append(v)
      shared_list.append("Shared" if len(num_functions_by_framework) > 1 else "Exclusive")

  data = {
    "NumFunctions": num_functions_list,
    "Frameworks": frameworks_list,
    "Shared": shared_list,
  }

  order = [framework_mapping.get(x, x) for x in [
    "aws_cloudformation_and_sam",# IaC
    "terraform",# IaC
    "aws_cdk_and_sst",# IaC
    "azure_resource_manager",# IaC
    "azure_functions",# General-purpose Cloud Provider
    "azure_durable_functions",# General-purpose Cloud Provider
    "gcp_functions",# General-purpose Cloud Provider
    "wrangler",# Specialized Cloud Provider
    "firebase",# Specialized Cloud Provider
    "vercel",# Specialized Cloud Provider
    "netlify",# Specialized Cloud Provider
    "fastly",# Specialized Cloud Provider
    "serverless",# Third-party
    "architect",# Third-party
    "hono",# Third-party
    "alexa_skills_kit",# Third-party
    "openwhisk",# Self-hosted
  ]]


  df = pd.DataFrame(data)

  fig, ax = plt.subplots(figsize=(8, 4))
  #sns.violinplot(x="NumFunctions", y="Categories", data=df, ax=ax, log_scale=False, inner=None, alpha=0.6)
  # for Matplotlib version >= 1.5
  #plt.gca().set_prop_cycle(None)
  sns.stripplot(
    data=data,
    x="NumFunctions", y="Frameworks", hue="Shared",
    jitter=0.2, log_scale=False, alpha=0.8,
    edgecolor="w", linewidth=0.5, order=order,
  )
  #ax.set_xscale("log")

  sns.move_legend(
    ax, "lower left",
    bbox_to_anchor=(0, 1),
    ncol=2,
    title=None,
    frameon=False,
  )

  plt.xlim(0.5, 48)
  plt.xticks([1, 5, 10, 20, 30, 40, 50])

  plt.xlabel("Number of Functions")
  plt.ylabel("")
  #plt.ylabel("Cml Fraction of Repositories [%]")

  #ax.yaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: f"{x*100:.0f}"))
  #ax.yaxis.set_minor_formatter(ticker.NullFormatter())

  # plt.xlim(0, max(df["Percentage"]) * 1.1)
  #plt.xlim(min(num_files_list), max(num_files_list))
  #plt.ylim(0.0, 1.05)

  # ax.xaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: ""))
  # ax.xaxis.set_minor_formatter(ticker.NullFormatter())

  plt.savefig(f"{out_dir}/num_functions_per_framework.png", transparent=True, bbox_inches='tight')
  plt.close()

def plot_num_functions_per_framework_category(repos, out_dir):
  num_functions_list = []
  categories_list = []
  shared_list = []
  for repo in repos:
    num_functions_by_category = {}
    
    functions = repo.get("Functions", [])
    for function in functions:
      framework = function["Framework"]
      framework_category = framework_category_mapping.get(framework, framework)
      if framework_category not in num_functions_by_category:
        num_functions_by_category[framework_category] = 0
      num_functions_by_category[framework_category] += 1

    for k, v in num_functions_by_category.items():
      categories_list.append(k)
      num_functions_list.append(v)
      shared_list.append("Shared" if len(num_functions_by_category) > 1 else "Exclusive")

  data = {
    "NumFunctions": num_functions_list,
    "Categories": categories_list,
    "Shared": shared_list,
  }

  df = pd.DataFrame(data)

  fig, ax = plt.subplots(figsize=(6, 3))
  #sns.violinplot(x="NumFunctions", y="Categories", data=df, ax=ax, log_scale=False, inner=None, alpha=0.6)
  # for Matplotlib version >= 1.5
  #plt.gca().set_prop_cycle(None)
  sns.stripplot(
    data=data,
    x="NumFunctions", y="Categories", hue="Shared",
    jitter=0.2, log_scale=False, alpha=0.8,
    edgecolor="w", linewidth=0.5, order=unique(framework_category_mapping.values())
  )
  #ax.set_xscale("log")

  sns.move_legend(
    ax, "lower left",
    bbox_to_anchor=(0, 1),
    ncol=2,
    title=None,
    frameon=False,
  )

  plt.xlim(0.5, 48)
  plt.xticks([1, 5, 10, 20, 30, 40, 50])

  plt.xlabel("Number of Functions")
  plt.ylabel("")
  #plt.ylabel("Cml Fraction of Repositories [%]")

  #ax.yaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: f"{x*100:.0f}"))
  #ax.yaxis.set_minor_formatter(ticker.NullFormatter())

  # plt.xlim(0, max(df["Percentage"]) * 1.1)
  #plt.xlim(min(num_files_list), max(num_files_list))
  #plt.ylim(0.0, 1.05)

  # ax.xaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: ""))
  # ax.xaxis.set_minor_formatter(ticker.NullFormatter())

  plt.savefig(f"{out_dir}/num_functions_per_framework_category.png", transparent=True, bbox_inches='tight')
  plt.close()

def plot_num_functions_per_platform(repos, out_dir):
  num_functions_list = []
  platforms_list = []
  shared_list = []
  for repo in repos:
    num_functions_by_platform = {}
    
    functions = repo.get("Functions", [])
    for function in functions:
      platform = function["Platform"]
      platform = platform_mapping.get(platform, platform)
      if platform not in num_functions_by_platform:
        num_functions_by_platform[platform] = 0
      num_functions_by_platform[platform] += 1

    for k, v in num_functions_by_platform.items():
      platforms_list.append(k)
      num_functions_list.append(v)

      shared_list.append("Shared" if len(num_functions_by_platform) > 1 else "Exclusive")

  data = {
    "NumFunctions": num_functions_list,
    "Platforms": platforms_list,
    "Shared": shared_list,
  }

  df = pd.DataFrame(data)

  fig, ax = plt.subplots(figsize=(6, 3))
  #sns.violinplot(x="NumFunctions", y="Categories", data=df, ax=ax, log_scale=False, inner=None, alpha=0.6)
  # for Matplotlib version >= 1.5
  #plt.gca().set_prop_cycle(None)
  sns.stripplot(
    data=data, x="NumFunctions", y="Platforms", hue="Shared",
    jitter=0.2, log_scale=False, alpha=0.8,
    edgecolor="w", linewidth=0.5, order=unique(platform_mapping.values()),
  )
  #ax.set_xscale("log")

  sns.move_legend(
    ax, "lower left",
    bbox_to_anchor=(0, 1),
    ncol=2,
    title=None,
    frameon=False,
  )

  plt.xlim(0.5, 48)
  plt.xticks([1, 5, 10, 20, 30, 40, 50])

  plt.xlabel("Number of Functions")
  plt.ylabel("")

  #ax.yaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: f"{x*100:.0f}"))
  #ax.yaxis.set_minor_formatter(ticker.NullFormatter())

  # plt.xlim(0, max(df["Percentage"]) * 1.1)
  #plt.xlim(min(num_files_list), max(num_files_list))
  #plt.ylim(0.0, 1.05)

  # ax.xaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: ""))
  # ax.xaxis.set_minor_formatter(ticker.NullFormatter())

  plt.savefig(f"{out_dir}/num_functions_per_platform.png", transparent=True, bbox_inches='tight')
  plt.close()

def plot_num_functions_per_execution_location(repos, out_dir):
  num_functions_list = []
  execution_locations_list = []
  shared_list = []
  for repo in repos:
    num_functions_by_execution_location = {}
    
    functions = repo.get("Functions", [])
    for function in functions:
      execution_location = function["Location"]
      execution_location = execution_location_mapping.get(execution_location, execution_location)
      if execution_location not in num_functions_by_execution_location:
        num_functions_by_execution_location[execution_location] = 0
      num_functions_by_execution_location[execution_location] += 1

    for k, v in num_functions_by_execution_location.items():
      execution_locations_list.append(k)
      num_functions_list.append(v)

      shared_list.append("Shared" if len(num_functions_by_execution_location) > 1 else "Exclusive")

  data = {
    "NumFunctions": num_functions_list,
    "ExecutionLocations": execution_locations_list,
    "Shared": shared_list,
  }

  df = pd.DataFrame(data)

  fig, ax = plt.subplots(figsize=(6, 3))
  #sns.violinplot(x="NumFunctions", y="Categories", data=df, ax=ax, log_scale=False, inner=None, alpha=0.6)
  # for Matplotlib version >= 1.5
  #plt.gca().set_prop_cycle(None)
  sns.stripplot(
    data=data, x="NumFunctions", y="ExecutionLocations", hue="Shared",
    jitter=0.1, log_scale=False, alpha=0.8,
    edgecolor="w", linewidth=0.5, order=unique(execution_location_mapping.values()),
  )
  #ax.set_xscale("log")

  sns.move_legend(
    ax, "lower left",
    bbox_to_anchor=(0, 1),
    ncol=2,
    title=None,
    frameon=False,
  )

  plt.xlim(0.5, 48)
  plt.xticks([1, 5, 10, 20, 30, 40, 50])

  plt.xlabel("Number of Functions")
  plt.ylabel("")
  #plt.ylabel("Cml Fraction of Repositories [%]")

  #ax.yaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: f"{x*100:.0f}"))
  #ax.yaxis.set_minor_formatter(ticker.NullFormatter())

  # plt.xlim(0, max(df["Percentage"]) * 1.1)
  #plt.xlim(min(num_files_list), max(num_files_list))
  #plt.ylim(0.0, 1.05)

  # ax.xaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: ""))
  # ax.xaxis.set_minor_formatter(ticker.NullFormatter())

  plt.savefig(f"{out_dir}/num_functions_per_execution_location.png", transparent=True, bbox_inches='tight')
  plt.close()

def plot_num_functions_per_trigger_type(repos, out_dir):
  num_functions_list = []
  trigger_types_list = []
  shared_list = []
  for repo in repos:
    num_functions_by_trigger_type = {}
    
    functions = repo.get("Functions", [])
    for function in functions:
      trigger_type = function["InvocationType"]
      trigger_type = trigger_type_mapping.get(trigger_type, trigger_type)
      if trigger_type not in num_functions_by_trigger_type:
        num_functions_by_trigger_type[trigger_type] = 0
      num_functions_by_trigger_type[trigger_type] += 1

    for k, v in num_functions_by_trigger_type.items():
      trigger_types_list.append(k)
      num_functions_list.append(v)

      shared_list.append("Shared" if len(num_functions_by_trigger_type) > 1 else "Exclusive")

  data = {
    "NumFunctions": num_functions_list,
    "InvocationTypes": trigger_types_list,
    "Shared": shared_list,
  }

  df = pd.DataFrame(data)

  fig, ax = plt.subplots(figsize=(6, 3))
  #sns.violinplot(x="NumFunctions", y="Categories", data=df, ax=ax, log_scale=False, inner=None, alpha=0.6)
  # for Matplotlib version >= 1.5
  #plt.gca().set_prop_cycle(None)
  sns.stripplot(
    data=data, x="NumFunctions", y="InvocationTypes", hue="Shared",
    jitter=0.1, log_scale=False, alpha=0.8,
    edgecolor="w", linewidth=0.5, order=unique(trigger_type_mapping.values()),
  )
  #ax.set_xscale("log")

  sns.move_legend(
    ax, "lower left",
    bbox_to_anchor=(0, 1),
    ncol=2,
    title=None,
    frameon=False,
  )

  plt.xlim(0.5, 48)
  plt.xticks([1, 5, 10, 20, 30, 40, 50])

  plt.xlabel("Number of Functions")
  plt.ylabel("")
  #plt.ylabel("Cml Fraction of Repositories [%]")

  #ax.yaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: f"{x*100:.0f}"))
  #ax.yaxis.set_minor_formatter(ticker.NullFormatter())

  # plt.xlim(0, max(df["Percentage"]) * 1.1)
  #plt.xlim(min(num_files_list), max(num_files_list))
  #plt.ylim(0.0, 1.05)

  # ax.xaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: ""))
  # ax.xaxis.set_minor_formatter(ticker.NullFormatter())

  plt.savefig(f"{out_dir}/num_functions_per_trigger_type.png", transparent=True, bbox_inches='tight')
  plt.close()

def plot_trigger_types_and_num_functions(repos, out_dir):
  data = { "TriggerFunctionCount": [], "FunctionCount": [], "TriggerType": [], "TriggerPercentage": [] }

  for repo in repos:
    functions = repo.get("Functions", [])
    function_count = len(functions)
    
    function_count_by_trigger_type = {}
    for function in functions:
      trigger_type = function["InvocationType"]
      mapped_trigger_type = trigger_type_mapping.get(trigger_type, trigger_type)

      if mapped_trigger_type not in function_count_by_trigger_type:
        function_count_by_trigger_type[mapped_trigger_type] = 0

      function_count_by_trigger_type[mapped_trigger_type] += 1

    for trigger_type, trigger_type_function_count in function_count_by_trigger_type.items():
      data["TriggerFunctionCount"].append(trigger_type_function_count)
      data["FunctionCount"].append(function_count)
      data["TriggerType"].append(trigger_type)
      data["TriggerPercentage"].append(trigger_type_function_count / function_count)

  for trigger_type in unique(trigger_type_mapping.values()):
    df = pd.DataFrame(data)

    df = df[df["TriggerType"] == trigger_type]

    fig, ax = plt.subplots(figsize=(6, 3))
    sns.scatterplot(
      x="FunctionCount",
      y="TriggerPercentage",
      data=df,
      #size="TriggerFunctionCount",
      #hue="TriggerType",
      #legend=False,
      #alpha=0.8,
    )
    #ax.set_xscale("log")

    #plt.xlim(30, 130_000)
    #plt.yticks([0, 5, 10, 20, 40, 60])

    #plt.xlabel("LoC")
    #plt.ylabel("Num Functions")

    plt.savefig(f"{out_dir}/trigger_type_{trigger_type}_and_num_functions.png", transparent=True, bbox_inches='tight')
    plt.close()

def plot_execution_location_and_num_functions(repos, out_dir):
  data = { "ExecutionLocationFunctionCount": [], "FunctionCount": [], "ExecutionLocation": [], "ExecutionLocationPercentage": [] }

  for repo in repos:
    functions = repo.get("Functions", [])
    function_count = len(functions)
    
    function_count_by_execution_location = {}
    for function in functions:
      execution_location = function["Location"]
      mapped_execution_location = execution_location_mapping.get(execution_location, execution_location)

      if mapped_execution_location not in function_count_by_execution_location:
        function_count_by_execution_location[mapped_execution_location] = 0

      function_count_by_execution_location[mapped_execution_location] += 1

    for execution_location, execution_location_function_count in function_count_by_execution_location.items():
      data["ExecutionLocationFunctionCount"].append(execution_location_function_count)
      data["FunctionCount"].append(function_count)
      data["ExecutionLocation"].append(execution_location)
      data["ExecutionLocationPercentage"].append(execution_location_function_count / function_count)

  for execution_location in unique(execution_location_mapping.values()):
    df = pd.DataFrame(data)

    df = df[df["ExecutionLocation"] == execution_location]

    fig, ax = plt.subplots(figsize=(6, 3))
    sns.scatterplot(
      x="FunctionCount",
      y="ExecutionLocationPercentage",
      data=df,
      #size="TriggerFunctionCount",
      #hue="TriggerType",
      #legend=False,
      #alpha=0.8,
    )
    #ax.set_xscale("log")

    #plt.xlim(30, 130_000)
    #plt.yticks([0, 5, 10, 20, 40, 60])

    #plt.xlabel("LoC")
    #plt.ylabel("Num Functions")

    plt.savefig(f"{out_dir}/execution_location_{execution_location}_and_num_functions.png", transparent=True, bbox_inches='tight')
    plt.close()

def plot_function_count_distribution(repos, num_bins, out_dir):
  function_count_distribution = {x: 0 for x in range(num_bins)}
  num_applications = 0

  for repo in repos:
    num_applications += 1

    functions = repo.get("Functions", [])

    num_functions = len(functions)
    bin_idx = num_functions - 1
    if bin_idx in function_count_distribution:
      function_count_distribution[bin_idx] += 1
    else:
      function_count_distribution[num_bins-1] += 1  

  data = {
    'NumFunctions': [x+1 for x in function_count_distribution.keys()],
    'Probability': [x / num_applications for x in function_count_distribution.values()],
  }

  df = pd.DataFrame(data)

  fig, ax = plt.subplots(figsize=(10, 6))
  sns.barplot(x='NumFunctions', y='Probability', data=df, ax=ax)

  current_ticks = plt.xticks()[0]
  current_labels = [str(x+1) for x in range(num_bins)]
  current_labels[num_bins-1] = f"{num_bins}+"

  plt.xticks(ticks=current_ticks, labels=current_labels)

  plt.xlabel('Number of Functions')
  plt.ylabel('Percentage of Applications')

  ax.yaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: f"{x*100:.0f}%"))
  ax.yaxis.set_minor_formatter(ticker.NullFormatter())

  plt.savefig(f"{out_dir}/function_count_distribution.png", transparent=True,  bbox_inches='tight')
  plt.close()

#################
### Platforms ###
#################

def print_num_platforms_per_application(repos):
  total_num_files = 0
  total_num_functions = 0
  total_num_repositories = 0

  num_platforms = {}
  num_functions_by_platform = {}
  num_repos_by_platform = {}

  num_frameworks = {}
  num_functions_by_framework = {}
  num_repos_by_framework = {}

  num_framework_combinations = {}
  num_platform_combinations = {}
  num_execution_location_combinations = {}
  num_trigger_type_combinations = {}

  platforms = set()
  frameworks = set()

  for repo in repos:
    total_num_repositories += 1

    repo_platforms = set()
    repo_frameworks = set()
    repo_execution_locations = set()
    repo_trigger_types = set()

    complexity = repo.get("Complexity", {})
    files = complexity.get("Files", [])
    total_num_files += len(files)
    
    functions = repo.get("Functions", [])
    for function in functions:
      total_num_functions += 1

      platform = function.get("Platform", "")
      mapped_platform = platform_mapping.get(platform, platform)

      framework = function.get("Framework", "")
      mapped_framework = framework_category_mapping.get(framework, framework)

      trigger_type = function.get("InvocationType", "")
      mapped_trigger_type = trigger_type_mapping.get(trigger_type,  trigger_type)

      execution_location = function.get("Location", "")
      mapped_execution_location = execution_location_mapping.get(execution_location, execution_location)

      repo_platforms.add(mapped_platform)
      repo_frameworks.add(mapped_framework)
      repo_trigger_types.add(mapped_trigger_type)
      repo_execution_locations.add(mapped_execution_location)
      
      platforms.add(mapped_platform)
      frameworks.add(mapped_framework)

      if mapped_platform not in num_functions_by_platform:
        num_functions_by_platform[mapped_platform] = 0

      if mapped_framework not in num_functions_by_framework:
        num_functions_by_framework[mapped_framework] = 0

      num_functions_by_platform[mapped_platform] += 1
      num_functions_by_framework[mapped_framework] += 1

    repo_num_platforms = len(repo_platforms)
    repo_num_frameworks = len(repo_frameworks)

    if repo_num_platforms not in num_platforms:
      num_platforms[repo_num_platforms] = 0
    num_platforms[repo_num_platforms] += 1

    if repo_num_frameworks not in num_frameworks:
      num_frameworks[repo_num_frameworks] = 0
    num_frameworks[repo_num_frameworks] += 1

    for repo_platform in repo_platforms:
      if repo_platform not in num_repos_by_platform:
        num_repos_by_platform[repo_platform] = 0
      num_repos_by_platform[repo_platform] += 1

    for repo_framework in repo_frameworks:
      if repo_framework not in num_repos_by_framework:
        num_repos_by_framework[repo_framework] = 0
      num_repos_by_framework[repo_framework] += 1

    if len(repo_platforms) > 0:
      platform_combination = ";".join(sorted(list(repo_platforms)))
      if platform_combination not in num_platform_combinations:
        num_platform_combinations[platform_combination] = 0
      num_platform_combinations[platform_combination] += 1

    if len(repo_frameworks) > 0:
      framework_combination = ";".join(sorted(list(repo_frameworks)))
      if framework_combination not in num_framework_combinations:
        num_framework_combinations[framework_combination] = 0
      num_framework_combinations[framework_combination] += 1

    if len(repo_execution_locations) > 0:
      execution_location_combination = ";".join(sorted(list(repo_execution_locations)))
      if execution_location_combination not in num_execution_location_combinations:
        num_execution_location_combinations[execution_location_combination] = 0
      num_execution_location_combinations[execution_location_combination] += 1

    if len(repo_trigger_types) > 0:
      trigger_type_combination = ";".join(sorted(list(repo_trigger_types)))
      if trigger_type_combination not in num_trigger_type_combinations:
        num_trigger_type_combinations[trigger_type_combination] = 0
      num_trigger_type_combinations[trigger_type_combination] += 1

  avg_num_functions_by_platform = {}
  for platform in platforms:
    avg_num_functions_by_platform[platform] = num_functions_by_platform.get(platform, 0) / num_repos_by_platform.get(platform, 1)

  avg_num_functions_by_framework = {}
  for framework in frameworks:
    avg_num_functions_by_framework[framework] = num_functions_by_framework.get(framework, 0) / num_repos_by_framework.get(framework, 1)

  print("#" * 80)
  print("General")
  print("#" * 80)
  print(f"Num Files: {total_num_files}")
  print(f"Num Functions: {total_num_functions}")
  print(f"Num Repositories: {total_num_repositories}")
  print("#" * 80)
  print("")

  print("#" * 80)
  print("Num Platforms per application")
  print("#" * 80)
  for k, v in sorted(num_platforms.items(), key=lambda x: x[1]):
    print(f"{k} Platforms = {v}")
  print("#" * 80)
  print("")

  print("#" * 80)
  print("Num Frameworks per application")
  print("#" * 80)
  for k, v in sorted(num_frameworks.items(), key=lambda x: x[1]):
    print(f"{k} Frameworks = {v}")
  print("#" * 80)
  print("")

  print("#" * 80)
  print("Avg Num Functions per Platform")
  print("#" * 80)
  for k, v in sorted(avg_num_functions_by_platform.items(), key=lambda x: x[1]):
    print(f"{k} = {v} ({num_functions_by_platform[k]} Functions | {num_repos_by_platform[k]} Applications)")
  print("#" * 80)
  print("")

  print("#" * 80)
  print("Avg Num Functions per Framework Category")
  print("#" * 80)
  for k, v in sorted(avg_num_functions_by_framework.items(), key=lambda x: x[1]):
    print(f"{k} = {v} ({num_functions_by_framework[k]} Functions | {num_repos_by_framework[k]} Applications)")
  print("#" * 80)
  print("")

  print("#" * 80)
  print("Framework Combinations")
  print("#" * 80)
  for k, v in sorted(num_framework_combinations.items(), key=lambda x: x[1]):
    print(f"{k} = {v}")
  print("#" * 80)
  print("")

  print("#" * 80)
  print("Platform Combinations")
  print("#" * 80)
  for k, v in sorted(num_platform_combinations.items(), key=lambda x: x[1]):
    print(f"{k} = {v}")
  print("#" * 80)
  print("")

  print("#" * 80)
  print("Execution Location Combinations")
  print("#" * 80)
  for k, v in sorted(num_execution_location_combinations.items(), key=lambda x: x[1]):
    print(f"{k} = {v}")
  print("#" * 80)
  print("")

  print("#" * 80)
  print("Trigger Type Combinations")
  print("#" * 80)
  for k, v in sorted(num_trigger_type_combinations.items(), key=lambda x: x[1]):
    print(f"{k} = {v}")
  print("#" * 80)
  print("")

def plot_platforms_per_function(repos, out_dir):
  platform_count = {}
  num_functions = 0

  for repo in repos:
    functions = repo.get("Functions", [])
    for function in functions:
        num_functions += 1

        platform = function.get("Platform", "")
        mapped_platform = platform_mapping.get(platform, platform)

        if mapped_platform not in platform_count:
          platform_count[mapped_platform] = 1
        else:
          platform_count[mapped_platform] += 1

  data = {
    "Platforms": platform_count.keys(),
    "Percentage": [x / num_functions for x in platform_count.values()],
  }

  df = pd.DataFrame(data)
  df['Platforms'] = pd.Categorical(df['Platforms'], categories=unique(platform_mapping.values()), ordered=True)
  df = df.sort_values('Platforms')

  fig, ax = plt.subplots(figsize=(10, 6))
  sns.barplot(x="Percentage", y="Platforms", data=df, orient="h", ax=ax)

  ax.xaxis.set_major_locator(ticker.MultipleLocator(0.05))

  for index, value in enumerate(df["Percentage"]):
    plt.text(value + 0.005, index, f"{value:.2%}", va="center")

  plt.xlim(0, max(df["Percentage"]) * 1.1)

  plt.xlabel("Percentage of Functions")
  plt.ylabel("Platform")

  ax.xaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: ""))
  ax.xaxis.set_minor_formatter(ticker.NullFormatter())

  plt.savefig(f"{out_dir}/platforms_per_function.png", transparent=True, bbox_inches='tight')
  plt.close()

def plot_platforms_per_application(repos, out_dir):
  platform_count = {}
  num_repos = len(repos)

  for repo in repos:
    repo_platforms = set()

    functions = repo.get("Functions", [])
    for function in functions:
      platform = function["Platform"]
      repo_platforms.add(platform)

    for platform in repo_platforms:
      mapped_platform = platform_mapping.get(platform, platform)

      if mapped_platform not in platform_count:
        platform_count[mapped_platform] = 1
      else:
        platform_count[mapped_platform] += 1

  data = {
    "Platforms": platform_count.keys(),
    "Percentage": [x / num_repos for x in platform_count.values()],
  }

  df = pd.DataFrame(data)
  df['Platforms'] = pd.Categorical(df['Platforms'], categories=unique(platform_mapping.values()), ordered=True)
  df = df.sort_values('Platforms')

  fig, ax = plt.subplots(figsize=(10, 6))
  sns.barplot(x="Percentage", y="Platforms", data=df, orient="h", ax=ax)

  ax.xaxis.set_major_locator(ticker.MultipleLocator(0.05))

  for index, value in enumerate(df["Percentage"]):
    plt.text(value + 0.005, index, f"{value:.2%}", va="center")

  plt.xlim(0, max(df["Percentage"]) * 1.1)

  plt.xlabel("Percentage of Applications")
  plt.ylabel("Platform")

  ax.xaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: ""))
  ax.xaxis.set_minor_formatter(ticker.NullFormatter())

  plt.savefig(f"{out_dir}/platforms_per_application.png", transparent=True, bbox_inches='tight')
  plt.close()

def plot_platforms_per_application_and_function(repos, out_dir):
  repo_platform_count = {}
  function_platform_count = {}
  num_repos = len(repos)
  num_functions = 0

  for repo in repos:
    repo_platforms = set()

    functions = repo.get("Functions", [])
    for function in functions:
      num_functions += 1

      platform = function.get("Platform", "")
      repo_platforms.add(platform)
      mapped_platform = platform_mapping.get(platform, platform)

      if mapped_platform not in function_platform_count:
        function_platform_count[mapped_platform] = 1
      else:
        function_platform_count[mapped_platform] += 1

    for platform in repo_platforms:
      mapped_platform = platform_mapping.get(platform, platform)

      if mapped_platform not in repo_platform_count:
        repo_platform_count[mapped_platform] = 1
      else:
        repo_platform_count[mapped_platform] += 1

  platforms = []
  percentages = []
  hues = []

  platforms.extend(function_platform_count.keys())
  percentages.extend([x / num_functions for x in function_platform_count.values()])
  hues.extend(["Functions"] * len(function_platform_count.keys()))

  platforms.extend(repo_platform_count.keys())
  percentages.extend([x / num_repos for x in repo_platform_count.values()])
  hues.extend(["Applications"] * len(repo_platform_count.keys()))

  data = {
    "Platforms": platforms,
    "Percentage": percentages,
    "Hue": hues,
  }

  order = unique(platform_mapping.values())
  hue_order = ["Functions", "Applications"]

  df = pd.DataFrame(data)
  df["Platforms"] = pd.Categorical(df["Platforms"], categories=order, ordered=True)
  df = df.sort_values("Platforms").reset_index()

  fig, ax = plt.subplots(figsize=(6, 3))
  sns.barplot(x="Percentage", y="Platforms", hue="Hue", order=order, hue_order=hue_order, data=df, orient="h", ax=ax)

  sns.move_legend(
    ax, "lower left",
    bbox_to_anchor=(0, 1),
    ncol=2,
    title=None,
    frameon=False,
  )

  ax.xaxis.set_major_locator(ticker.MultipleLocator(0.05))

  for index, row in df.iterrows():
    percentage = row["Percentage"]
    platform = row["Platforms"]
    hue = row["Hue"]
    if hue == "Functions":
      plt.text(percentage + 0.0025, (index//2)-0.2, f"{percentage:.2%}", va="center", fontsize=8)
    else:
      plt.text(percentage + 0.0025, (index//2)+0.2, f"{percentage:.2%}", va="center", fontsize=8)

  plt.xlim(0, max(df["Percentage"]) * 1.15)

  plt.xlabel("Proportion [%]")
  plt.ylabel("")

  ax.xaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: ""))
  ax.xaxis.set_minor_formatter(ticker.NullFormatter())

  plt.savefig(f"{out_dir}/platforms_per_application_and_function.png", transparent=True, bbox_inches='tight')
  plt.close()

def plot_platforms_matrix(repos, out_dir):
  num_platform_combinations = {}
  for repo in repos:
    repo_platforms = set()

    functions = repo.get("Functions", [])
    for function in functions:
      platform = function["Platform"]
      mapped_platform = platform_mapping.get(platform, platform)

      repo_platforms.add(mapped_platform)

    repo_platforms = list(repo_platforms)
    repo_platforms_combinations = itertools.combinations(repo_platforms, 2)

    for left, right in repo_platforms_combinations:
      key = ";".join(sort_respecting([left, right], unique(platform_mapping.values())))

      if key not in num_platform_combinations:
        num_platform_combinations[key] = 0
      num_platform_combinations[key] += 1

  row_platforms = []
  column_platforms = []
  num_repositories = []

  for k, v in num_platform_combinations.items():
    left, right = k.split(";")

    row_platforms.append(right)
    column_platforms.append(left)
    num_repositories.append(v)

  data = {
    "Row": row_platforms,
    "Column": column_platforms,
    "Value": num_repositories,
  }

  df = pd.DataFrame(data)
  df = df.pivot(index="Row", columns="Column", values="Value")
  df = df.reindex(index=unique(platform_mapping.values()), columns=unique(platform_mapping.values()))
  df = df.drop([unique(platform_mapping.values())[0]])
  df = df.drop(columns=[unique(platform_mapping.values())[-1]])
  df = df.dropna(axis="columns", how="all")
  df = df.dropna(axis="rows", how="all")

  mask = df.to_numpy()
  mask = np.arange(mask.shape[0])[:,None] < np.arange(mask.shape[1])

  steps = [1, 2, 5, 10]
  for step in steps:
    if max(num_repositories) / step <= 11:
      break

  ticks = [0]
  tick = 0
  while tick < max(num_repositories):
    tick = tick + step
    ticks.append(tick)

  fig, ax = plt.subplots(figsize=(6, 3))

  sns.heatmap(df, annot=True, fmt=".0f", cmap="Blues", cbar=False, square=True, mask=mask, vmin=0, vmax=max(num_repositories), cbar_kws={"ticks": ticks})

  ax.spines[:].set_visible(True)

  #ax.xaxis.set_major_locator(ticker.MultipleLocator(0.05))

  #plt.xlim(0, max(df["Percentage"]) * 1.1)

  plt.xlabel("")
  plt.ylabel("")

  # ax.xaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: ""))
  # ax.xaxis.set_minor_formatter(ticker.NullFormatter())

  plt.savefig(f"{out_dir}/platforms_matrix.png", transparent=True, bbox_inches='tight')
  plt.close()

# def plot_platforms_per_framework(repos, max_per_bar, out_dir):
#   pass

def plot_platforms_per_execution_location(repos, max_per_bar, out_dir):
  execution_locations_per_platform = {x: {y: 0 for y in unique(execution_location_mapping.values())} for x in unique(platform_mapping.values())}
  platforms_per_execution_location = {x: {y: 0 for y in unique(platform_mapping.values())} for x in unique(execution_location_mapping.values())}

  num_functions_per_execution_location = {x: 0 for x in unique(execution_location_mapping.values())}

  for repo in repos:
    functions = repo.get("Functions", [])

    for function in functions:
      execution_location = function.get("Location", "")
      mapped_execution_location = execution_location_mapping.get(execution_location, execution_location)

      platform = function.get("Platform", "")
      mapped_platform = platform_mapping.get(platform, platform)

      execution_locations_per_platform[mapped_platform][mapped_execution_location] += 1
      platforms_per_execution_location[mapped_execution_location][mapped_platform] += 1
      num_functions_per_execution_location[mapped_execution_location] += 1

  labeled_platforms_per_execution_location = {x: [] for x in unique(execution_location_mapping.values())}

  has_other = False

  for execution_location, platforms in platforms_per_execution_location.items():
    platforms = sorted(platforms.items(), key=lambda x: x[1], reverse=True)
    if len(platforms) > max_per_bar:
      has_other = True
      labeled_platforms = [k for k, v in platforms[:max_per_bar-1]]
      labeled_platforms_per_execution_location[execution_location].extend(labeled_platforms)
    else:
      labeled_platforms = [k for k, v in platforms]
      labeled_platforms_per_execution_location[execution_location].extend(labeled_platforms)

  fig, ax = plt.subplots(figsize=(10, 6))

  if has_other:
    df = pd.DataFrame({
      "ExecutionLocation": unique(execution_location_mapping.values()),
      "Percentage": [1.0 for _ in unique(execution_location_mapping.values())],
    })
    sns.barplot(x="Percentage", y="ExecutionLocation", data=df, orient="h", ax=ax, label="Other", order=unique(execution_location_mapping.values()))

  offset_per_execution_location = {x: 1.0 for x in unique(execution_location_mapping.values())}

  # First pass to calculate others
  for platform, execution_locations in list(execution_locations_per_platform.items())[::-1]:
    for execution_location, count in execution_locations.items():
      if platform not in labeled_platforms_per_execution_location[execution_location]:
        offset_per_execution_location[execution_location] -= count / num_functions_per_execution_location[execution_location]

  # Second pass to calculate others
  for platform, execution_locations in list(execution_locations_per_platform.items())[::-1]:
    execution_locations_count = {x: 0 for x in unique(execution_location_mapping.values())}

    for execution_location, count in execution_locations.items():
      if platform in labeled_platforms_per_execution_location[execution_location]:
        execution_locations_count[execution_location] = count
      else:
        execution_locations_count[execution_location] = 0

    df = pd.DataFrame({
      "ExecutionLocation": execution_locations_count.keys(),
      "Percentage": [offset_per_execution_location[execution_location] for execution_location in execution_locations_count.keys()],
    })

    sns.barplot(x="Percentage", y="ExecutionLocation", data=df, orient="h", ax=ax, label=platform, order=unique(execution_location_mapping.values()))

    for execution_location in execution_locations_count.keys():
      offset_per_execution_location[execution_location] -= execution_locations_count[execution_location] / num_functions_per_execution_location[execution_location]

  handles, labels = ax.get_legend_handles_labels()
  ax.legend(handles[::-1], labels[::-1], title='Line', loc='upper left')
  # sns.move_legend(ax, "upper left", bbox_to_anchor=(0, -0.3))

  sns.move_legend(
    ax, "lower left",
    bbox_to_anchor=(0, 1), ncol=3,
    title=None, frameon=False,
  )

  ax.xaxis.set_major_locator(ticker.MultipleLocator(0.2))

  plt.xlabel("Distribution [%] of Function Platforms")
  plt.ylabel("Execution Location")

  plt.xlim(0, 1)

  ax.xaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: f"{x*100:.0f}"))
  ax.xaxis.set_minor_formatter(ticker.NullFormatter())

  # plt.show()
  plt.savefig(f"{out_dir}/platforms_per_execution_location.png", transparent=True, bbox_inches="tight", pad_inches=1)
  plt.close()

def plot_platforms_per_trigger_type(repos, max_per_bar, out_dir):
  trigger_types_per_platform = {x: {y: 0 for y in unique(trigger_type_mapping.values())} for x in unique(platform_mapping.values())}
  platforms_per_trigger_type = {x: {y: 0 for y in unique(platform_mapping.values())} for x in unique(trigger_type_mapping.values())}

  num_functions_per_trigger_type = {x: 0 for x in unique(trigger_type_mapping.values())}

  for repo in repos:
    functions = repo.get("Functions", [])

    for function in functions:
      trigger_type = function.get("InvocationType", "")
      mapped_trigger_type = trigger_type_mapping.get(trigger_type, trigger_type)

      platform = function.get("Platform", "")
      mapped_platform = platform_mapping.get(platform, platform)

      trigger_types_per_platform[mapped_platform][mapped_trigger_type] += 1
      platforms_per_trigger_type[mapped_trigger_type][mapped_platform] += 1
      num_functions_per_trigger_type[mapped_trigger_type] += 1

  has_other = False

  labeled_platforms_per_trigger_type = {x: [] for x in unique(trigger_type_mapping.values())}

  for trigger_type, platforms in platforms_per_trigger_type.items():
    platforms = sorted(platforms.items(), key=lambda x: x[1], reverse=True)
    if len(platforms) > max_per_bar:
      has_other = True
      labeled_platforms = [k for k, v in platforms[:max_per_bar-1]]
      labeled_platforms_per_trigger_type[trigger_type].extend(labeled_platforms)
    else:
      labeled_platforms = [k for k, v in platforms]
      labeled_platforms_per_trigger_type[trigger_type].extend(labeled_platforms)

  fig, ax = plt.subplots(figsize=(10, 6))

  if has_other:
    df = pd.DataFrame({
      "TriggerType": unique(trigger_type_mapping.values()),
      "Percentage": [1.0 for _ in unique(trigger_type_mapping.values())],
    })
    sns.barplot(x="Percentage", y="TriggerType", data=df, orient="h", ax=ax, label="Other", order=unique(trigger_type_mapping.values()))

  offset_per_trigger_type = {x: 1.0 for x in unique(trigger_type_mapping.values())}

  # First pass to calculate others
  for platform, trigger_types in list(trigger_types_per_platform.items())[::-1]:
    for trigger_type, count in trigger_types.items():
      if platform not in labeled_platforms_per_trigger_type[trigger_type]:
        offset_per_trigger_type[trigger_type] -= count / num_functions_per_trigger_type[trigger_type]

  # Second pass to calculate others
  for platform, trigger_types in list(trigger_types_per_platform.items())[::-1]:
    trigger_types_count = {x: 0 for x in unique(trigger_type_mapping.values())}

    for trigger_type, count in trigger_types.items():
      if platform in labeled_platforms_per_trigger_type[trigger_type]:
        trigger_types_count[trigger_type] = count
      else:
        trigger_types_count[trigger_type] = 0

    df = pd.DataFrame({
      "TriggerType": trigger_types_count.keys(),
      "Percentage": [offset_per_trigger_type[trigger_type] for trigger_type in trigger_types_count.keys()],
    })

    sns.barplot(x="Percentage", y="TriggerType", data=df, orient="h", ax=ax, label=platform, order=unique(trigger_type_mapping.values()))

    for trigger_type in trigger_types_count.keys():
      offset_per_trigger_type[trigger_type] -= trigger_types_count[trigger_type] / num_functions_per_trigger_type[trigger_type]

  handles, labels = ax.get_legend_handles_labels()
  ax.legend(handles[::-1], labels[::-1], title='Line', loc='upper left')
  # sns.move_legend(ax, "upper left", bbox_to_anchor=(0, -0.3))

  sns.move_legend(
    ax, "lower left",
    bbox_to_anchor=(0, 1), ncol=3,
    title=None, frameon=False,
  )

  ax.xaxis.set_major_locator(ticker.MultipleLocator(0.2))

  plt.xlabel("Distribution [%] of Function Platforms")
  plt.ylabel("Trigger Type")

  plt.xlim(0, 1)

  ax.xaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: f"{x*100:.0f}"))
  ax.xaxis.set_minor_formatter(ticker.NullFormatter())

  # plt.show()
  plt.savefig(f"{out_dir}/platforms_per_trigger_type.png", transparent=True, bbox_inches="tight", pad_inches=1)
  plt.close()

##################
### Frameworks ###
##################

def plot_frameworks_per_function(repos, out_dir):
  framework_count = {}
  num_functions = 0

  for repo in repos:
    functions = repo.get("Functions", [])
    for function in functions:
        num_functions += 1

        framework = function.get("Framework", "")
        mapped_framework = framework_mapping.get(framework, framework)

        if mapped_framework not in framework_count:
          framework_count[mapped_framework] = 1
        else:
          framework_count[mapped_framework] += 1

  data = {
    "Frameworks": framework_count.keys(),
    "Percentage": [x / num_functions for x in framework_count.values()],
  }

  df = pd.DataFrame(data)
  df['Frameworks'] = pd.Categorical(df['Frameworks'], categories=unique(framework_mapping.values()), ordered=True)
  df = df.sort_values('Frameworks')

  fig, ax = plt.subplots(figsize=(10, 6))
  sns.barplot(x="Percentage", y="Frameworks", data=df, orient="h", ax=ax)

  ax.xaxis.set_major_locator(ticker.MultipleLocator(0.05))

  for index, value in enumerate(df["Percentage"]):
    plt.text(value + 0.0025, index, f'{value:.2%}', va="center")

  plt.xlim(0, max(df["Percentage"]) * 1.1)

  plt.xlabel("Percentage of Functions")
  plt.ylabel("Frameworks")

  ax.xaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: ""))
  ax.xaxis.set_minor_formatter(ticker.NullFormatter())

  plt.savefig(f"{out_dir}/frameworks_per_function.png", transparent=True, bbox_inches='tight')
  plt.close()

def plot_frameworks_per_application(repos, out_dir):
  framework_count = {}
  num_repos = len(repos)

  for repo in repos:
    repo_frameworks = set()

    functions = repo.get("Functions",  [])
    for function in functions:
      framework = function["Framework"]
      repo_frameworks.add(framework)

    for framework in repo_frameworks:
      mapped_framework = framework_mapping.get(framework, framework)

      if mapped_framework not in framework_count:
        framework_count[mapped_framework] = 1
      else:
        framework_count[mapped_framework] += 1

  data = {
    "Frameworks": framework_count.keys(),
    "Percentage": [x / num_repos for x in framework_count.values()],
  }

  df = pd.DataFrame(data)
  df['Frameworks'] = pd.Categorical(df['Frameworks'], categories=unique(framework_mapping.values()), ordered=True)
  df = df.sort_values('Frameworks')

  fig, ax = plt.subplots(figsize=(10, 6))
  sns.barplot(x="Percentage", y="Frameworks", data=df, orient="h", ax=ax)

  ax.xaxis.set_major_locator(ticker.MultipleLocator(0.05))

  for index, value in enumerate(df["Percentage"]):
    plt.text(value + 0.0025, index, f'{value:.2%}', va="center")

  plt.xlim(0, max(df["Percentage"]) * 1.1)

  plt.xlabel("Percentage of Applications")
  plt.ylabel("Frameworks")

  ax.xaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: ""))
  ax.xaxis.set_minor_formatter(ticker.NullFormatter())

  plt.savefig(f"{out_dir}/frameworks_per_application.png", transparent=True, bbox_inches='tight')
  plt.close()

def plot_frameworks_per_application_and_function(repos, out_dir):
  repo_framework_count = {}
  function_framework_count = {}
  num_repos = len(repos)
  num_functions = 0

  for repo in repos:
    repo_frameworks = set()

    functions = repo.get("Functions", [])
    for function in functions:
      num_functions += 1

      framework = function.get("Framework", "")
      repo_frameworks.add(framework)
      mapped_framework = framework_mapping.get(framework, framework)

      if mapped_framework not in function_framework_count:
        function_framework_count[mapped_framework] = 1
      else:
        function_framework_count[mapped_framework] += 1

    for framework in repo_frameworks:
      mapped_framework = framework_mapping.get(framework, framework)

      if mapped_framework not in repo_framework_count:
        repo_framework_count[mapped_framework] = 1
      else:
        repo_framework_count[mapped_framework] += 1

  frameworks = []
  percentages = []
  hues = []

  frameworks.extend(function_framework_count.keys())
  percentages.extend([x / num_functions for x in function_framework_count.values()])
  hues.extend(["Functions"] * len(function_framework_count.keys()))

  frameworks.extend(repo_framework_count.keys())
  percentages.extend([x / num_repos for x in repo_framework_count.values()])
  hues.extend(["Applications"] * len(repo_framework_count.keys()))

  data = {
    "Frameworks": frameworks,
    "Percentage": percentages,
    "Hue": hues,
  }

  order = unique(framework_mapping.values())
  hue_order = ["Functions", "Applications"]

  df = pd.DataFrame(data)
  df["Frameworks"] = pd.Categorical(df["Frameworks"], categories=order, ordered=True)
  df = df.sort_values("Frameworks").reset_index()

  fig, ax = plt.subplots(figsize=(10, 6))
  sns.barplot(x="Percentage", y="Frameworks", hue="Hue", order=order, hue_order=hue_order, data=df, orient="h", ax=ax)

  sns.move_legend(
    ax, "lower left",
    bbox_to_anchor=(0, 1),
    ncol=2,
    title=None,
    frameon=False,
  )

  ax.xaxis.set_major_locator(ticker.MultipleLocator(0.05))

  for index, row in df.iterrows():
    percentage = row["Percentage"]
    framework = row["Frameworks"]
    hue = row["Hue"]
    if hue == "Functions":
      plt.text(percentage + 0.0025, (index//2)-0.2, f"{percentage:.2%}", va="center", fontsize=8)
    else:
      plt.text(percentage + 0.0025, (index//2)+0.2, f"{percentage:.2%}", va="center", fontsize=8)

  plt.xlim(0, max(df["Percentage"]) * 1.1)

  plt.xlabel("Proportion [%]")
  plt.ylabel("")

  ax.xaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: ""))
  ax.xaxis.set_minor_formatter(ticker.NullFormatter())

  plt.savefig(f"{out_dir}/frameworks_per_application_and_function.png", transparent=True, bbox_inches='tight')
  plt.close()

def plot_framework_categories_per_application_and_function(repos, out_dir):
  repo_framework_category_count = {}
  function_framework_category_count = {}
  num_repos = len(repos)
  num_functions = 0

  for repo in repos:
    repo_framework_categories = set()

    functions = repo.get("Functions", [])
    for function in functions:
      num_functions += 1

      framework_category = function.get("Framework", "")
      repo_framework_categories.add(framework_category)
      mapped_framework_category = framework_category_mapping.get(framework_category, framework_category)

      if mapped_framework_category not in function_framework_category_count:
        function_framework_category_count[mapped_framework_category] = 1
      else:
        function_framework_category_count[mapped_framework_category] += 1

    for framework_category in repo_framework_categories:
      mapped_framework_category = framework_category_mapping.get(framework_category, framework_category)

      if mapped_framework_category not in repo_framework_category_count:
        repo_framework_category_count[mapped_framework_category] = 1
      else:
        repo_framework_category_count[mapped_framework_category] += 1

  framework_categories = []
  percentages = []
  hues = []

  framework_categories.extend(function_framework_category_count.keys())
  percentages.extend([x / num_functions for x in function_framework_category_count.values()])
  hues.extend(["Functions"] * len(function_framework_category_count.keys()))

  framework_categories.extend(repo_framework_category_count.keys())
  percentages.extend([x / num_repos for x in repo_framework_category_count.values()])
  hues.extend(["Applications"] * len(repo_framework_category_count.keys()))

  data = {
    "Categories": framework_categories,
    "Percentage": percentages,
    "Hue": hues,
  }

  order = unique(framework_category_mapping.values())
  hue_order = ["Functions", "Applications"]

  df = pd.DataFrame(data)
  df["Categories"] = pd.Categorical(df["Categories"], categories=order, ordered=True)
  df = df.sort_values("Categories").reset_index()

  fig, ax = plt.subplots(figsize=(6, 3))
  sns.barplot(x="Percentage", y="Categories", hue="Hue", order=order, hue_order=hue_order, data=df, orient="h", ax=ax)

  sns.move_legend(
    ax, "lower left",
    bbox_to_anchor=(0, 1),
    ncol=2,
    title=None,
    frameon=False,
  )

  ax.xaxis.set_major_locator(ticker.MultipleLocator(0.05))

  for index, row in df.iterrows():
    percentage = row["Percentage"]
    hue = row["Hue"]
    if hue == "Functions":
      plt.text(percentage + 0.0025, (index//2)-0.2, f"{percentage:.2%}", va="center", fontsize=8)
    else:
      plt.text(percentage + 0.0025, (index//2)+0.2, f"{percentage:.2%}", va="center", fontsize=8)

  plt.xlim(0, max(df["Percentage"]) * 1.15)

  plt.xlabel("Proportion [%]")
  plt.ylabel("")

  ax.xaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: ""))
  ax.xaxis.set_minor_formatter(ticker.NullFormatter())

  plt.savefig(f"{out_dir}/framework_categories_per_application_and_function.png", transparent=True, bbox_inches='tight')
  plt.close()

def plot_frameworks_matrix(repos, out_dir):
  num_framework_combinations = {}

  for repo in repos:
    repo_frameworks = set()

    functions = repo.get("Functions", [])
    for function in functions:
      framework = function["Framework"]
      mapped_framework = framework_mapping.get(framework, framework)

      repo_frameworks.add(mapped_framework)

    repo_frameworks = list(repo_frameworks)
    repo_frameworks_combinations = itertools.combinations(repo_frameworks, 2)

    for left, right in repo_frameworks_combinations:
      key = ";".join(sort_respecting([left, right], unique(framework_mapping.values())))

      if key not in num_framework_combinations:
        num_framework_combinations[key] = 0
      num_framework_combinations[key]  += 1

  row_frameworks = []
  column_frameworks = []
  num_repositories = []

  for k, v in num_framework_combinations.items():
    left, right = k.split(";")

    row_frameworks.append(right)
    column_frameworks.append(left)
    num_repositories.append(v)

  data = {
    "Row": row_frameworks,
    "Column": column_frameworks,
    "Value": num_repositories,
  }

  df = pd.DataFrame(data)
  df = df.pivot(index="Row", columns="Column", values="Value")
  df = df.reindex(index=unique(framework_mapping.values()), columns=unique(framework_mapping.values()))
  df = df.drop([unique(framework_mapping.values())[0]])
  df = df.drop(columns=[unique(framework_mapping.values())[-1]])
  df = df.dropna(axis="columns", how="all")
  df = df.dropna(axis="rows", how="all")

  mask = df.to_numpy()
  mask = np.arange(mask.shape[0])[:,None] < np.arange(mask.shape[1])

  steps = [1, 2, 5, 10]
  for step in steps:
    if max(num_repositories) / step <= 11:
      break

  ticks = [0]
  tick = 0
  while tick < max(num_repositories):
    tick = tick + step
    ticks.append(tick)

  fig, ax = plt.subplots(figsize=(6, 3))

  sns.heatmap(df, annot=True, fmt=".0f", cmap="Blues", cbar=False, square=True, mask=mask, vmin=0, vmax=max(num_repositories), cbar_kws={"ticks": ticks})

  ax.spines[:].set_visible(True)

  plt.xticks(rotation=45, ha='right')

  #ax.xaxis.set_major_locator(ticker.MultipleLocator(0.05))

  #plt.xlim(0, max(df["Percentage"]) * 1.1)

  plt.xlabel("")
  plt.ylabel("")

  # ax.xaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: ""))
  # ax.xaxis.set_minor_formatter(ticker.NullFormatter())

  plt.savefig(f"{out_dir}/frameworks_matrix.png", transparent=True, bbox_inches='tight')
  plt.close()

def plot_framework_categories_matrix(repos, out_dir):
  num_framework_categories_combinations = {}

  for repo in repos:
    repo_framework_categoriess = set()

    functions = repo.get("Functions", [])
    for function in functions:
      framework = function["Framework"]
      mapped_framework_categories = framework_category_mapping.get(framework, framework)

      repo_framework_categoriess.add(mapped_framework_categories)

    repo_framework_categoriess = list(repo_framework_categoriess)
    repo_framework_categoriess_combinations = itertools.combinations(repo_framework_categoriess, 2)

    for left, right in repo_framework_categoriess_combinations:
      key = ";".join(sort_respecting([left, right], unique(framework_category_mapping.values())))

      if key not in num_framework_categories_combinations:
        num_framework_categories_combinations[key] = 0
      num_framework_categories_combinations[key]  += 1

  row_framework_categoriess = []
  column_framework_categoriess = []
  num_repositories = []

  for k, v in num_framework_categories_combinations.items():
    left, right = k.split(";")

    row_framework_categoriess.append(right)
    column_framework_categoriess.append(left)
    num_repositories.append(v)

  data = {
    "Row": row_framework_categoriess,
    "Column": column_framework_categoriess,
    "Value": num_repositories,
  }

  df = pd.DataFrame(data)
  df = df.pivot(index="Row", columns="Column", values="Value")
  df = df.reindex(index=unique(framework_category_mapping.values()), columns=unique(framework_category_mapping.values()))
  df = df.drop([unique(framework_category_mapping.values())[0]])
  df = df.drop(columns=[unique(framework_category_mapping.values())[-1]])
  df = df.dropna(axis="columns", how="all")
  df = df.dropna(axis="rows", how="all")

  mask = df.to_numpy()
  mask = np.arange(mask.shape[0])[:,None] < np.arange(mask.shape[1])

  steps = [1, 2, 5, 10]
  for step in steps:
    if max(num_repositories) / step <= 11:
      break

  ticks = [0]
  tick = 0
  while tick < max(num_repositories):
    tick = tick + step
    ticks.append(tick)

  fig, ax = plt.subplots(figsize=(6, 3))

  sns.heatmap(df, annot=True, fmt=".0f", cmap="Blues", cbar=False, square=True, mask=mask, vmin=0, vmax=max(num_repositories), cbar_kws={"ticks": ticks})

  ax.spines[:].set_visible(True)

  #ax.xaxis.set_major_locator(ticker.MultipleLocator(0.05))

  #plt.xlim(0, max(df["Percentage"]) * 1.1)

  plt.xlabel("")
  plt.ylabel("")

  # ax.xaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: ""))
  # ax.xaxis.set_minor_formatter(ticker.NullFormatter())

  plt.savefig(f"{out_dir}/framework_categories_matrix.png", transparent=True, bbox_inches='tight')
  plt.close()

# def plot_frameworks_per_platform(repos, max_per_bar, out_dir):
#   pass

def plot_frameworks_per_trigger_type(repos, max_per_bar, out_dir):
  trigger_types_per_framework = {x: {y: 0 for y in unique(trigger_type_mapping.values())} for x in unique(framework_category_mapping.values())}
  frameworks_per_trigger_type = {x: {y: 0 for y in unique(framework_category_mapping.values())} for x in unique(trigger_type_mapping.values())}

  num_functions_per_trigger_type = {x: 0 for x in unique(trigger_type_mapping.values())}

  for repo in repos:
    functions = repo.get("Functions", [])

    for function in functions:
      trigger_type = function.get("InvocationType", "")
      mapped_trigger_type = trigger_type_mapping.get(trigger_type, trigger_type)

      framework = function.get("Framework", "")
      mapped_framework = framework_category_mapping.get(framework, framework)

      trigger_types_per_framework[mapped_framework][mapped_trigger_type] += 1
      frameworks_per_trigger_type[mapped_trigger_type][mapped_framework] += 1
      num_functions_per_trigger_type[mapped_trigger_type] += 1

  has_other = False

  labeled_frameworks_per_trigger = {x: [] for x in unique(trigger_type_mapping.values())}

  for trigger_type, frameworks in frameworks_per_trigger_type.items():
    frameworks = sorted(frameworks.items(), key=lambda x: x[1], reverse=True)
    if len(frameworks) > max_per_bar:
      has_other = True
      labeled_frameworks = [k for k, v in frameworks[:max_per_bar-1]]
      labeled_frameworks_per_trigger[trigger_type].extend(labeled_frameworks)
    else:
      labeled_frameworks = [k for k, v in frameworks]
      labeled_frameworks_per_trigger[trigger_type].extend(labeled_frameworks)

  num_colors = 0
  if has_other:
    num_colors += 1
  for execution_location, trigger_types in trigger_types_per_framework.items():
    if sum(trigger_types.values()) > 0:
      num_colors += 1

  current_palette = sns.color_palette()
  colors = current_palette[:num_colors][::-1]

  color_offset = 0

  fig, ax = plt.subplots(figsize=(8, 3))

  if has_other:
    df = pd.DataFrame({
      "TriggerType": unique(trigger_type_mapping.values()),
      "Percentage": [1.0 for _ in unique(trigger_type_mapping.values())],
    })
    sns.barplot(x="Percentage", y="TriggerType", data=df, color=colors[color_offset], orient="h", ax=ax, label="Other", order=unique(trigger_type_mapping.values()))
    color_offset += 1

  offset_per_trigger_type = {x: 1.0 for x in unique(trigger_type_mapping.values())}

  # First pass to calculate others
  for framework, trigger_types in list(trigger_types_per_framework.items())[::-1]:
    for trigger_type, count in trigger_types.items():
      if framework not in labeled_frameworks_per_trigger[trigger_type]:
        offset_per_trigger_type[trigger_type] -= count / num_functions_per_trigger_type[trigger_type]

  # Second pass to calculate others
  for framework, trigger_types in list(trigger_types_per_framework.items())[::-1]:
    trigger_types_count = {x: 0 for x in unique(trigger_type_mapping.values())}

    for trigger_type, count in trigger_types.items():
      if framework in labeled_frameworks_per_trigger[trigger_type]:
        trigger_types_count[trigger_type] = count
      else:
        trigger_types_count[trigger_type] = 0

    df = pd.DataFrame({
      "TriggerType": trigger_types_count.keys(),
      "Percentage": [offset_per_trigger_type[trigger_type] for trigger_type in trigger_types_count.keys()],
    })

    sns.barplot(x="Percentage", y="TriggerType", data=df, color=colors[color_offset], orient="h", ax=ax, label=framework, order=unique(trigger_type_mapping.values()))
    if sum(trigger_types_count.values()) > 0:
      color_offset += 1

    for trigger_type in trigger_types_count.keys():
      offset_per_trigger_type[trigger_type] -= trigger_types_count[trigger_type] / num_functions_per_trigger_type[trigger_type]

  handles, labels = ax.get_legend_handles_labels()
  ax.legend(handles[::-1], labels[::-1], title='Line', loc='upper left')
  # sns.move_legend(ax, "upper left", bbox_to_anchor=(0, -0.3))

  sns.move_legend(
    ax, "lower left",
    bbox_to_anchor=(0, 1),
    ncol=3,
    title=None, frameon=False,
  )

  ax.xaxis.set_major_locator(ticker.MultipleLocator(0.2))

  plt.xlabel("Distribution [%]")
  plt.ylabel("")

  plt.xlim(0, 1)

  ax.xaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: f"{x*100:.0f}"))
  ax.xaxis.set_minor_formatter(ticker.NullFormatter())

  # plt.show()
  plt.savefig(f"{out_dir}/frameworks_per_trigger_type.png", transparent=True, bbox_inches="tight", pad_inches=1)
  plt.close()

def plot_frameworks_per_execution_location(repos, max_per_bar, out_dir):
  execution_locations_per_framework = {x: {y: 0 for y in unique(execution_location_mapping.values())} for x in unique(framework_category_mapping.values())}
  frameworks_per_execution_location = {x: {y: 0 for y in unique(framework_category_mapping.values())} for x in unique(execution_location_mapping.values())}

  num_functions_per_execution_location = {x: 0 for x in unique(execution_location_mapping.values())}

  for repo in repos:
    functions = repo.get("Functions", [])

    for function in functions:
      execution_location = function.get("Location", "")
      mapped_execution_location = execution_location_mapping.get(execution_location, execution_location)

      framework = function.get("Framework", "")
      mapped_framework = framework_category_mapping.get(framework, framework)

      execution_locations_per_framework[mapped_framework][mapped_execution_location] += 1
      frameworks_per_execution_location[mapped_execution_location][mapped_framework] += 1
      num_functions_per_execution_location[mapped_execution_location] += 1

  has_other = False

  labeled_frameworks_per_execution_location = {x: [] for x in unique(execution_location_mapping.values())}

  for execution_location, frameworks in frameworks_per_execution_location.items():
    frameworks = sorted(frameworks.items(), key=lambda x: x[1], reverse=True)
    if len(frameworks) > max_per_bar:
      has_other = True
      labeled_frameworks = [k for k, v in frameworks[:max_per_bar-1]]
      labeled_frameworks_per_execution_location[execution_location].extend(labeled_frameworks)
    else:
      labeled_frameworks = [k for k, v in frameworks]
      labeled_frameworks_per_execution_location[execution_location].extend(labeled_frameworks)

  fig, ax = plt.subplots(figsize=(10, 6))

  if has_other:
    df = pd.DataFrame({
      "ExecutionLocation": unique(execution_location_mapping.values()),
      "Percentage": [1.0 for _ in unique(execution_location_mapping.values())],
    })
    sns.barplot(x="Percentage", y="ExecutionLocation", data=df, orient="h", ax=ax, label="Other", order=unique(execution_location_mapping.values()))

  offset_per_execution_location = {x: 1.0 for x in unique(execution_location_mapping.values())}

  # First pass to calculate others
  for framework, execution_locations in list(execution_locations_per_framework.items())[::-1]:
    for execution_location, count in execution_locations.items():
      if framework not in labeled_frameworks_per_execution_location[execution_location]:
        offset_per_execution_location[execution_location] -= count / num_functions_per_execution_location[execution_location]

  # Second pass to calculate others
  for framework, execution_locations in list(execution_locations_per_framework.items())[::-1]:
    execution_locations_count = {x: 0 for x in unique(execution_location_mapping.values())}

    for execution_location, count in execution_locations.items():
      if framework in labeled_frameworks_per_execution_location[execution_location]:
        execution_locations_count[execution_location] = count
      else:
        execution_locations_count[execution_location] = 0

    df = pd.DataFrame({
      "TriggerType": execution_locations_count.keys(),
      "Percentage": [offset_per_execution_location[execution_location] for execution_location in execution_locations_count.keys()],
    })

    sns.barplot(x="Percentage", y="TriggerType", data=df, orient="h", ax=ax, label=framework, order=unique(execution_location_mapping.values()))

    for execution_location in execution_locations_count.keys():
      offset_per_execution_location[execution_location] -= execution_locations_count[execution_location] / num_functions_per_execution_location[execution_location]

  handles, labels = ax.get_legend_handles_labels()
  ax.legend(handles[::-1], labels[::-1], title='Line', loc='upper left')
  # sns.move_legend(ax, "upper left", bbox_to_anchor=(0, -0.3))

  sns.move_legend(
    ax, "lower left",
    bbox_to_anchor=(0, 1), ncol=3,
    title=None, frameon=False,
  )

  ax.xaxis.set_major_locator(ticker.MultipleLocator(0.2))

  plt.xlabel("Distribution [%] of Function Frameworks by Category")
  plt.ylabel("Execution Locations")

  plt.xlim(0, 1)

  ax.xaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: f"{x*100:.0f}"))
  ax.xaxis.set_minor_formatter(ticker.NullFormatter())

  # plt.show()
  plt.savefig(f"{out_dir}/frameworks_per_execution_location.png", transparent=True, bbox_inches="tight", pad_inches=1)
  plt.close()

#####################
### Trigger Types ###
#####################

def plot_trigger_types_per_function(repos, out_dir):
  trigger_type_count = {}
  num_functions = 0

  for repo in repos:
    functions = repo.get("Functions", [])
    for function in functions:
        num_functions += 1

        trigger_type = function.get("InvocationType", "")
        mapped_trigger_type = trigger_type_mapping.get(trigger_type, trigger_type)

        if mapped_trigger_type not in trigger_type_count:
          trigger_type_count[mapped_trigger_type] = 1
        else:
          trigger_type_count[mapped_trigger_type] += 1

  data = {
    "TriggerTypes": trigger_type_count.keys(),
    "Percentage": [x / num_functions for x in trigger_type_count.values()],
  }

  df = pd.DataFrame(data)
  df['TriggerTypes'] = pd.Categorical(df['TriggerTypes'], categories=unique(trigger_type_mapping.values()), ordered=True)
  df = df.sort_values('TriggerTypes')

  fig, ax = plt.subplots(figsize=(10, 6))
  sns.barplot(x="Percentage", y="TriggerTypes", data=df, orient="h", ax=ax)

  ax.xaxis.set_major_locator(ticker.MultipleLocator(0.05))

  for index, value in enumerate(df["Percentage"]):
    plt.text(value + 0.0025, index, f'{value:.2%}', va="center")

  plt.xlim(0, max(df["Percentage"]) * 1.1)

  plt.xlabel("Percentage of Functions")
  plt.ylabel("Trigger Types")

  ax.xaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: ""))
  ax.xaxis.set_minor_formatter(ticker.NullFormatter())

  plt.savefig(f"{out_dir}/trigger_types_per_function.png", transparent=True, bbox_inches='tight')
  plt.close()

def plot_trigger_types_per_application(repos, out_dir):
  trigger_type_count = {x: 0 for x in unique(trigger_type_mapping.values())}
  num_repos = len(repos)

  for repo in repos:
    functions = repo.get("Functions", [])

    trigger_types = {}
    for function in functions:
      trigger_type = function.get("InvocationType", "")
      mapped_trigger_type = trigger_type_mapping.get(trigger_type, trigger_type)
      trigger_types[mapped_trigger_type] = True
    
    for trigger_type in trigger_types.keys():
      if trigger_type in trigger_type_count:
        trigger_type_count[trigger_type] += 1
      else:
        trigger_type_count[trigger_type] = 1

  data = {
    "TriggerTypes": trigger_type_count.keys(),
    "Percentage": [x / num_repos for x in trigger_type_count.values()],
  }

  df = pd.DataFrame(data)
  df['TriggerTypes'] = pd.Categorical(df['TriggerTypes'], categories=unique(trigger_type_mapping.values()), ordered=True)
  df = df.sort_values('TriggerTypes')

  fig, ax = plt.subplots(figsize=(10, 6))
  sns.barplot(x="Percentage", y="TriggerTypes", data=df, orient="h", ax=ax)

  ax.xaxis.set_major_locator(ticker.MultipleLocator(0.05))

  for index, value in enumerate(df["Percentage"]):
    plt.text(value + 0.0025, index, f'{value:.2%}', va="center")

  plt.xlim(0, max(df["Percentage"]) * 1.1)

  plt.xlabel("Percentage of Applications")
  plt.ylabel("Trigger Types")

  ax.xaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: ""))
  ax.xaxis.set_minor_formatter(ticker.NullFormatter())

  plt.savefig(f"{out_dir}/trigger_types_per_application.png", transparent=True, bbox_inches='tight')
  plt.close()

def plot_trigger_types_per_application_and_function(repos, out_dir):
  repo_trigger_type_count = {}
  function_trigger_type_count = {}
  num_repos = len(repos)
  num_functions = 0

  for repo in repos:
    repo_trigger_types = set()

    functions = repo.get("Functions", [])
    for function in functions:
      num_functions += 1

      trigger_type = function.get("InvocationType", "")
      repo_trigger_types.add(trigger_type)
      mapped_trigger_type = trigger_type_mapping.get(trigger_type, trigger_type)

      if mapped_trigger_type not in function_trigger_type_count:
        function_trigger_type_count[mapped_trigger_type] = 1
      else:
        function_trigger_type_count[mapped_trigger_type] += 1

    for trigger_type in repo_trigger_types:
      mapped_trigger_type = trigger_type_mapping.get(trigger_type, trigger_type)

      if mapped_trigger_type not in repo_trigger_type_count:
        repo_trigger_type_count[mapped_trigger_type] = 1
      else:
        repo_trigger_type_count[mapped_trigger_type] += 1

  trigger_types = []
  percentages = []
  hues = []

  trigger_types.extend(function_trigger_type_count.keys())
  percentages.extend([x / num_functions for x in function_trigger_type_count.values()])
  hues.extend(["Functions"] * len(function_trigger_type_count.keys()))

  trigger_types.extend(repo_trigger_type_count.keys())
  percentages.extend([x / num_repos for x in repo_trigger_type_count.values()])
  hues.extend(["Applications"] * len(repo_trigger_type_count.keys()))

  data = {
    "TriggerTypes": trigger_types,
    "Percentage": percentages,
    "Hue": hues,
  }

  order = unique(trigger_type_mapping.values())
  hue_order = ["Functions", "Applications"]

  df = pd.DataFrame(data)
  df["TriggerTypes"] = pd.Categorical(df["TriggerTypes"], categories=order, ordered=True)
  df = df.sort_values("TriggerTypes").reset_index()

  fig, ax = plt.subplots(figsize=(6, 3))
  sns.barplot(x="Percentage", y="TriggerTypes", hue="Hue", order=order, hue_order=hue_order, data=df, orient="h", ax=ax)

  sns.move_legend(
    ax, "lower left",
    bbox_to_anchor=(0, 1),
    ncol=2,
    title=None,
    frameon=False,
  )

  ax.xaxis.set_major_locator(ticker.MultipleLocator(0.05))

  for index, row in df.iterrows():
    percentage = row["Percentage"]
    trigger_type = row["TriggerTypes"]
    hue = row["Hue"]
    if hue == "Functions":
      plt.text(percentage + 0.0025, (index//2)-0.2, f"{percentage:.2%}", va="center", fontsize=8)
    else:
      plt.text(percentage + 0.0025, (index//2)+0.2, f"{percentage:.2%}", va="center", fontsize=8)

  plt.xlim(0, max(df["Percentage"]) * 1.15)

  plt.xlabel("Proportion [%]")
  plt.ylabel("")

  ax.xaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: ""))
  ax.xaxis.set_minor_formatter(ticker.NullFormatter())

  plt.savefig(f"{out_dir}/trigger_types_per_application_and_function.png", transparent=True, bbox_inches='tight')
  plt.close()

def plot_trigger_types_matrix(repos, out_dir):
  num_trigger_type_combinations = {}

  for repo in repos:
    repo_trigger_types = set()

    functions = repo.get("Functions", [])
    for function in functions:
      trigger_type = function["InvocationType"]
      mapped_trigger_type = trigger_type_mapping.get(trigger_type, trigger_type)

      repo_trigger_types.add(mapped_trigger_type)

    repo_trigger_types = list(repo_trigger_types)
    repo_trigger_types_combinations = itertools.combinations(repo_trigger_types, 2)

    for left, right in repo_trigger_types_combinations:
      key = ";".join(sort_respecting([left, right], unique(trigger_type_mapping.values())))

      if key not in num_trigger_type_combinations:
        num_trigger_type_combinations[key] = 0
      num_trigger_type_combinations[key]  += 1

  row_trigger_types = []
  column_trigger_types = []
  num_repositories = []

  for k, v in num_trigger_type_combinations.items():
    left, right = k.split(";")

    row_trigger_types.append(right)
    column_trigger_types.append(left)
    num_repositories.append(v)

  data = {
    "Row": row_trigger_types,
    "Column": column_trigger_types,
    "Value": num_repositories,
  }

  df = pd.DataFrame(data)
  df = df.pivot(index="Row", columns="Column", values="Value")
  df = df.reindex(index=unique(trigger_type_mapping.values()), columns=unique(trigger_type_mapping.values()))
  df = df.drop([unique(trigger_type_mapping.values())[0]])
  df = df.drop(columns=[unique(trigger_type_mapping.values())[-1]])
  df = df.dropna(axis="columns", how="all")
  df = df.dropna(axis="rows", how="all")

  mask = df.to_numpy()
  mask = np.arange(mask.shape[0])[:,None] < np.arange(mask.shape[1])

  steps = [1, 2, 5, 10]
  for step in steps:
    if max(num_repositories) / step <= 11:
      break

  ticks = [0]
  tick = 0
  while tick < max(num_repositories):
    tick = tick + step
    ticks.append(tick)

  fig, ax = plt.subplots(figsize=(6, 3))

  sns.heatmap(df, annot=True, fmt=".0f", cmap="Blues", cbar=False, square=True, mask=mask, vmin=0, vmax=max(num_repositories), cbar_kws={"ticks": ticks})

  plt.xticks(rotation=45, ha='right')

  ax.spines[:].set_visible(True)

  #ax.xaxis.set_major_locator(ticker.MultipleLocator(0.05))

  #plt.xlim(0, max(df["Percentage"]) * 1.1)

  plt.xlabel("")
  plt.ylabel("")

  # ax.xaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: ""))
  # ax.xaxis.set_minor_formatter(ticker.NullFormatter())

  plt.savefig(f"{out_dir}/trigger_types_matrix.png", transparent=True, bbox_inches='tight')
  plt.close()

def plot_trigger_types_per_execution_location(repos, max_per_bar, out_dir):
  trigger_types_per_execution_location = {x: {y: 0 for y in unique(trigger_type_mapping.values())} for x in unique(execution_location_mapping.values())}
  execution_locations_per_trigger_type = {x: {y: 0 for y in unique(execution_location_mapping.values())} for x in unique(trigger_type_mapping.values())}

  num_functions_per_execution_location = {x: 0 for x in unique(execution_location_mapping.values())}

  for repo in repos:
    functions = repo.get("Functions", [])

    for function in functions:
      trigger_type = function.get("InvocationType", "")
      mapped_trigger_type = trigger_type_mapping.get(trigger_type, trigger_type)

      execution_location = function.get("Location", "")
      mapped_execution_location = execution_location_mapping.get(execution_location, execution_location)

      trigger_types_per_execution_location[mapped_execution_location][mapped_trigger_type] += 1
      execution_locations_per_trigger_type[mapped_trigger_type][mapped_execution_location] += 1

      num_functions_per_execution_location[mapped_execution_location] += 1

  has_other = False

  labeled_trigger_types_per_execution_location = {x: [] for x in unique(execution_location_mapping.values())}
  for execution_location, trigger_types in trigger_types_per_execution_location.items():
    trigger_types = sorted(trigger_types.items(), key=lambda x: x[1], reverse=True)
    if len(trigger_types) > max_per_bar:
      has_other = True
      labeled_trigger_types = [k for k, v in trigger_types[:max_per_bar-1]]
      labeled_trigger_types_per_execution_location[execution_location].extend(labeled_trigger_types)
    else:
      labeled_trigger_types = [k for k, v in trigger_types]
      labeled_trigger_types_per_execution_location[execution_location].extend(labeled_trigger_types)

  fig, ax = plt.subplots(figsize=(10, 6))

  if has_other:
    df = pd.DataFrame({
      "Location": unique(execution_location_mapping.values()),
      "Percentage": [1.0 for _ in unique(execution_location_mapping.values())]
    })
    sns.barplot(x="Percentage", y="Location", data=df, orient="h", ax=ax, label="Other", order=unique(execution_location_mapping.values()))

  offset_per_execution_location = {x: 1.0 for x in unique(execution_location_mapping.values())}

  # First pass to calculate others
  for trigger_type, execution_locations in list(execution_locations_per_trigger_type.items())[::-1]:
    for execution_location, count in execution_locations.items():
      if trigger_type not in labeled_trigger_types_per_execution_location[execution_location]:
        offset_per_execution_location[execution_location] -= count / num_functions_per_execution_location[execution_location]

  # Second pass to calculate others
  for trigger_type, execution_locations in list(execution_locations_per_trigger_type.items())[::-1]:
    execution_locations_count = {x: 0 for x in unique(execution_location_mapping.values())}

    for execution_location, count in execution_locations.items():
      if trigger_type in labeled_trigger_types_per_execution_location[execution_location]:
        execution_locations_count[execution_location] = count
      else:
        execution_locations_count[execution_location] = 0

    df = pd.DataFrame({
      "Location": execution_locations_count.keys(),
      "Percentage": [offset_per_execution_location[execution_location] for execution_location in execution_locations_count.keys()],
    })

    sns.barplot(x="Percentage", y="Location", data=df, orient="h", ax=ax, label=trigger_type, order=unique(execution_location_mapping.values()))

    for execution_location in execution_locations_count.keys():
      offset_per_execution_location[execution_location] -= execution_locations_count[execution_location] / num_functions_per_execution_location[execution_location]

  handles, labels = ax.get_legend_handles_labels()
  ax.legend(handles[::-1], labels[::-1], title='Line', loc='upper left')

  sns.move_legend(
    ax, "lower left",
    bbox_to_anchor=(0, 1), ncol=3,
    title=None, frameon=False,
  )

  ax.xaxis.set_major_locator(ticker.MultipleLocator(0.2))

  plt.xlabel("Distribution [%] of Trigger Types")
  plt.ylabel("Execution Location")

  plt.xlim(0, 1)

  ax.xaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: f"{x*100:.0f}"))
  ax.xaxis.set_minor_formatter(ticker.NullFormatter())

  plt.savefig(f"{out_dir}/trigger_types_per_execution_location.png", transparent=True, bbox_inches="tight", pad_inches=1)
  plt.close()

def plot_trigger_types_per_framework(repos, max_per_bar, out_dir):
  trigger_types_per_framework = {x: {y: 0 for y in unique(trigger_type_mapping.values())} for x in unique(framework_mapping.values())}
  frameworks_per_trigger_type = {x: {y: 0 for y in unique(framework_mapping.values())} for x in unique(trigger_type_mapping.values())}

  num_functions_per_framework = {x: 0 for x in unique(framework_mapping.values())}

  for repo in repos:
    functions = repo.get("Functions", [])

    for function in functions:
      trigger_type = function.get("InvocationType", "")
      mapped_trigger_type = trigger_type_mapping.get(trigger_type, trigger_type)

      framework = function.get("Framework", "")
      mapped_framework = framework_mapping.get(framework, framework)

      trigger_types_per_framework[mapped_framework][mapped_trigger_type] += 1
      frameworks_per_trigger_type[mapped_trigger_type][mapped_framework] += 1

      num_functions_per_framework[mapped_framework] += 1

  has_other = False

  labeled_trigger_types_per_framework = {x: [] for x in unique(framework_mapping.values())}
  for framework, trigger_types in trigger_types_per_framework.items():
    trigger_types = sorted(trigger_types.items(), key=lambda x: x[1], reverse=True)
    if len(trigger_types) > max_per_bar:
      has_other = True
      labeled_trigger_types = [k for k, v in trigger_types[:max_per_bar-1]]
      labeled_trigger_types_per_framework[framework].extend(labeled_trigger_types)
    else:
      labeled_trigger_types = [k for k, v in trigger_types]
      labeled_trigger_types_per_framework[framework].extend(labeled_trigger_types)

  fig, ax = plt.subplots(figsize=(10, 6))

  if has_other:
    df = pd.DataFrame({
      "Framework": unique(framework_mapping.values()),
      "Percentage": [1.0 for _ in unique(framework_mapping.values())]
    })
    sns.barplot(x="Percentage", y="Framework", data=df, orient="h", ax=ax, label="Other", order=unique(framework_mapping.values()))

  offset_per_framework = {x: 1.0 for x in unique(framework_mapping.values())}

  # First pass to calculate others
  for trigger_type, frameworks in list(frameworks_per_trigger_type.items())[::-1]:
    for framework, count in frameworks.items():
      if trigger_type not in labeled_trigger_types_per_framework[framework]:
        offset_per_framework[framework] -= count / num_functions_per_framework[framework]

  # Second pass to calculate others
  for trigger_type, frameworks in list(frameworks_per_trigger_type.items())[::-1]:
    frameworks_count = {x: 0 for x in unique(framework_mapping.values())}

    for framework, count in frameworks.items():
      if trigger_type in labeled_trigger_types_per_framework[framework]:
        frameworks_count[framework] = count
      else:
        frameworks_count[framework] = 0

    df = pd.DataFrame({
      "Framework": frameworks_count.keys(),
      "Percentage": [offset_per_framework[framework] for framework in frameworks_count.keys()],
    })

    sns.barplot(x="Percentage", y="Framework", data=df, orient="h", ax=ax, label=trigger_type, order=unique(framework_mapping.values()))

    for framework in frameworks_count.keys():
      offset_per_framework[framework] -= frameworks_count[framework] / num_functions_per_framework[framework]

  handles, labels = ax.get_legend_handles_labels()
  ax.legend(handles[::-1], labels[::-1], title='Line', loc='upper left')

  sns.move_legend(
    ax, "lower left",
    bbox_to_anchor=(0, 1), ncol=3,
    title=None, frameon=False,
  )

  ax.xaxis.set_major_locator(ticker.MultipleLocator(0.2))

  plt.xlabel("Distribution [%] of Trigger Types")
  plt.ylabel("Framework")

  plt.xlim(0, 1)

  ax.xaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: f"{x*100:.0f}"))
  ax.xaxis.set_minor_formatter(ticker.NullFormatter())

  plt.savefig(f"{out_dir}/trigger_type_per_framework.png", transparent=True, bbox_inches="tight", pad_inches=1)
  plt.close()

def plot_trigger_types_per_platform(repos, max_per_bar, out_dir):
  trigger_types_per_platform = {x: {y: 0 for y in unique(trigger_type_mapping.values())} for x in unique(platform_mapping.values())}
  platforms_per_trigger_type = {x: {y: 0 for y in unique(platform_mapping.values())} for x in unique(trigger_type_mapping.values())}

  num_functions_per_platform = {x: 0 for x in unique(platform_mapping.values())}

  for repo in repos:
    functions = repo.get("Functions", [])

    for function in functions:
      trigger_type = function.get("InvocationType", "")
      mapped_trigger_type = trigger_type_mapping.get(trigger_type, trigger_type)

      platform = function.get("Platform", "")
      mapped_platform = platform_mapping.get(platform, platform)

      trigger_types_per_platform[mapped_platform][mapped_trigger_type] += 1
      platforms_per_trigger_type[mapped_trigger_type][mapped_platform] += 1

      num_functions_per_platform[mapped_platform] += 1

  has_other = False

  labeled_trigger_types_per_platform = {x: [] for x in unique(platform_mapping.values())}
  for platform, trigger_types in trigger_types_per_platform.items():
    trigger_types = sorted(trigger_types.items(), key=lambda x: x[1], reverse=True)
    if len(trigger_types) > max_per_bar:
      has_other = True
      labeled_trigger_types = [k for k, v in trigger_types[:max_per_bar-1]]
      labeled_trigger_types_per_platform[platform].extend(labeled_trigger_types)
    else:
      labeled_trigger_types = [k for k, v in trigger_types]
      labeled_trigger_types_per_platform[platform].extend(labeled_trigger_types)

  fig, ax = plt.subplots(figsize=(10, 6))

  if has_other:
    df = pd.DataFrame({
      "Platform": unique(platform_mapping.values()),
      "Percentage": [1.0 for _ in unique(platform_mapping.values())]
    })
    sns.barplot(x="Percentage", y="Platform", data=df, orient="h", ax=ax, label="Other", order=unique(platform_mapping.values()))

  offset_per_platform = {x: 1.0 for x in unique(platform_mapping.values())}

  # First pass to calculate others
  for trigger_type, platforms in list(platforms_per_trigger_type.items())[::-1]:
    for platform, count in platforms.items():
      if trigger_type not in labeled_trigger_types_per_platform[platform]:
        offset_per_platform[platform] -= count / num_functions_per_platform[platform]

  # Second pass to calculate others
  for trigger_type, platforms in list(platforms_per_trigger_type.items())[::-1]:
    platforms_count = {x: 0 for x in unique(platform_mapping.values())}

    for platform, count in platforms.items():
      if trigger_type in labeled_trigger_types_per_platform[platform]:
        platforms_count[platform] = count
      else:
        platforms_count[platform] = 0

    df = pd.DataFrame({
      "Platform": platforms_count.keys(),
      "Percentage": [offset_per_platform[platform] for platform in platforms_count.keys()],
    })

    sns.barplot(x="Percentage", y="Platform", data=df, orient="h", ax=ax, label=trigger_type, order=unique(platform_mapping.values()))

    for platform in platforms_count.keys():
      offset_per_platform[platform] -= platforms_count[platform] / num_functions_per_platform[platform]

  handles, labels = ax.get_legend_handles_labels()
  ax.legend(handles[::-1], labels[::-1], title='Line', loc='upper left')

  sns.move_legend(
    ax, "lower left",
    bbox_to_anchor=(0, 1), ncol=3,
    title=None, frameon=False,
  )

  ax.xaxis.set_major_locator(ticker.MultipleLocator(0.2))

  plt.xlabel("Distribution [%] of Trigger Types")
  plt.ylabel("Platforms")

  plt.xlim(0, 1)

  ax.xaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: f"{x*100:.0f}"))
  ax.xaxis.set_minor_formatter(ticker.NullFormatter())

  plt.savefig(f"{out_dir}/trigger_type_per_platform.png", transparent=True, bbox_inches="tight", pad_inches=1)
  plt.close()                                                                                                                                                                                                                                                                                                                                                                                                                                                             

###########################
### Execution Locations ###
###########################

def plot_execution_locations_per_function(repos, out_dir):
  execution_location_count = {}
  num_functions = 0

  for repo in repos:
    functions = repo.get("Functions", [])
    for function in functions:
        num_functions += 1

        execution_location = function.get("Location", "")
        mapped_execution_location = execution_location_mapping.get(execution_location, execution_location)

        if mapped_execution_location not in execution_location_count:
          execution_location_count[mapped_execution_location] = 1
        else:
          execution_location_count[mapped_execution_location] += 1

  data = {
    "ExecutionLocation": execution_location_count.keys(),
    "Percentage": [x / num_functions for x in execution_location_count.values()],
  }

  df = pd.DataFrame(data)
  df['ExecutionLocation'] = pd.Categorical(df['ExecutionLocation'], categories=unique(execution_location_mapping.values()), ordered=True)
  df = df.sort_values('ExecutionLocation')

  fig, ax = plt.subplots(figsize=(10, 6))
  sns.barplot(x="Percentage", y="ExecutionLocation", data=df, orient="h", ax=ax)

  ax.xaxis.set_major_locator(ticker.MultipleLocator(0.05))

  for index, value in enumerate(df["Percentage"]):
    plt.text(value + 0.0025, index, f'{value:.2%}', va="center")

  plt.xlim(0, max(df["Percentage"]) * 1.1)

  plt.xlabel("Percentage of Functions")
  plt.ylabel("Execution Location")

  ax.xaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: ""))
  ax.xaxis.set_minor_formatter(ticker.NullFormatter())

  plt.savefig(f"{out_dir}/execution_location_per_function.png", transparent=True, bbox_inches='tight')
  plt.close()

def plot_execution_locations_per_application(repos, out_dir):
  execution_location_count = {}
  num_repos = len(repos)

  for repo in repos:
    functions = repo.get("Functions", [])
    execution_locations = {}

    for function in functions:
      execution_location = function.get("Location", "")
      mapped_execution_location = execution_location_mapping.get(execution_location, execution_location)
      execution_locations[mapped_execution_location] = True

    for execution_location in execution_locations.keys():
      if execution_location in execution_location_count:
        execution_location_count[execution_location] += 1
      else:
        execution_location_count[execution_location] = 1

  data = {
    "ExecutionLocation": execution_location_count.keys(),
    "Percentage": [x / num_repos for x in execution_location_count.values()],
  }

  df = pd.DataFrame(data)
  df['ExecutionLocation'] = pd.Categorical(df['ExecutionLocation'], categories=unique(execution_location_mapping.values()), ordered=True)
  df = df.sort_values('ExecutionLocation')

  fig, ax = plt.subplots(figsize=(10, 6))
  sns.barplot(x="Percentage", y="ExecutionLocation", data=df, orient="h", ax=ax)

  ax.xaxis.set_major_locator(ticker.MultipleLocator(0.05))

  for index, value in enumerate(df["Percentage"]):
    plt.text(value + 0.0025, index, f'{value:.2%}', va="center")

  plt.xlim(0, max(df["Percentage"]) * 1.1)

  plt.xlabel("Percentage of Applications")
  plt.ylabel("Execution Location")

  ax.xaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: ""))
  ax.xaxis.set_minor_formatter(ticker.NullFormatter())

  plt.savefig(f"{out_dir}/execution_location_per_application.png", transparent=True, bbox_inches='tight')
  plt.close()

def plot_execution_locations_per_application_and_function(repos, out_dir):
  repo_execution_location_count = {}
  function_execution_location_count = {}
  num_repos = len(repos)
  num_functions = 0

  for repo in repos:
    repo_execution_locations = set()

    functions = repo.get("Functions", [])
    for function in functions:
      num_functions += 1

      execution_location = function.get("Location", "")
      repo_execution_locations.add(execution_location)
      mapped_execution_location = execution_location_mapping.get(execution_location, execution_location)

      if mapped_execution_location not in function_execution_location_count:
        function_execution_location_count[mapped_execution_location] = 1
      else:
        function_execution_location_count[mapped_execution_location] += 1

    for execution_location in repo_execution_locations:
      mapped_execution_location = execution_location_mapping.get(execution_location, execution_location)

      if mapped_execution_location not in repo_execution_location_count:
        repo_execution_location_count[mapped_execution_location] = 1
      else:
        repo_execution_location_count[mapped_execution_location] += 1

  execution_locations = []
  percentages = []
  hues = []

  execution_locations.extend(function_execution_location_count.keys())
  percentages.extend([x / num_functions for x in function_execution_location_count.values()])
  hues.extend(["Functions"] * len(function_execution_location_count.keys()))

  execution_locations.extend(repo_execution_location_count.keys())
  percentages.extend([x / num_repos for x in repo_execution_location_count.values()])
  hues.extend(["Applications"] * len(repo_execution_location_count.keys()))

  data = {
    "Location": execution_locations,
    "Percentage": percentages,
    "Hue": hues,
  }

  order = unique(execution_location_mapping.values())
  hue_order = ["Functions", "Applications"]

  df = pd.DataFrame(data)
  df["Location"] = pd.Categorical(df["Location"], categories=order, ordered=True)
  df = df.sort_values("Location").reset_index()

  fig, ax = plt.subplots(figsize=(6, 2))
  sns.barplot(x="Percentage", y="Location", hue="Hue", order=order, hue_order=hue_order, data=df, orient="h", ax=ax)

  sns.move_legend(
    ax, "lower left",
    bbox_to_anchor=(0, 1),
    ncol=2,
    title=None,
    frameon=False,
  )

  ax.xaxis.set_major_locator(ticker.MultipleLocator(0.05))

  for index, row in df.iterrows():
    percentage = row["Percentage"]
    execution_location = row["Location"]
    hue = row["Hue"]
    if hue == "Functions":
      plt.text(percentage + 0.0025, (index//2)-0.2, f"{percentage:.2%}", va="center", fontsize=8)
    else:
      plt.text(percentage + 0.0025, (index//2)+0.2, f"{percentage:.2%}", va="center", fontsize=8)

  plt.xlim(0, max(df["Percentage"]) * 1.15)

  plt.xlabel("Proportion [%]")
  plt.ylabel("")

  ax.xaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: ""))
  ax.xaxis.set_minor_formatter(ticker.NullFormatter())

  plt.savefig(f"{out_dir}/execution_locations_per_application_and_function.png", transparent=True, bbox_inches='tight')
  plt.close()

def plot_execution_locations_matrix(repos, out_dir):
  num_execution_location_combinations = {}

  for repo in repos:
    repo_execution_locations = set()

    functions = repo.get("Functions", [])
    for function in functions:
      execution_location = function["Location"]
      mapped_execution_location = execution_location_mapping.get(execution_location, execution_location)

      repo_execution_locations.add(mapped_execution_location)

    repo_execution_locations = list(repo_execution_locations)
    repo_execution_locations_combinations = itertools.combinations(repo_execution_locations, 2)

    for left, right in repo_execution_locations_combinations:
      key = ";".join(sort_respecting([left, right], unique(execution_location_mapping.values())))

      if key not in num_execution_location_combinations:
        num_execution_location_combinations[key] = 0
      num_execution_location_combinations[key]  += 1

  row_execution_locations = []
  column_execution_locations = []
  num_repositories = []

  for k, v in num_execution_location_combinations.items():
    left, right = k.split(";")

    row_execution_locations.append(right)
    column_execution_locations.append(left)
    num_repositories.append(v)

  data = {
    "Row": row_execution_locations,
    "Column": column_execution_locations,
    "Value": num_repositories,
  }

  df = pd.DataFrame(data)
  df = df.pivot(index="Row", columns="Column", values="Value")
  df = df.reindex(index=unique(execution_location_mapping.values()), columns=unique(execution_location_mapping.values()))
  df = df.drop([unique(execution_location_mapping.values())[0]])
  df = df.drop(columns=[unique(execution_location_mapping.values())[-1]])
  df = df.dropna(axis="columns", how="all")
  df = df.dropna(axis="rows", how="all")

  mask = df.to_numpy()
  mask = np.arange(mask.shape[0])[:,None] < np.arange(mask.shape[1])

  steps = [1, 2, 5, 10]
  for step in steps:
    if max(num_repositories) / step <= 11:
      break

  ticks = [0]
  tick = 0
  while tick < max(num_repositories):
    tick = tick + step
    ticks.append(tick)

  fig, ax = plt.subplots(figsize=(6, 3))

  sns.heatmap(df, annot=True, fmt=".0f", cmap="Blues", cbar=False, square=True, mask=mask, vmin=0, vmax=max(num_repositories), cbar_kws={"ticks": ticks})

  ax.spines[:].set_visible(True)

  #ax.xaxis.set_major_locator(ticker.MultipleLocator(0.05))

  #plt.xlim(0, max(df["Percentage"]) * 1.1)

  plt.xlabel("")
  plt.ylabel("")

  # ax.xaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: ""))
  # ax.xaxis.set_minor_formatter(ticker.NullFormatter())

  plt.savefig(f"{out_dir}/execution_locations_matrix.png", transparent=True, bbox_inches='tight')
  plt.close()

def plot_execution_locations_per_trigger_type(repos, max_per_bar, out_dir):
  execution_locations_per_trigger_type = {x: {y: 0 for y in unique(execution_location_mapping.values())} for x in unique(trigger_type_mapping.values())}
  trigger_types_per_execution_location = {x: {y: 0 for y in unique(trigger_type_mapping.values())} for x in unique(execution_location_mapping.values())}

  num_functions_per_trigger_type = {x: 0 for x in unique(trigger_type_mapping.values())}

  for repo in repos:
    functions = repo.get("Functions", [])

    for function in functions:
      execution_location = function.get("Location", "")
      mapped_execution_location = execution_location_mapping.get(execution_location, execution_location)

      trigger_type = function.get("InvocationType", "")
      mapped_trigger_type = trigger_type_mapping.get(trigger_type, trigger_type)

      execution_locations_per_trigger_type[mapped_trigger_type][mapped_execution_location] += 1
      trigger_types_per_execution_location[mapped_execution_location][mapped_trigger_type] += 1

      num_functions_per_trigger_type[mapped_trigger_type] += 1

  has_other = False

  labeled_execution_locations_per_trigger_type = {x: [] for x in unique(trigger_type_mapping.values())}
  for trigger_type, execution_locations in execution_locations_per_trigger_type.items():
    execution_locations = sorted(execution_locations.items(), key=lambda x: x[1], reverse=True)
    if len(execution_locations) > max_per_bar:
      has_other = True
      labeled_execution_locations = [k for k, v in execution_locations[:max_per_bar-1]]
      labeled_execution_locations_per_trigger_type[trigger_type].extend(labeled_execution_locations)
    else:
      labeled_execution_locations = [k for k, v in execution_locations]
      labeled_execution_locations_per_trigger_type[trigger_type].extend(labeled_execution_locations)


  num_colors = 0
  if has_other:
    num_colors += 1
  for execution_location, trigger_types in trigger_types_per_execution_location.items():
    if sum(trigger_types.values()) > 0:
      num_colors += 1

  current_palette = sns.color_palette()
  colors = current_palette[:num_colors][::-1]

  color_offset = 0

  fig, ax = plt.subplots(figsize=(6, 3))

  if has_other:
    df = pd.DataFrame({
      "TriggerType": unique(trigger_type_mapping.values()),
      "Percentage": [1.0 for _ in unique(trigger_type_mapping.values())]
    })
    sns.barplot(x="Percentage", y="TriggerType", data=df, orient="h", ax=ax, label="Other", color=colors[color_offset], order=unique(trigger_type_mapping.values()))
    color_offset += 1

  offset_per_trigger_type = {x: 1.0 for x in unique(trigger_type_mapping.values())}

  # First pass to calculate others
  for execution_location, trigger_types in list(trigger_types_per_execution_location.items())[::-1]:
    for trigger_type, count in trigger_types.items():
      if execution_location not in labeled_execution_locations_per_trigger_type[trigger_type]:
        offset_per_trigger_type[trigger_type] -= count / num_functions_per_trigger_type[trigger_type]

  # Second pass to calculate others
  for execution_location, trigger_types in list(trigger_types_per_execution_location.items())[::-1]:
    trigger_types_count = {x: 0 for x in unique(trigger_type_mapping.values())}

    for trigger_type, count in trigger_types.items():
      if execution_location in labeled_execution_locations_per_trigger_type[trigger_type]:
        trigger_types_count[trigger_type] = count
      else:
        trigger_types_count[trigger_type] = 0

    df = pd.DataFrame({
      "TriggerType": trigger_types_count.keys(),
      "Percentage": [offset_per_trigger_type[trigger_type] for trigger_type in trigger_types_count.keys()],
    })

    sns.barplot(x="Percentage", y="TriggerType", data=df, orient="h", ax=ax, color=colors[color_offset], label=execution_location, order=unique(trigger_type_mapping.values()))
    if sum(trigger_types_count.values()) > 0:
      color_offset += 1

    for trigger_type in trigger_types_count.keys():
      offset_per_trigger_type[trigger_type] -= trigger_types_count[trigger_type] / num_functions_per_trigger_type[trigger_type]

  handles, labels = ax.get_legend_handles_labels()
  ax.legend(handles[::-1], labels[::-1], title='Line', loc='upper left')

  sns.move_legend(
    ax, "lower left",
    bbox_to_anchor=(0, 1), ncol=3,
    title=None, frameon=False,
  )

  ax.xaxis.set_major_locator(ticker.MultipleLocator(0.2))

  plt.xlabel("Distribution [%]")
  plt.ylabel("")

  plt.xlim(0, 1)

  ax.xaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: f"{x*100:.0f}"))
  ax.xaxis.set_minor_formatter(ticker.NullFormatter())

  plt.savefig(f"{out_dir}/execution_location_per_trigger_type.png", transparent=True, bbox_inches="tight", pad_inches=1)
  plt.close()

def plot_execution_locations_per_framework(repos, max_per_bar, out_dir):
  execution_locations_per_framework = {x: {y: 0 for y in unique(execution_location_mapping.values())} for x in unique(framework_category_mapping.values())}
  frameworks_per_execution_location = {x: {y: 0 for y in unique(framework_category_mapping.values())} for x in unique(execution_location_mapping.values())}

  num_functions_per_framework = {x: 0 for x in unique(framework_category_mapping.values())}

  for repo in repos:
    functions = repo.get("Functions", [])

    for function in functions:
      execution_location = function.get("Location", "")
      mapped_execution_location = execution_location_mapping.get(execution_location, execution_location)

      framework = function.get("Framework", "")
      mapped_framework = framework_category_mapping.get(framework, framework)

      execution_locations_per_framework[mapped_framework][mapped_execution_location] += 1
      frameworks_per_execution_location[mapped_execution_location][mapped_framework] += 1

      num_functions_per_framework[mapped_framework] += 1

  has_other = False

  labeled_execution_locations_per_framework = {x: [] for x in unique(framework_category_mapping.values())}
  for framework, execution_locations in execution_locations_per_framework.items():
    execution_locations = sorted(execution_locations.items(), key=lambda x: x[1], reverse=True)
    if len(execution_locations) > max_per_bar:
      has_other = True
      labeled_execution_locations = [k for k, v in execution_locations[:max_per_bar-1]]
      labeled_execution_locations_per_framework[framework].extend(labeled_execution_locations)
    else:
      labeled_execution_locations = [k for k, v in execution_locations]
      labeled_execution_locations_per_framework[framework].extend(labeled_execution_locations)

  fig, ax = plt.subplots(figsize=(10, 6))

  if has_other:
    df = pd.DataFrame({
      "Framework": unique(framework_category_mapping.values()),
      "Percentage": [1.0 for _ in unique(framework_category_mapping.values())]
    })
    sns.barplot(x="Percentage", y="Framework", data=df, orient="h", ax=ax, label="Other", order=unique(framework_category_mapping.values()))

  offset_per_framework = {x: 1.0 for x in unique(framework_category_mapping.values())}

  # First pass to calculate others
  for execution_location, frameworks in list(frameworks_per_execution_location.items())[::-1]:
    for framework, count in frameworks.items():
      if execution_location not in labeled_execution_locations_per_framework[framework]:
        offset_per_framework[framework] -= count / num_functions_per_framework[framework]

  # Second pass to calculate others
  for execution_location, frameworks in list(frameworks_per_execution_location.items())[::-1]:
    frameworks_count = {x: 0 for x in unique(framework_category_mapping.values())}

    for framework, count in frameworks.items():
      if execution_location in labeled_execution_locations_per_framework[framework]:
        frameworks_count[framework] = count
      else:
        frameworks_count[framework] = 0

    df = pd.DataFrame({
      "TriggerType": frameworks_count.keys(),
      "Percentage": [offset_per_framework[framework] for framework in frameworks_count.keys()],
    })

    sns.barplot(x="Percentage", y="TriggerType", data=df, orient="h", ax=ax, label=execution_location, order=unique(framework_category_mapping.values()))

    for framework in frameworks_count.keys():
      offset_per_framework[framework] -= frameworks_count[framework] / num_functions_per_framework[framework]

  handles, labels = ax.get_legend_handles_labels()
  ax.legend(handles[::-1], labels[::-1], title='Line', loc='upper left')

  sns.move_legend(
    ax, "lower left",
    bbox_to_anchor=(0, 1), ncol=3,
    title=None, frameon=False,
  )

  ax.xaxis.set_major_locator(ticker.MultipleLocator(0.2))

  plt.xlabel("Distribution [%] of Execution Locations")
  plt.ylabel("Frameworks")

  plt.xlim(0, 1)

  ax.xaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: f"{x*100:.0f}"))
  ax.xaxis.set_minor_formatter(ticker.NullFormatter())

  plt.savefig(f"{out_dir}/execution_location_per_framework.png", transparent=True, bbox_inches="tight", pad_inches=1)
  plt.close()

def plot_execution_locations_per_platform(repos, max_per_bar, out_dir):
  execution_locations_per_platform = {x: {y: 0 for y in unique(execution_location_mapping.values())} for x in unique(platform_mapping.values())}
  platforms_per_execution_location = {x: {y: 0 for y in unique(platform_mapping.values())} for x in unique(execution_location_mapping.values())}

  num_functions_per_platform = {x: 0 for x in unique(platform_mapping.values())}

  for repo in repos:
    functions = repo.get("Functions", [])

    for function in functions:
      execution_location = function.get("Location", "")
      mapped_execution_location = execution_location_mapping.get(execution_location, execution_location)

      platform = function.get("Platform", "")
      mapped_platform = platform_mapping.get(platform, platform)

      execution_locations_per_platform[mapped_platform][mapped_execution_location] += 1
      platforms_per_execution_location[mapped_execution_location][mapped_platform] += 1

      num_functions_per_platform[mapped_platform] += 1

  has_other = False

  labeled_execution_locations_per_platform = {x: [] for x in unique(platform_mapping.values())}
  for platform, execution_locations in execution_locations_per_platform.items():
    execution_locations = sorted(execution_locations.items(), key=lambda x: x[1], reverse=True)
    if len(execution_locations) > max_per_bar:
      has_other = True
      labeled_execution_locations = [k for k, v in execution_locations[:max_per_bar-1]]
      labeled_execution_locations_per_platform[platform].extend(labeled_execution_locations)
    else:
      labeled_execution_locations = [k for k, v in execution_locations]
      labeled_execution_locations_per_platform[platform].extend(labeled_execution_locations)

  fig, ax = plt.subplots(figsize=(10, 6))

  if has_other:
    df = pd.DataFrame({
      "Platform": unique(platform_mapping.values()),
      "Percentage": [1.0 for _ in unique(platform_mapping.values())]
    })
    sns.barplot(x="Percentage", y="Platform", data=df, orient="h", ax=ax, label="Other", order=unique(platform_mapping.values()))

  offset_per_platform = {x: 1.0 for x in unique(platform_mapping.values())}

  # First pass to calculate others
  for execution_location, platforms in list(platforms_per_execution_location.items())[::-1]:
    for platform, count in platforms.items():
      if execution_location not in labeled_execution_locations_per_platform[platform]:
        offset_per_platform[platform] -= count / num_functions_per_platform[platform]

  # Second pass to calculate others
  for execution_location, platforms in list(platforms_per_execution_location.items())[::-1]:
    platforms_count = {x: 0 for x in unique(platform_mapping.values())}

    for platform, count in platforms.items():
      if execution_location in labeled_execution_locations_per_platform[platform]:
        platforms_count[platform] = count
      else:
        platforms_count[platform] = 0

    df = pd.DataFrame({
      "TriggerType": platforms_count.keys(),
      "Percentage": [offset_per_platform[platform] for platform in platforms_count.keys()],
    })

    sns.barplot(x="Percentage", y="TriggerType", data=df, orient="h", ax=ax, label=execution_location, order=unique(platform_mapping.values()))

    for platform in platforms_count.keys():
      offset_per_platform[platform] -= platforms_count[platform] / num_functions_per_platform[platform]

  handles, labels = ax.get_legend_handles_labels()
  ax.legend(handles[::-1], labels[::-1], title='Line', loc='upper left')

  sns.move_legend(
    ax, "lower left",
    bbox_to_anchor=(0, 1), ncol=3,
    title=None, frameon=False,
  )

  ax.xaxis.set_major_locator(ticker.MultipleLocator(0.2))

  plt.xlabel("Distribution [%] of Execution Locations")
  plt.ylabel("Platforms")

  plt.xlim(0, 1)

  ax.xaxis.set_major_formatter(ticker.FuncFormatter(lambda x,y: f"{x*100:.0f}"))
  ax.xaxis.set_minor_formatter(ticker.NullFormatter())

  plt.savefig(f"{out_dir}/execution_location_per_platform.png", transparent=True, bbox_inches="tight", pad_inches=1)
  plt.close()

def main():
  repos = None
  with open("data/exportedRepositories/faasRepositories.json", "r") as f:
    repos = json.loads(f.read())

  # Files
  plot_num_files_per_application(repos, out_dir="charts/files")
  plot_number_of_files_repository_distribution_v1(repos, out_dir="charts/files")
  plot_number_of_files_repository_distribution_v2(repos, out_dir="charts/files")
  plot_number_of_files_repository_distribution_v3(repos, out_dir="charts/files")
  plot_number_of_files_category_distribution(repos, out_dir="charts/files")
  plot_distribution_of_files_by_category_by_application_size(repos, out_dir="charts/files")

  # Lines of Code
  plot_loc_repository_distribution_v1(repos, out_dir="charts/linesOfCode")
  plot_loc_repository_distribution_v2(repos, out_dir="charts/linesOfCode")
  plot_loc_language_distribution(repos, out_dir="charts/linesOfCode")
  plot_loc_per_application(repos, out_dir="charts/linesOfCode")

  # Languages
  plot_languages_by_loc_files_and_applications(repos, out_dir="charts/languages")

  # Number of Functions
  plot_function_count_distribution(repos, 20, out_dir="charts/numFunctions")
  plot_num_functions_per_framework_category(repos, out_dir="charts/numFunctions")
  plot_num_functions_per_framework(repos, out_dir="charts/numFunctions")
  plot_num_functions_per_platform(repos, out_dir="charts/numFunctions")
  plot_num_functions_per_execution_location(repos, out_dir="charts/numFunctions")
  plot_num_functions_per_trigger_type(repos, out_dir="charts/numFunctions")
  plot_javascript_loc_and_num_functions(repos, out_dir="charts/numFunctions")
  plot_trigger_types_and_num_functions(repos, out_dir="charts/numFunctions")
  plot_execution_location_and_num_functions(repos, out_dir="charts/numFunctions")

  # Platforms
  print_num_platforms_per_application(repos)

  plot_platforms_per_application(repos, out_dir="charts/platforms")
  plot_platforms_per_function(repos, out_dir="charts/platforms")
  plot_platforms_per_application_and_function(repos, out_dir="charts/platforms")
  plot_platforms_matrix(repos, out_dir="charts/platforms")
  # plot_platforms_per_framework(repos, 5, out_dir="charts/platforms")
  plot_platforms_per_trigger_type(repos, 5, out_dir="charts/platforms")
  plot_platforms_per_execution_location(repos, 5, out_dir="charts/platforms")

  # Frameworks
  plot_frameworks_per_application(repos, out_dir="charts/frameworks")
  plot_frameworks_per_function(repos, out_dir="charts/frameworks")
  plot_frameworks_per_application_and_function(repos, out_dir="charts/frameworks")
  plot_frameworks_matrix(repos, out_dir="charts/frameworks")
  plot_framework_categories_matrix(repos, out_dir="charts/frameworks")
  plot_framework_categories_per_application_and_function(repos, out_dir="charts/frameworks")
  # plot_frameworks_per_platform(repos, 5, out_dir="charts/frameworks")
  plot_frameworks_per_trigger_type(repos, 5, out_dir="charts/frameworks")
  plot_frameworks_per_execution_location(repos, 5, out_dir="charts/frameworks")

  # Trigger Types
  plot_trigger_types_per_application(repos, out_dir="charts/triggerTypes")
  plot_trigger_types_per_function(repos, out_dir="charts/triggerTypes")
  plot_trigger_types_per_application_and_function(repos, out_dir="charts/triggerTypes")
  plot_trigger_types_matrix(repos, out_dir="charts/triggerTypes")
  plot_trigger_types_per_execution_location(repos, 5, out_dir="charts/triggerTypes")
  plot_trigger_types_per_framework(repos, 5, out_dir="charts/triggerTypes")
  plot_trigger_types_per_platform(repos, 5, out_dir="charts/triggerTypes")
  
  # Execution Location
  plot_execution_locations_per_application(repos, out_dir="charts/executionLocations")
  plot_execution_locations_per_function(repos, out_dir="charts/executionLocations")
  plot_execution_locations_per_application_and_function(repos, out_dir="charts/executionLocations")
  plot_execution_locations_matrix(repos, out_dir="charts/executionLocations")
  plot_execution_locations_per_trigger_type(repos, 5, out_dir="charts/executionLocations")
  plot_execution_locations_per_framework(repos, 5, out_dir="charts/executionLocations")
  plot_execution_locations_per_platform(repos, 5, out_dir="charts/executionLocations")

if __name__ == "__main__":
  main()