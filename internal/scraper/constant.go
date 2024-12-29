package scraper

// defines the maximum number of concurrent go routines
// this affects leetcode questions query and output file writes
// this also consequently loosely represent req/s
const MAX_GO_ROUTINES = 35

// proxies should in theory protect against IP bans, but it seems like they're not reliable (at least the free ones)
// anyways, for this leetcode in specific, it probably doesn't matter since we don't need that many queries
const USE_PROXY = true

// some proxies can cause timeouts, gateway errors, etc, so having a retry mechanism is useful
const MAX_RETRY = 5

// urls for proxies and leetcode-related queries
const (
	PROXIES_API_URL = "https://api.proxyscrape.com/v2/?request=displayproxies&protocol=http&timeout=1500&country=us,ca&ssl=all&anonymity=elite"
	LEETCODE_URL    = "https://leetcode.com/graphql/"
)

// where the html files of the leetcode problems will be stored
const QUESTIONS_OUTPUT_DIR = "problems/"

// the graphql queries for leetcode-related queries
const (
	QUESTIONS_COUNT_QUERY_STRING = "query problemsetQuestionList($categorySlug:String,$filters:QuestionListFilterInput){problemsetQuestionList:questionList(categorySlug:$categorySlug filters:$filters) {total:totalNum}}"
	QUESTIONS_QUERY_STRING       = "query problemsetQuestionList($categorySlug:String,$limit:Int,$skip:Int,$filters:QuestionListFilterInput){ problemsetQuestionList:questionList( categorySlug:$categorySlug limit:$limit skip:$skip filters:$filters){ questions:data{ frontendQuestionId:questionFrontendId titleSlug title difficulty hints topicTags{name} codeDefinition content}}}"
)

// in case we don't define a user-agents.txt
const DEFAULT_USER_AGENT = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
