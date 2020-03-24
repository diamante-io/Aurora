package cmp

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

const fileLengthLimit = 100

var findResultMetaXDR = regexp.MustCompile(`"result_meta_xdr": "(.*)",`)

// removeRegexps contains a list of regular expressions that, when matched,
// will be changed to an empty string. This is done to exclude known
// differences in responses between two Aurora version.
//
// Let's say that next Aurora version adds a new bool field:
// `is_authorized` on account balances list. You want to remove this
// field so it's not reported for each `/accounts/{id}` response.
var removeRegexps = []*regexp.Regexp{
	regexp.MustCompile(`This (is ){0,1}usually`),
	// Removes joined transaction (join=transactions) added in Aurora 0.19.0.
	// Remove for future versions.
	regexp.MustCompile(`(?msU)"transaction":\s*{\s*("memo|"_links)[\n\s\S]*][\n\s\S]*}(,\s{9}|,)`),
	// regexp.MustCompile(`\s*"is_authorized": true,`),
	// regexp.MustCompile(`\s*"is_authorized": false,`),
	// regexp.MustCompile(`\s*"successful": true,`),
	// regexp.MustCompile(`\s*"transaction_count": [0-9]+,`),
	// regexp.MustCompile(`\s*"last_modified_ledger": [0-9]+,`),
	// regexp.MustCompile(`\s*"public_key": "G.*",`),
}

type Response struct {
	Domain string
	Path   string
	Stream bool

	StatusCode int
	Body       string
	// NormalizedBody is body without parts that identify a single
	// server (ex. domain) and fields known to be different between
	// instances (ex. `result_meta_xdr`).
	NormalizedBody string
}

func NewResponse(domain, path string, stream bool) *Response {
	response := &Response{
		Domain: domain,
		Path:   path,
		Stream: stream,
	}

	req, err := http.NewRequest("GET", domain+path, nil)
	if err != nil {
		panic(err)
	}

	client := &http.Client{}

	if stream {
		req.Header.Add("Accept", "text/event-stream")
		client.Timeout = 500 * time.Millisecond
	}

	resp, err := client.Do(req)
	if err != nil {
		response.Body = err.Error()
		response.NormalizedBody = err.Error()
		return response
	}

	response.StatusCode = resp.StatusCode

	if resp.StatusCode != http.StatusOK &&
		resp.StatusCode != http.StatusNotFound &&
		resp.StatusCode != http.StatusNotAcceptable {
		panic(resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	// We ignore the error below to timeout streaming requests.
	// net/http: request canceled (Client.Timeout exceeded while reading body)
	if err != nil && !stream {
		response.Body = err.Error()
		response.NormalizedBody = err.Error()
		return response
	}

	if string(body) == "" {
		panic("Empty body")
	}

	response.Body = string(body)

	normalizedBody := response.Body
	// `result_meta_xdr` can differ between core instances (confirmed this with core team)
	normalizedBody = findResultMetaXDR.ReplaceAllString(normalizedBody, "")
	// Remove Aurora URL from the _links
	normalizedBody = strings.Replace(normalizedBody, domain, "", -1)

	for _, reg := range removeRegexps {
		normalizedBody = reg.ReplaceAllString(normalizedBody, "")
	}

	response.NormalizedBody = normalizedBody
	return response
}

func (r *Response) Equal(other *Response) bool {
	return r.NormalizedBody == other.NormalizedBody
}

func (r *Response) Size() int {
	return len(r.Body)
}

func (r *Response) SaveDiff(outputDir string, other *Response) {
	if r.Path != other.Path {
		panic("Paths are different")
	}

	fileName := pathToFileName(r.Path, r.Stream)

	if len(fileName) > fileLengthLimit {
		fileName = fileName[0:fileLengthLimit]
	}

	fileA := fmt.Sprintf("%s/%s.old", outputDir, fileName)
	fileB := fmt.Sprintf("%s/%s.new", outputDir, fileName)
	fileDiff := fmt.Sprintf("%s/%s.diff", outputDir, fileName)

	err := ioutil.WriteFile(fileA, []byte(r.Path+"\n\n"+r.Body), 0744)
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(fileB, []byte(other.Path+"\n\n"+other.Body), 0744)
	if err != nil {
		panic(err)
	}

	// Ignore `err`, user will generate diff manually
	out, _ := exec.Command("diff", fileA, fileB).Output()

	if len(out) != 0 {
		err = ioutil.WriteFile(fileDiff, out, 0744)
		if err != nil {
			panic(err)
		}
	}
}

// GetPaths finds all URLs in the response body and returns paths
// (without domain).
func (r *Response) GetPaths() []string {
	// escapedDomain := strings.Replace(r.Domain, `\`, `\\`, -1)
	var linksRegexp = regexp.MustCompile(`"` + r.Domain + `(.*?)["{]`)
	found := linksRegexp.FindAllSubmatch([]byte(r.Body), -1)
	links := make([]string, 0, len(found))

	for _, link := range found {
		l := strings.Replace(string(link[1]), "\\u0026", "&", -1)
		links = append(links, l)
	}

	return links
}

func pathToFileName(path string, stream bool) string {
	if stream {
		path = "stream_" + path
	}
	path = strings.Replace(path, "/", "_", -1)
	path = strings.Replace(path, "?", "_", -1)
	path = strings.Replace(path, "&", "_", -1)
	path = strings.Replace(path, "=", "_", -1)
	return path
}
