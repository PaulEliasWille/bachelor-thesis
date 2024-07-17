// create a entrypoint for the application
package main

const (
	REPOSITORY_INFOS_DIRECTORY                             = "data/repositoryInfos"
	REPOSITORIES_DIRECTORY                                 = "data/repositories"
	RAW_REPOSITORY_EVENTS_DIRECTORY                        = "data/rawRepositoryEvents"
	REPOSITORY_EVENTS_DIRECTORY                            = "data/repositoryEvents"
	REPOSITORIES_ISSUES_COMMITS_AND_CONTRIBUTORS_DIRECTORY = "data/repositoryIssuesCommitsAndContributorsDirectory"
	REPOSITORIES_DATA_DIRECTORY                            = "data/repositoriesData"
	REPOSITORIES_EXPORT_DIRECTORY                          = "data/exportedRepositories"

	REPOSITORY_QUERIES_PATH             = "data/repositoryQueries.json"
	REPOSITORY_IDS_PATH                 = "data/repositoryIds.json"
	RELEVANT_REPOSITORY_IDS_PATH        = "data/relevantRepositoryIds.json"
	HIGHLY_RELEVANT_REPOSITORY_IDS_PATH = "data/highlyRelevantRepositoryIds.json"
)

var (
	DOCUMENTATION_FILES = []string{
		"readme.md",
		"README.md",
		"readme",
		"README",
	}
	KEYWORDS = []string{
		"serverless",
		"faas",
		"baas",
		"cold start",
		"lambda",
		"step function",
		"cloud run", // TODO: re-run
		"cloud function",
		"function compute", // TODO: re-run
		"azure function",
		"oracle function",
		"gcp function",
		"ibm function",
		"oracle fn",
		"cloud fn",
		"openwhisk",
		"fission", // TODO: re-run
		"kubeless",
		"openfaas", // TODO: re-run
		"nuclio",   // TODO: re-run
		"knative",  // TODO: re-run
		"fn project",
	}
	EXCLUDE_KEYWORDS = []string{
		"example",
		"demo",
		"tutorial",
		"playground",
		"learn",
		"teach",
		"exercise",
		"course",
		"practice",
		"template",
		"sample",
		"workshop",
		"lecture",
		"study",
		"boilerplate",
		"starter kit",
		"showcase",
		"framework",
		"library",
		"plugin",
	}
	EXCLUDE_DIRECTORIES = []string{
		"node_modules",
		"test",
		"demo",
		"example",
		"tutorial",
		"docs",
	}
)

func prepareRepositoryQueries() {
	language := "javascript"
	query := GitHubRepositoryQuery{
		CreatedAt: &DateRange{Date{2000, 1, 1}, Date{2024, 3, 1}},
		Stars:     &Range{5, 500_000},
		Language:  &language,
	}
	ChunkGitHubRepositoryQuery(query, 20, REPOSITORY_QUERIES_PATH)
}

func executeRepositoryQueries() {
	repositoryQueries, err := LoadRepositoryQueries(REPOSITORY_QUERIES_PATH)
	if err != nil {
		panic(err)
	}

	ScrapeGitHub(repositoryQueries, 20, REPOSITORY_INFOS_DIRECTORY)
}

func extractRepositoryIds() {
	if err := AggregateRepositoryIds(REPOSITORY_INFOS_DIRECTORY, REPOSITORY_IDS_PATH); err != nil {
		panic(err)
	}
}

func downloadRepositoryDocumentationFiles() {
	repositoryIds, err := LoadRepositoryIds(REPOSITORY_IDS_PATH)
	if err != nil {
		panic(err)
	}

	DownloadRepositoryFiles(
		20,
		repositoryIds,
		REPOSITORY_INFOS_DIRECTORY,
		DOCUMENTATION_FILES,
		REPOSITORIES_DIRECTORY,
		false,
	)
}

func downloadRepositoryEvents() {
	repositoryIds, err := LoadRepositoryIds(REPOSITORY_IDS_PATH)
	if err != nil {
		panic(err)
	}

	DownloadRepositoryEvents(
		20,
		repositoryIds,
		DateRange{Date{2015, 1, 1}, Date{2024, 3, 1}},
		KEYWORDS,
		RAW_REPOSITORY_EVENTS_DIRECTORY,
	)
}

func aggregateRepositoryEvents() {
	if err := AggregateRepositoryEvents(RAW_REPOSITORY_EVENTS_DIRECTORY, REPOSITORY_EVENTS_DIRECTORY); err != nil {
		panic(err)
	}
}

func filterRelevantRepositoryIds() {
	repositoryIds, err := LoadRepositoryIds(REPOSITORY_IDS_PATH)
	if err != nil {
		panic(err)
	}

	relevantRepositoryIds := FilterRelevantRepositoryIds(
		repositoryIds,
		KEYWORDS,
		EXCLUDE_KEYWORDS,
		DOCUMENTATION_FILES,
		REPOSITORY_INFOS_DIRECTORY,
		REPOSITORY_EVENTS_DIRECTORY,
		REPOSITORIES_DIRECTORY,
	)

	if err := SaveRepositoryIds(relevantRepositoryIds, RELEVANT_REPOSITORY_IDS_PATH); err != nil {
		panic(err)
	}
}

func cloneRelevantRepositories() {
	repositoryIds, err := LoadRepositoryIds(RELEVANT_REPOSITORY_IDS_PATH)
	if err != nil {
		panic(err)
	}

	if err := DownloadRepositories(repositoryIds, REPOSITORY_INFOS_DIRECTORY, REPOSITORIES_DIRECTORY); err != nil {
		panic(err)
	}
}

func downloadRelevantRepositoriesIssuesCommitsAndContributors() {
	repositoryIds, err := LoadRepositoryIds(RELEVANT_REPOSITORY_IDS_PATH)
	if err != nil {
		panic(err)
	}

	DownloadRepositoriesIssuesCommitsAndContributors(
		20,
		repositoryIds,
		REPOSITORY_INFOS_DIRECTORY,
		REPOSITORIES_ISSUES_COMMITS_AND_CONTRIBUTORS_DIRECTORY,
		false,
	)
}

func aggregateRelevantRepositoryData() {
	repositoryIds, err := LoadRepositoryIds(RELEVANT_REPOSITORY_IDS_PATH)
	if err != nil {
		panic(err)
	}

	AggregateRepositoriesData(
		20,
		repositoryIds,
		REPOSITORIES_DIRECTORY,
		REPOSITORY_INFOS_DIRECTORY,
		REPOSITORIES_ISSUES_COMMITS_AND_CONTRIBUTORS_DIRECTORY,
		EXCLUDE_DIRECTORIES,
		REPOSITORIES_DATA_DIRECTORY,
	)
}

func filterHighlyRelevantRepositories() {
	repositoryIds, err := LoadRepositoryIds(RELEVANT_REPOSITORY_IDS_PATH)
	if err != nil {
		panic(err)
	}

	manualRemovedRepositoryIds := []RepositoryId{
		174904499, // example
		47403260,  // example
		15363408,  // example
		95603023,  // example
		57147380,  // example
		99688826,  // example
		107663169, // example
		224018331, // example
		319244686, // example
		525651593, // example
		162722550, // example
		261166328, // example
		280929892, // example
		288110967, // example
		227072603, // example
		316015476, // example
		568696693, // example
		243334432, // example
		289366340, // example
		379907862, // example
		230054296, // example
		222734057, // example
		235677266, // example
		286450295, // example
		324201161, // example
		304344049, // framework
		275154725, // framework
		44249545,  // framework
		77491536,  // framework
		206197127, // framework
	}

	highlyRelevantRepositoryIds := FilterHighlyRelevantRepositoryIds(
		repositoryIds,
		manualRemovedRepositoryIds,
		REPOSITORIES_DATA_DIRECTORY,
	)

	if err := SaveRepositoryIds(highlyRelevantRepositoryIds, HIGHLY_RELEVANT_REPOSITORY_IDS_PATH); err != nil {
		panic(err)
	}
}

func exportRepositories() {
	highlyRelevantRepositoryIds, err := LoadRepositoryIds(HIGHLY_RELEVANT_REPOSITORY_IDS_PATH)
	if err != nil {
		panic(err)
	}

	if err := ExportRepositories(
		highlyRelevantRepositoryIds,
		REPOSITORIES_DATA_DIRECTORY,
		REPOSITORIES_EXPORT_DIRECTORY,
	); err != nil {
		panic(err)
	}
}

// main function
func main() {
	// prepareRepositoryQueries()
	// executeRepositoryQueries()
	// extractRepositoryIds()
	// downloadRepositoryDocumentationFiles()
	// downloadRepositoryEvents()
	// aggregateRepositoryEvents()
	// filterRelevantRepositoryIds()
	// cloneRelevantRepositories()
	// downloadRelevantRepositoriesIssuesCommitsAndContributors()
	aggregateRelevantRepositoryData()
	filterHighlyRelevantRepositories()
	exportRepositories()
	// 2768 -> 2651 -> 2383
	// 764 -> 647 -> 572 -> 354

	// 214 -> 222 -> 227

	// TODO
	// [✓] search for indicative npm packages (i.e. cloud functions framework, etc.)
	// [✓] search for serverless entry point (handler)
	// [✓] summarize highly relevant repositories
	// [✓] categorize highly relevant repositories
	// [✓] aggregate all the data so far into a csv for manual processing
	// [✓] apply stronger filtering constraints to reduce number of (highly) relevant repositories
	// [✓] remove ai summary and category
	// [✓] manual check hit rate of highly relevant repositories
	//     -> random sample of 50 repositories yielded 37 hits and 13 misses
	// [✓] add more exclude keywords for relevancy checks (from wonderless)
	// [✓] add topics and ~labels~ to search for exclude keywords
	//     -> labels require an api call, to avoid that we will skip it
	// [✓] remove repositories using highly specialized frameworks (i.e. vercel or netlify)
	//     -> no real benefit since use case is mostly specified by the framework
	// [✓] filter out "short lived" projects
	//     -> working on commit data requires api calls, so we use the createdAt and pushedAt fields
	// [✓] first extract data for relevant repositories, then filter for highly relevant repositories
	//     -> remove stupid abstraction layer "repositoryChecks"
	//     -> remove code duplication caused by that
	// [✓] check if files are in directory listed in exclude keywords
	// [✓] get commits, issues and contributors for highly relevant repositories
	// [✓] filter final repositories by commits, issues and contributors#
	// [✓] use package.json to identify applications in the repository
	// [✓] do high relevancy filtering on application and not on repository level
	// [✓] detect number of functions, used platform/framework, invocation type and location based on scans for specific
	//     frameworks and parsing of configuration/code files
	// [✓] exclude "demo", "test", "example" directories from analysis
	// [✓] iterate to find more frameworks (look at repositories without found frameworks)

	// 3. June
	// [✓] improve detection of function location (edge or region)
	// [✓] improve detection of function invocation type

	// 4. June
	// [✓] extract statistics about functions (providers, frameworks, edge, etc.)
	// [ ] identify areas that require manual analysis (e.g. most trigger types are unknown -> review trigger types)
	// [ ] plan presentation (what comes first, maybe prepare some slides)

	// 5. June
	// [ ] further work on presentation

}
